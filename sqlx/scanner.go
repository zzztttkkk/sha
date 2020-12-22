package sqlx

import "github.com/jmoiron/sqlx"

type Scanner interface {
	Scan(rows *sqlx.Rows) error
}

type ColumnScanner struct {
	temp  []interface{}
	after func()
}

// !!!! should not execute any sql in `after`
//
// row.Scan(dist...)
func NewColumnsScanner(dist []interface{}, after func()) *ColumnScanner {
	return &ColumnScanner{temp: dist, after: after}
}

func (s *ColumnScanner) Scan(rows *sqlx.Rows) error {
	for rows.Next() {
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
	after   func()
}

// !!!! should not execute any sql in `after`
//
// row.ScanStruct(*tempPtr)
func NewStructScanner(tempPtr *interface{}, after func()) *StructScanner {
	return &StructScanner{tempPtr: tempPtr, after: after}
}

func (s *StructScanner) Scan(rows *sqlx.Rows) error {
	for rows.Next() {
		if err := rows.StructScan(*s.tempPtr); err != nil {
			return err
		}
		if s.after != nil {
			s.after()
		}
	}
	return nil
}
