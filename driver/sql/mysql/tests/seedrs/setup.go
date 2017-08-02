package seedrs

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"

	"github.com/josephbuchma/seedr"
	"github.com/josephbuchma/seedr/driver/sql/mysql"
)

func openTestDB() *sql.DB {
	db, err := sql.Open("mysql", "root:@/seedr_test?parseTime=true")
	if err != nil {
		panic(err)
	}
	return db
}

var TestSeedr = seedr.New("test_seedr",
	seedr.SetCreateDriver(mysql.New(openTestDB())),
	seedr.SetFieldMapper(
		seedr.RegexpTagFieldMapper(
			`.*gorm:"column:\s*(\w+).*"`, seedr.SnakeFieldMapper(),
		),
	),
)
