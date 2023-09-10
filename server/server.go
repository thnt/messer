package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	mysqldriver "github.com/go-sql-driver/mysql"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Server interface
type Server interface {
	Start() error
	Stop() error
	CreateUser(username, password, name string) error
}

type session struct {
	uid uint
	ttl int
	ts  int64
}
type server struct {
	db      *gorm.DB
	mqtt    mqtt.Client
	wwwRoot http.FileSystem

	sessions     map[string]session
	mutex        sync.Mutex
	notiRegistry map[string]chan<- bool

	total       int64
	totalLastId uint
}

const datetimeFormat = "2006-01-02T15:04:05Z0700"

// New returns a new server
func New(wwwRoot http.FileSystem) (Server, error) {
	dsn := fmt.Sprintf(
		"%v:%v@tcp(%v)/%v?charset=utf8mb4&parseTime=true",
		conf.Database.Username, conf.Database.Password, conf.Database.Addr, conf.Database.DBName,
	)
	log.Printf("connecting to database...")
	level := logger.Silent
	if conf.Env == "development" {
		level = logger.Info
	}
	db, err := gorm.Open(mysql.Open(dsn)), &gorm.Config{
		Logger: logger.Default.LogMode(level),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	s := &server{
		db:      db,
		wwwRoot: wwwRoot,
	}
	if err := s.migrate(); err != nil {
		return s, fmt.Errorf("failed to migrate database: %w", err)
	}

	opts := mqtt.NewClientOptions().
		AddBroker("tcp://" + conf.MQTT.Addr).
		// SetClientID(clientID).
		SetCleanSession(true).
		SetAutoReconnect(true).
		SetConnectRetryInterval(10 * time.Second).
		SetUsername(conf.MQTT.Username).
		SetPassword(conf.MQTT.Password).
		SetOnConnectHandler(s.onMQTTConnected)
	s.mqtt = mqtt.NewClient(opts)

	return s, nil
}

func (s *server) migrate() error {
	if err := s.db.AutoMigrate(&Metric{}); err != nil {
		return fmt.Errorf("migrate Metric: %w", err)
	}
	if err := s.db.AutoMigrate(&User{}); err != nil {
		return fmt.Errorf("migrate User: %w", err)
	}

	return nil
}

func (s *server) onMQTTConnected(c mqtt.Client) {
	log.Println("connected to mqtt server")
	c.Subscribe(conf.MQTT.Topic, 1, s.onMQTTMessage)
}

func (s *server) onMQTTMessage(c mqtt.Client, msg mqtt.Message) {
	var payload struct {
		D []struct {
			Tag   string
			Value float64
		}
		Ts string
	}
	if err := json.Unmarshal(msg.Payload(), &payload); err != nil {
		log.Printf("failed to decode payload: %v: %v", string(msg.Payload()), err)
		return
	}

	var ts int64
	if d, err := time.Parse(datetimeFormat, payload.Ts); err == nil {
		ts = d.Unix()
	}
	for _, d := range payload.D {
		if strings.HasSuffix(d.Tag, ":Ts") && ts == 0 {
			ts = int64(d.Value)
		}
	}
	if ts <= 0 {
		log.Printf("failed to get timestamp from payload")
		return
	}

	var metrics []*Metric
	for _, d := range payload.D {
		if strings.HasSuffix(d.Tag, ":Ts") {
			continue
		}

		m := &Metric{
			Src:       msg.Topic(),
			Value:     d.Value,
			Timestamp: ts,
			Name:      d.Tag,
		}
		if conf.MQTT.MetricSrc != "" {
			m.Src = conf.MQTT.MetricSrc
		}

		metrics = append(metrics, m)
	}

	if len(metrics) > 0 {
		if err := s.saveMetrics(metrics...); err != nil {
			log.Printf("failed to save metrics: %v", err)
		}
	}

	go s.notify()
}

func (s *server) notify() {
	for _, c := range s.notiRegistry {
		go func(c chan<- bool) {
			c <- true
		}(c)
	}
}

func (s *server) Start() error {
	token := s.mqtt.Connect()
	if err := token.Error(); err != nil || !token.WaitTimeout(30*time.Second) {
		if err == nil {
			err = errors.New("timeout")
		}
		return fmt.Errorf("failed to connect to MQTT server: %v", err)
	}
	log.Printf("connected to MQTT server")

	go func() {
		for {
			time.Sleep(10 * time.Second)
			s.purgeSessions()
		}
	}()
	log.Printf("starting http server: %v", conf.HTTPAddr)
	return http.ListenAndServe(conf.HTTPAddr, http.HandlerFunc(s.serveHTTP))
}

func (s *server) Stop() error {
	token := s.mqtt.Unsubscribe(conf.MQTT.Topic)
	if err := token.Error(); err != nil || !token.WaitTimeout(10*time.Second) {
		if err == nil {
			err = errors.New("timeout")
		}
		return fmt.Errorf("failed to unsubscribe MQTT server: %v", err)
	}

	s.mqtt.Disconnect(10)
	log.Println("disconnected from MQTT server")

	return nil
}

func (s *server) saveMetrics(m ...*Metric) error {
	return s.db.Create(m).Error
}

func (s *server) serveHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		// handle panic
		if rcv := recover(); rcv != nil {
			s.responseError(w, http.StatusInternalServerError, "500: Internal Server Error")
		}
	}()
	if strings.HasPrefix(r.URL.Path, "/api/") {
		var res interface{}
		var err error
		defer func(t time.Time) {
			log.Printf("API: %v %v: ms=%v error=%v", r.Method, r.URL.Path, time.Since(t).Milliseconds(), err)
		}(time.Now())
		switch strings.TrimSuffix(r.URL.Path, "/") {
		case "/api/metrics":
			res, err = s.apiGetMetrics(r)
		case "/api/login":
			res, err = s.apiLogin(r)
		default:
			s.responseError(w, http.StatusNotFound, "404 Not Found")
			return
		}

		if err != nil {
			code := http.StatusBadRequest
			if errors.Is(err, errUnauthorized) {
				code = http.StatusUnauthorized
			}
			s.responseJSON(w, map[string]string{"error": err.Error()}, code)
			return
		}
		s.responseJSON(w, res)
		return
	}

	path := r.URL.Path
	r.URL.Path = "/dist" + path
	staticHandler := http.FileServer(s.wwwRoot)

	if path != "/" {
		if f, err := s.wwwRoot.Open(strings.TrimSuffix(r.URL.Path, "/")); err == nil {
			if inf, err := f.Stat(); err == nil {
				if inf.IsDir() {
					s.responseError(w, http.StatusForbidden, "403 Forbidden")
					return
				}
			}
		}
	}

	staticHandler.ServeHTTP(w, r)
}

func (s *server) getTotal(src string) (int64, error) {
	query := s.db.Model(&Metric{})
	var total int64

	if src != "" {
		query = query.Where("src = ?", src)
	}

	var lastMetric Metric
	var conds []any
	if src != "" {
		conds = append(conds, "src = ?", src)
	}
	if err := s.db.Order("id desc").First(&lastMetric, conds...).Error; err != nil {
		return 0, fmt.Errorf("get last metric: %v", err)
	}

	if s.totalLastId > 0 {
		query = query.Where("id > ?", s.totalLastId)
	}

	if err := query.Group("timestamp").Count(&total).Error; err != nil {
		return 0, fmt.Errorf("count: %w", err)
	}

	if lastMetric.ID > 0 {
		s.total += total
		s.totalLastId = lastMetric.ID
	}

	return s.total, nil
}

func (s *server) apiGetMetrics(r *http.Request) (interface{}, error) {
	_, err := s.currentUser(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errUnauthorized, err)
	}

	q := r.URL.Query()

	var src string
	if v := q.Get("src"); v != "" {
		src = v
	}
	var from, to int64
	if v := q.Get("from"); v != "" {
		if vv, err := strconv.Atoi(v); err == nil && vv > 0 {
			from = int64(vv)
		}
	}
	if v := q.Get("to"); v != "" {
		if vv, err := strconv.Atoi(v); err == nil && vv > 0 {
			to = int64(vv)
		}
	}

	var limit, skip int
	if v := q.Get("limit"); v != "" {
		if vv, err := strconv.Atoi(v); err == nil && vv > 0 {
			limit = vv
		}
	}
	if v := q.Get("skip"); v != "" {
		if vv, err := strconv.Atoi(v); err == nil && vv > 0 {
			skip = vv
		}
	}
	var watch int
	if v := q.Get("watch"); v != "" {
		if vv, err := strconv.Atoi(v); err == nil && vv > 0 {
			watch = vv
		}
	}
	if watch > 60 {
		watch = 60
	}

	if limit == 0 {
		limit = 10
	}

	query := s.db.Model(&Metric{})

	var wheres []string
	var args []any
	if src != "" {
		query = query.Where("src = ?", src)
		wheres = append(wheres, "src = ?")
		args = append(args, src)
	}
	if from > 0 {
		query = query.Where("timestamp >= ?", from)
		wheres = append(wheres, fmt.Sprintf("timestamp >= %v", from))
	} else if skip == 0 {
		wheres = append(wheres, fmt.Sprintf("timestamp >= %v", time.Now().AddDate(0, 0, -30).Unix()))
	}
	if to > 0 {
		query = query.Where("timestamp <= ?", to)
		wheres = append(wheres, fmt.Sprintf("timestamp <= %v", to))
	}
	var where string
	if len(wheres) > 0 {
		where = "WHERE " + strings.Join(wheres, " AND ")
	}

	var total int64
	if watch == 0 {
		if from+to == 0 {
			total, err = s.getTotal(src)
			if err != nil {
				return nil, err
			}
		} else if err := query.Group("timestamp").Count(&total).Error; err != nil {
			return nil, fmt.Errorf("count: %w", err)
		}
	}

	results := []map[string]any{}

	var ch chan bool
	if watch > 0 {
		ch = make(chan bool, 3)
		time.AfterFunc(time.Duration(watch)*time.Second, func() {
			ch <- false
		})
		id := s.registerNoti(ch)
		defer s.unregisterNoti(id)
	}

	res := map[string]interface{}{
		"total":   total,
		"metrics": &results,
	}

	const MAX_METRICS_PER_TIMESTAMP = 20

	var names []string
	{
		q := s.db.Model(&Metric{}).Distinct("name")
		if src != "" {
			q = q.Where("src = ?", src)
		}
		if err := q.Scan(&names).Error; err != nil {
			return nil, fmt.Errorf("get metrics name: %v", err)
		}
	}
	if len(names) == 0 {
		return res, nil
	}

	var mcols []string
	for _, n := range names {
		mcols = append(mcols, fmt.Sprintf(`MAX(CASE WHEN name = '%v' THEN value END) '%v'`, n, n))
	}

	for {
		sql := fmt.Sprintf(`
		SELECT timestamp as Timestamp, %v
		FROM (
			SELECT MAX(id) id FROM metrics %v GROUP BY name, timestamp ORDER BY id DESC LIMIT %v
		) t LEFT JOIN metrics ON t.id = metrics.id
		GROUP BY timestamp
		ORDER BY timestamp DESC
		LIMIT %v, %v
	`, strings.Join(mcols, ","), where, (limit+skip)*MAX_METRICS_PER_TIMESTAMP, skip, limit)
		query = s.db.Raw(sql, args...)
		if err := query.Scan(&results).Error; err != nil {
			return nil, fmt.Errorf("query: %w", err)
		}
		if len(results) > 0 || ch == nil {
			break
		}
		if !<-ch {
			res["timeout"] = true
			break
		}
	}

	return res, nil
}

func (s *server) responseError(w http.ResponseWriter, code int, content string) {
	w.WriteHeader(code)
	w.Write([]byte(content))
}

func (s *server) responseJSON(w http.ResponseWriter, data interface{}, opts ...interface{}) {
	status := http.StatusOK
	if len(opts) > 0 {
		if s, ok := opts[0].(int); ok && s > 0 {
			status = s
		}
	}

	b, _ := json.Marshal(data)
	w.Header().Set("Content-Type", "application/json")
	if d, ok := data.(*loginResponse); ok {
		ck := http.Cookie{
			Name:     conf.Cookie.Name,
			Value:    d.ssid,
			MaxAge:   86400,
			HttpOnly: true,
			Path:     "/api",
		}
		w.Header().Set("Set-Cookie", ck.String())
	} else if d, ok := data.(*logoutResponse); ok {
		ck := http.Cookie{
			Name:     conf.Cookie.Name,
			Value:    d.ssid,
			MaxAge:   -1,
			HttpOnly: true,
			Path:     "/api",
		}
		w.Header().Set("Set-Cookie", ck.String())
	}
	w.WriteHeader(status)
	w.Write(b)
}

func (s *server) CreateUser(username, password, name string) error {
	if m, err := regexp.MatchString("^[[:alnum:]]{4,20}$", username); err != nil || !m {
		return errors.New("invalid username")
	}
	if len(password) < 4 {
		return errors.New("invalid password")
	}
	if len(name) > 50 {
		return errors.New("invalid name")
	}
	if name == "" {
		name = username
	}

	hashedPwd, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err := s.db.Create(&User{
		Username: username,
		Name:     name,
		Password: string(hashedPwd),
	}).Error; err != nil {
		return err
	}

	return nil
}

var errUnauthorized = errors.New("unauthorized")

func (s *server) apiAuthorize(r *http.Request) (interface{}, error) {
	u, err := s.currentUser(r)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", errUnauthorized, err)
	}

	return u, nil
}

type loginResponse struct {
	*User
	ssid string
}

func (s *server) apiLogin(r *http.Request) (interface{}, error) {
	if r.Method == http.MethodGet {
		return s.apiAuthorize(r)
	}
	if r.Method == http.MethodDelete {
		return s.apiLogout(r)
	}

	var req struct {
		Username string
		Password string
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return nil, fmt.Errorf("decode request: %w", err)
	}

	if req.Username == "" || req.Password == "" {
		return nil, errors.New("missing username or password")
	}

	var u User
	if err := s.db.Find(&u, "username = ?", req.Username).Error; err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}

	if u.ID <= 0 || bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(req.Password)) != nil {
		return nil, errors.New("wrong username or password")
	}

	sid := s.addSession(u.ID, 86400)

	return &loginResponse{
		User: &u,
		ssid: sid,
	}, nil
}

type logoutResponse struct {
	ssid string
}

func (s *server) apiLogout(r *http.Request) (interface{}, error) {
	if c, err := r.Cookie(conf.Cookie.Name); err == nil && c.Value != "" {
		return &logoutResponse{
			ssid: c.Value,
		}, nil
	}

	return nil, nil
}

func (s *server) currentUser(r *http.Request) (*User, error) {
	var sid string
	if c, err := r.Cookie(conf.Cookie.Name); err == nil {
		sid = c.Value
	}
	if sid == "" {
		return nil, errors.New("missing ssid")
	}

	ss := s.getSession(sid)
	if ss == nil || ss.isExpired() {
		return nil, errors.New("session not found")
	}

	var u User
	if err := s.db.First(&u, ss.uid).Error; err != nil || u.ID == 0 {
		return nil, errors.New("user not found")
	}

	return &u, nil
}

func (s *server) addSession(uid uint, ttl int) string {
	b := make([]byte, 8)
	rand.Read(b)
	now := time.Now()
	sid := fmt.Sprintf("%x%x", b, now.Unix())

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.sessions == nil {
		s.sessions = map[string]session{}
	}
	s.sessions[sid] = session{
		uid: uid,
		ttl: ttl,
		ts:  now.Unix(),
	}

	return sid
}

func (s *server) getSession(id string) *session {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if ss, ok := s.sessions[id]; ok {
		return &ss
	}

	return nil
}

func (s *server) purgeSessions() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	for k, v := range s.sessions {
		if v.ts+int64(v.ttl) < time.Now().Unix() {
			delete(s.sessions, k)
		}
	}
}

func (s *session) isExpired() bool {
	return s.ts+int64(s.ttl)+10 < time.Now().Unix()
}

func (s *server) registerNoti(ch chan<- bool) string {
	b := make([]byte, 4)
	rand.Read(b)
	now := time.Now()
	id := fmt.Sprintf("%x%x", b, now.Unix())

	s.mutex.Lock()
	defer s.mutex.Unlock()
	if s.notiRegistry == nil {
		s.notiRegistry = make(map[string]chan<- bool)
	}
	if _, ok := s.notiRegistry[id]; ok {
		return s.registerNoti(ch)
	}
	s.notiRegistry[id] = ch

	return id
}

func (s *server) unregisterNoti(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.notiRegistry, id)
}
