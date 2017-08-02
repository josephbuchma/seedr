package sql

import (
	"bytes"
	"fmt"
	"strings"
)

type SetVarer interface {
	SetVar(name string, val interface{}) string
}

type Placeholderer interface {
	Placeholders(n int) string
}

type SQLBuilder struct {
	bytes.Buffer
	SetVarer
	Placeholderer
}

func (s *SQLBuilder) sem() *SQLBuilder {
	if s.Len() > 0 {
		s.WriteString(";")
	}
	return s
}

func (s *SQLBuilder) placeholders(n int) {
	s.WriteString("(")
	s.WriteString(s.Placeholderer.Placeholders(n))
	s.WriteString(")")
}

func (s *SQLBuilder) Insert(table string, fields []string, n int) *SQLBuilder {
	iflds := strings.Join(fields, ", ")
	s.sem()
	s.WriteString(fmt.Sprintf("\nINSERT INTO %s (%s) VALUES ", table, iflds))
	if n > 1 {
		s.WriteString("\n")
	}
	for i := 0; i < n-1; i++ {
		s.placeholders(len(fields))
		s.WriteString(",\n")
	}
	s.placeholders(len(fields))
	return s
}

func (s *SQLBuilder) SetVar(varName string, val interface{}) *SQLBuilder {
	s.sem()
	s.WriteString(s.SetVarer.SetVar(varName, val))
	return s
}

func (s *SQLBuilder) Select(cols []string) *SQLBuilder {
	s.sem()
	s.WriteString(fmt.Sprintf("\nSELECT %s ", strings.Join(cols, ", ")))
	return s
}

func (s *SQLBuilder) From(table string) *SQLBuilder {
	s.WriteString(fmt.Sprintf("FROM %s", table))
	return s
}

func (s *SQLBuilder) Where(col string) *SQLBuilder {
	s.WriteString(" WHERE ")
	s.WriteString(col)
	return s
}

func (s *SQLBuilder) Eql(v interface{}) *SQLBuilder {
	s.WriteString(fmt.Sprintf("=%v", v))
	return s
}

func (s *SQLBuilder) Between(a interface{}) *SQLBuilder {
	s.WriteString(fmt.Sprintf(" BETWEEN %v", a))
	return s
}

func (s *SQLBuilder) And(a interface{}) *SQLBuilder {
	s.WriteString(fmt.Sprintf(" AND %v", a))
	return s
}

func (s *SQLBuilder) String() string {
	str := s.Buffer.String()
	return str
}
