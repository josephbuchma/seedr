package mysql

import (
	"bytes"
	"fmt"

	"github.com/josephbuchma/seedr/driver/sql"
)

type mySqlBuilder struct{}

func (_ mySqlBuilder) SetVar(name string, value interface{}) string {
	return fmt.Sprintf("\nSET @%s = %v", name, value)
}

func (_ mySqlBuilder) Placeholders(n int) string {
	s := bytes.Buffer{}
	s.WriteString("?")
	for i := 0; i < n-1; i++ {
		s.WriteString(",?")
	}
	return s.String()
}

func bsql() *sql.SQLBuilder {
	return &sql.SQLBuilder{SetVarer: mySqlBuilder{}, Placeholderer: mySqlBuilder{}}
}

const (
	lastInsertID = "LAST_INSERT_ID()"
)

func insertSQL(table string, insertFields []string) string {
	return bsql().Insert(table, insertFields, 1).String()
}

func insertBatchSQL(n int, table string, insertFields []string) string {
	return bsql().Insert(table, insertFields, n).String()
}

func selectLastSQL(table, pk string, selectFields []string) string {
	return bsql().Select(selectFields).From(table).Where(pk).Eql(lastInsertID).String()
}

func selectBetweenSQL(table, pk string, fields []string, a, b interface{}) string {
	return bsql().Select(fields).From(table).Where(pk).Between(a).And(b).String()
}
