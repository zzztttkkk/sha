package toml

import (
	btoml "github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
	"github.com/savsgio/gotils"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
)

func FromBytes(conf interface{}, data []byte) error {
	_, err := btoml.Decode(string(data), conf)
	return err
}

var envReg = regexp.MustCompile(`\$ENV{\s*\w+\s*}`)

func doReplace(fp, name string, value *reflect.Value, path []string) {
	key := strings.Join(path, ".") + "." + name

	rawValue := (*value).Interface().(string)
	if strings.HasPrefix(rawValue, "file://") {
		_fp := rawValue[7:]
		var f *os.File
		var e error

		if !strings.HasPrefix(_fp, "/") {
			_fp = filepath.Join(filepath.Dir(fp), _fp)
		}

		f, e = os.Open(_fp)
		if e != nil {
			log.Fatalf("suna.utils.toml: key: `%s`; raw: `%s`; err: `%s`\n", key, rawValue, e.Error())
		}
		defer f.Close()

		data, e := ioutil.ReadAll(f)
		if e != nil {
			log.Fatalf("suna.utils.toml: key: `%s`; raw: `%s`; err: `%s`\n", key, rawValue, e.Error())
		}
		value.SetString(string(data))
		return
	}

	s := envReg.ReplaceAllFunc(
		gotils.S2B(rawValue),
		func(data []byte) []byte {
			envK := strings.TrimSpace(string(data[5 : len(data)-1]))
			return []byte(os.Getenv(envK))
		},
	)
	value.SetString(string(s))
}

func reflectMap(filePath string, value reflect.Value, path []string) {
	ele := value.Elem()
	t := ele.Type()

	for i := 0; i < ele.NumField(); i++ {
		filed := ele.Field(i)

		tf := t.Field(i)
		if tf.Tag.Get("toml") == "-" {
			continue
		}
		switch filed.Type().Kind() {
		case reflect.String:
			doReplace(filePath, tf.Name, &filed, path)
		case reflect.Struct:
			cp := path[:]
			cp = append(cp, tf.Name)
			reflectMap(filePath, filed.Addr(), cp)
		}
	}
}

func FromFile(conf interface{}, fp string) error {
	f, e := os.Open(fp)
	if e != nil {
		panic(e)
	}
	defer f.Close()

	v, e := ioutil.ReadAll(f)
	if e != nil {
		panic(e)
	}

	err := FromBytes(conf, v)
	if err != nil {
		return err
	}
	reflectMap(fp, reflect.ValueOf(conf), []string{})
	return nil
}

func FromFiles(conf interface{}, defaultV interface{}, fps ...string) {
	t := conf
	ct := reflect.TypeOf(conf).Elem()

	for _, fp := range fps {
		ele := reflect.New(ct).Interface()
		err := FromFile(ele, fp)
		if err != nil {
			panic(err)
		}
		if t == nil {
			t = ele
		} else {
			if err := mergo.Merge(t, ele, mergo.WithOverride); err != nil {
				panic(err)
			}
		}
	}

	if defaultV != nil {
		if err := mergo.Merge(t, defaultV); err != nil {
			panic(err)
		}
	}
}

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
