package sqls

import (
	"fmt"
	"strings"
)

type Dict map[string]interface{}

func (dict Dict) ToMap() map[string]interface{} {
	return dict
}

func (dict Dict) Keys() (lst []string) {
	for k, _ := range dict {
		lst = append(lst, k)
	}
	return
}

func (dict Dict) Values() (lst []interface{}) {
	for _, v := range dict {
		lst = append(lst, v)
	}
	return
}

func (dict Dict) Expand(formatter func(string, int) string) (keys []string, vals []interface{}) {
	var ind = 0
	for k, v := range dict {
		keys = append(keys, formatter(k, ind))
		vals = append(vals, v)
		ind++
	}
	return
}

func (dict Dict) ForUpdate() (string, []interface{}) {
	keys, values := dict.Expand(func(s string, i int) string { return s + "=?" })
	return strings.Join(keys, ","), values
}

func (dict Dict) ForCreate(tableName string) (string, []interface{}) {
	placeholder := ""
	last := len(dict) - 1
	keys, values := dict.Expand(
		func(s string, i int) string {
			placeholder += "?"
			if i < last {
				placeholder += ","
			}
			return s
		},
	)

	return fmt.Sprintf(
		"insert into %s (%s) values(%s)",
		tableName,
		strings.Join(keys, ","),
		placeholder,
	), values
}
