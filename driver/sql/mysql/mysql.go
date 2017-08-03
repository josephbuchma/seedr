// Package mysql is a MySQL driver for Seedr.
// Entity (which represents table name in this case) and PrimaryKey must be specified for each Factory.
package mysql

import (
	"database/sql"
	"errors"

	"github.com/josephbuchma/seedr/driver"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

// MySQL driver for Seedr
type MySQL struct {
	db *sql.DB
}

// New creates new Driver.
func New(db *sql.DB) driver.Driver {
	return &MySQL{db}
}

// Create inserts payload Data into database and returns inserted records
// Entity is a table name. If PrimaryKey is not provided, no results will be returned.
func (my *MySQL) Create(p driver.Payload) (results []map[string]interface{}, err error) {
	tx, err := my.db.Begin()
	if err != nil {
		return nil, err
	}
	d := drv{tx}
	ret, err := d.insert(insertPayload{p.Entity, p.PrimaryKey, p.InsertFields, p.ReturnFields, p.Data})
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	err = tx.Commit()
	return ret, err
}

type drv struct {
	db *sql.Tx
}

type insertPayload struct {
	table, pk                  string
	insertFields, returnFields []string
	data                       []map[string]interface{}
}

func (ip insertPayload) AllValues() []interface{} {
	sls := make([]interface{}, len(ip.insertFields)*len(ip.data))
	i := 0
	for _, d := range ip.data {
		for _, f := range ip.insertFields {
			sls[i] = d[f]
			i++
		}
	}
	return sls
}

func makePtrs(v []interface{}) []interface{} {
	ptrs := make([]interface{}, len(v))
	for i := range v {
		ptrs[i] = &v[i]
	}
	return ptrs
}

func (my drv) queryRowScan(sql string, vals []interface{}, result []interface{}) {
	panicOnError(my.db.QueryRow(sql, vals...).Scan(makePtrs(result)...))
}

func (my drv) exec(sql string, vals []interface{}) error {
	_, err := my.db.Exec(sql, vals...)
	return err
}

func (my drv) query(sql string, vals []interface{}, results []interface{}) error {
	rows, err := my.db.Query(sql, vals...)
	if err != nil {
		return err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return err
	}

	shift := 0
	for rows.Next() {
		r := results[shift : shift+len(cols)]
		err := rows.Scan(makePtrs(r)...)
		if err != nil {
			return err
		}
		shift += len(cols)
	}
	return rows.Err()
}

func (my *drv) insert(ins insertPayload) ([]map[string]interface{}, error) {
	inserted, err := my.insertRows(len(ins.data), ins.table, ins.pk, ins.insertFields, ins.returnFields, ins.AllValues())
	if err != nil {
		return nil, err
	}
	ret := make([]map[string]interface{}, 0, len(ins.data))
	for i := 0; i < len(inserted); {
		r := make(map[string]interface{})
		for _, f := range ins.returnFields {
			r[f] = inserted[i]
			i++
		}
		ret = append(ret, r)
	}

	return ret, nil
}

func (my *drv) insertRows(n int, table, pk string, insertFields, selectFields []string, vals []interface{}) ([]interface{}, error) {
	var firstRecordID int64
	var err error
	if len(vals) == 0 || n == 0 {
		return nil, errors.New("Nothing to create")
	}
	if len(vals)/n != len(insertFields) {
		// must be unreachable
		panic("INVALID LENGTH OF VALS")
	}
	ret := make([]interface{}, n*len(selectFields))

	// Insert and fetch first record
	s := insertSQL(table, insertFields)
	if err := my.exec(s, vals[0:len(insertFields)]); err != nil {
		return nil, err
	}
	if pk == "" {
		return nil, nil
	}
	if n == 1 {
		s = selectLastSQL(table, pk, selectFields)
		err := my.query(s, nil, ret[0:len(selectFields)])
		return ret, err
	}
	s = selectLastSQL(table, pk, []string{pk})
	err = my.db.QueryRow(s).Scan(&firstRecordID)
	if err != nil {
		return nil, err
	}

	s = insertBatchSQL(n-1, table, insertFields)
	err = my.exec(s, vals[len(insertFields):])
	if err != nil {
		return nil, err
	}

	s = selectBetweenSQL(table, pk, selectFields, firstRecordID, int(firstRecordID)+len(vals))
	err = my.query(s, nil, ret)
	return ret, err
}

func indexOfStr(str string, s []string) int {
	for i, v := range s {
		if v == str {
			return i
		}
	}
	panic("failed to find string in slice")
}
