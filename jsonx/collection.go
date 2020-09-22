package jsonx

import "github.com/savsgio/gotils"

type Collection struct {
	raw _JsonCollection
}

//revive:disable:cyclomatic
func ParseCollection(v interface{}) (*Collection, error) {
	var data []byte
	switch rv := v.(type) {
	case string:
		data = gotils.S2B(rv)
	case *string:
		if rv == nil {
			data = nil
		} else {
			data = gotils.S2B(*rv)
		}
	case []byte:
		data = rv
	case *[]byte:
		if rv == nil {
			data = nil
		} else {
			data = *rv
		}
	default:
		return nil, ErrJsonValue
	}

	var dist interface{}
	err := Unmarshal(data, &dist)
	if err != nil {
		return nil, err
	}

	c := &Collection{}

	switch rv := dist.(type) {
	case map[string]interface{}:
		c.raw = Object(rv)
	case []interface{}:
		c.raw = Array(rv)
	default:
		return nil, ErrJsonValue
	}
	return c, nil
}

func MustParse(v interface{}) *Collection {
	c, e := ParseCollection(v)
	if e != nil {
		panic(e)
	}
	return c
}

func Stringify(c *Collection) ([]byte, error) { return Marshal(c.raw) }

func MustStringify(c *Collection) []byte {
	v, e := Marshal(c.raw)
	if e != nil {
		panic(e)
	}
	return v
}

func (c *Collection) Len() int {
	switch rv := c.raw.(type) {
	case Object:
		return len(rv)
	case Array:
		return len(rv)
	}
	return 0
}

func (c *Collection) Raw() _JsonCollection { return c.raw }

func (c *Collection) Get(key string) (interface{}, error) {
	return get(c.raw, key)
}

func (c *Collection) MustGet(key string) interface{} {
	v, e := c.Get(key)
	if e != nil {
		panic(e)
	}
	return v
}

func (c *Collection) GetInt(key string) (int64, error) {
	return getInt64(c.raw, key)
}

func (c *Collection) MustGetInt(key string) int64 {
	v, e := getInt64(c.raw, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (c *Collection) GetFloat(key string) (float64, error) {
	return getFloat(c.raw, key)
}

func (c *Collection) MustGetFloat(key string) float64 {
	v, e := getFloat(c.raw, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (c *Collection) GetBool(key string) (bool, error) {
	return getBool(c.raw, key)
}

func (c *Collection) MustGetBool(key string) bool {
	v, e := getBool(c.raw, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (c *Collection) GetString(key string) (string, error) {
	return getString(c.raw, key)
}

func (c *Collection) MustGetString(key string) string {
	v, e := getString(c.raw, key)
	if e != nil {
		panic(e)
	}
	return v
}

func (c *Collection) IsNull(key string) (bool, error) {
	return isNull(c.raw, key)
}

func (c *Collection) MustIsNull(key string) bool {
	v, e := isNull(c.raw, key)
	if e != nil {
		panic(e)
	}
	return v
}
