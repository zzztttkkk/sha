package utils

import (
	"github.com/BurntSushi/toml"
	"github.com/zzztttkkk/sha/internal"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"time"
)

type TomlDuration struct {
	time.Duration
}

func (d *TomlDuration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = internal.ParseDuration(string(text))
	return err
}

func (d *TomlDuration) UnmarshalJSON(text []byte) error {
	return d.UnmarshalText(text)
}

func confFromTomlBytes(conf interface{}, data []byte) error {
	_, err := toml.Decode(string(data), conf)
	return err
}

var confEnvReg = regexp.MustCompile(`\$ENV{.*?}`)

func confDoReplace(fp, name string, value *reflect.Value, path []string) {
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
			log.Fatalf("sha.utils.config: file: `%s`; key: `%s`; raw: `%s`; err: `%s`\n", fp, key, rawValue, e.Error())
		}
		defer f.Close()

		data, e := ioutil.ReadAll(f)
		if e != nil {
			log.Fatalf("sha.utils.config: file: `%s`; key: `%s`; raw: `%s`; err: `%s`\n", fp, key, rawValue, e.Error())
		}
		value.SetString(string(data))
		return
	}

	s := confEnvReg.ReplaceAllFunc(
		B(rawValue),
		func(data []byte) []byte {
			envK := strings.TrimSpace(string(data[5 : len(data)-1]))
			v := os.Getenv(envK)
			if len(v) < 1 {
				log.Fatalf("sha.utils.config: file: `%s`; key: `%s`;  empty env variable `%s`\n", fp, key, envK)
			}
			return []byte(v)
		},
	)
	value.SetString(string(s))
}

func confReflectMap(filePath string, value reflect.Value, path []string) {
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
			confDoReplace(filePath, tf.Name, &filed, path)
		case reflect.Struct:
			cp := path[:]
			cp = append(cp, tf.Name)
			confReflectMap(filePath, filed.Addr(), cp)
		}
	}
}

type _ConfNamespace struct{}

var Conf _ConfNamespace

func (_ConfNamespace) LoadFromFile(conf interface{}, fp string) error {
	f, e := os.Open(fp)
	if e != nil {
		panic(e)
	}
	defer f.Close()

	v, e := ioutil.ReadAll(f)
	if e != nil {
		panic(e)
	}

	e = confFromTomlBytes(conf, v)
	if e != nil {
		return e
	}
	confReflectMap(fp, reflect.ValueOf(conf), []string{})

	fp, err := filepath.Abs(fp)
	if err != nil {
		panic(err)
	}
	log.Printf("sha.utils.config: load from file `%s`\n%s\n", fp, v)
	return nil
}
