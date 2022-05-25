package server

import "time"

type Metric struct {
	ID        uint      `json:"id"`
	Src       string    `json:"src"`
	Name      string    `json:"name" gorm:"type:varchar(255)"`
	Value     float64   `json:"value"`
	Timestamp int64     `json:"ts" gorm:"index"`
	CreatedAt time.Time `json:"-"`
}

type User struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Username  string    `json:"username"`
	Password  string    `json:"-"`
	CreatedAt time.Time `json:"-"`
	UpdatedAt time.Time `json:"-"`
}
