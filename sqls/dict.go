package sqls

import (
	"fmt"
)

type Dict map[string]interface{}

func (m Dict) Replace(k string, fn func(interface{}) interface{}) {
	v, ok := m[k]
	if ok {
		m[k] = fn(v)
	}
}

func (m Dict) Reset(k string, fn func(interface{}) interface{}) {
	m[k] = fn(m[k])
}

func (m Dict) ForUpdate() (placeholder string, values []interface{}) {
	i := 1
	l := len(m)
	for k, v := range m {
		values = append(values, v)

		if driverName == "postgres" {
			placeholder += fmt.Sprintf("%s=$%d", k, i)
		} else {
			placeholder += fmt.Sprintf("%s=?", k)
		}

		if i < l {
			placeholder += ","
		}
		i++
	}
	return
}
