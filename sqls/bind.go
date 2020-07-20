package sqls

func BindNamed(q string, dict map[string]interface{}) (string, []interface{}) {
	s, vl, err := Leader().BindNamed(q, dict)
	if err != nil {
		panic(err)
	}
	return s, vl
}
