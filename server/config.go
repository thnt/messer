package server

import (
	"errors"
	"os"

	"github.com/JeremyLoy/config"
)

type cfg struct {
	ConfigFile string
	Env        string `config:"APP_ENV"`
	HTTPAddr   string `config:"HTTP_ADDR"`

	Cookie struct {
		Name string
	}

	Database struct {
		Addr     string
		Username string
		Password string
		DBName   string
	} `config:"DB"`

	MQTT struct {
		Addr      string
		Username  string
		Password  string
		Topic     string
		MetricSrc string `config:"METRIC_SRC"`
	}
}

var conf = cfg{
	HTTPAddr: "127.0.0.1:9000",
}

func init() {
	conf.Cookie.Name = "ssid"

	configFile := ".env"
	if f := os.Getenv("CONFIG_FILE"); f != "" {
		configFile = f
	}
	if _, err := os.Stat(configFile); !os.IsNotExist(err) {
		conf.ConfigFile = configFile
		config.From(configFile).FromEnv().To(&conf)
	} else {
		config.FromEnv().To(&conf)
	}
}

func (c cfg) Validate() error {
	if c.Database.Addr == "" || c.Database.DBName == "" || c.Database.Username == "" || c.Database.Password == "" {
		return errors.New("missing database config")
	}

	if c.MQTT.Addr == "" || c.MQTT.Topic == "" {
		return errors.New("missing MQTT config")
	}

	return nil
}

func Config() cfg {
	return conf
}
