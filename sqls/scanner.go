package sqls

import "github.com/jmoiron/sqlx"

type BeforeScan func(dist *[]interface{})
type AfterScan func(dist *[]interface{}) error

type Scanner struct {
	beforeScan BeforeScan
	dist       []interface{}
	afterScan  AfterScan
	isStruct   bool
}

func (s *Scanner) Scan(rows *sqlx.Rows) int {
	c := 0
	for rows.Next() {
		if s.beforeScan != nil {
			s.beforeScan(&s.dist)
		}
		if s.isStruct {
			if err := rows.StructScan(s.dist[0]); err != nil {
				panic(err)
			}
		} else {
			if err := rows.Scan(s.dist...); err != nil {
				panic(err)
			}
		}
		if s.afterScan != nil {
			if err := s.afterScan(&s.dist); err != nil {
				panic(err)
			}
		}
		c++
	}
	return c
}

func NewScanner(dist []interface{}, before BeforeScan, after AfterScan) *Scanner {
	return &Scanner{
		beforeScan: before,
		afterScan:  after,
		dist:       dist,
	}
}

func NewStructScanner(before BeforeScan, after AfterScan) *Scanner {
	return &Scanner{
		beforeScan: before,
		afterScan:  after,
		dist:       make([]interface{}, 1, 1),
		isStruct:   true,
	}
}
