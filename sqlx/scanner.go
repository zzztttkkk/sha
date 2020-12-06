package sqlx

import "github.com/jmoiron/sqlx"

type Scanner interface {
	Scan(rows *sqlx.Rows) error
}

type ColumnScanner struct {
	temp   []interface{}
	before func()
	after  func()
}

// scan row to `temps []interface{}`
func NewColumnsScanner(dist []interface{}, before func(), after func()) *ColumnScanner {
	return &ColumnScanner{
		temp:   dist,
		before: before,
		after:  after,
	}
}

func (s *ColumnScanner) Scan(rows *sqlx.Rows) error {
	for rows.Next() {
		if s.before != nil {
			s.before()
		}
		if err := rows.Scan(s.temp...); err != nil {
			return err
		}
		if s.after != nil {
			s.after()
		}
	}
	return nil
}

type StructScanner struct {
	tempPtr *interface{}
	before  func()
	after   func()
}

func NewStructScanner(tempPtr *interface{}, before func(), after func()) *StructScanner {
	return &StructScanner{
		tempPtr: tempPtr,
		before:  before,
		after:   after,
	}
}

func (s *StructScanner) Scan(rows *sqlx.Rows) error {
	for rows.Next() {
		if s.before != nil {
			s.before()
		}

		if err := rows.StructScan(*s.tempPtr); err != nil {
			return err
		}
		if s.after != nil {
			s.after()
		}
	}
	return nil
}
