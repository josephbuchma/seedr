// Package test_models is an example of basic models setup.
// NOTE: this project has no any direct relation to Gorm ORM,
// tag is for example.
package models

import (
	"time"

	"github.com/go-sql-driver/mysql"
)

type User struct {
	ID        int
	Email     string
	Name      string
	Active    bool
	Checkin   mysql.NullTime
	CreatedAt time.Time
}

type Article struct {
	ID        int
	UserID    int `gorm:"column:author_id"`
	Title     string
	Body      string
	CreatedAt time.Time
}

type Club struct {
	ID   int
	Name string
}

type HellotaFields struct {
	A int
	B string
	C string
	D time.Time
	E bool
	F int
	G string
	H string
	I time.Time
	J bool
	K int
	L string
	M string
	N time.Time
	O bool
	P int
	Q string
	R string
	S time.Time
	T bool
	U int
	V string
	W string
	X time.Time
	Y bool
	Z int
}
