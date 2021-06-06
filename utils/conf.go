package utils

import (
	"fmt"
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

func (d *TomlDuration) UnmarshalJSON(text []byte) error { return d.UnmarshalText(text) }

func confFromTomlBytes(conf interface{}, data []byte) error {
	_, err := toml.Decode(string(data), conf)
	return err
}

var confEnvReg = regexp.MustCompile(`\$ENV{.*?}`)

func confDoReplace(fp, name string, rawValue string, path []string) string {
	key := strings.Join(path, ".") + "." + name

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
		return string(data)
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
	return S(s)
}

func confReflectMap(filePath string, value reflect.Value, path []string) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
		if value.Kind() == reflect.Ptr {
			return
		}
	}

	if value.Kind() == reflect.Interface {
		value = reflect.ValueOf(value.Interface())
	}

	vt := value.Type()
	switch vt.Kind() {
	case reflect.Struct:
		for i := 0; i < value.NumField(); i++ {
			ft := value.Type().Field(i)
			if ft.Tag.Get("toml") == "-" {
				continue
			}
			fv := value.Field(i)
			rawV := fv
			if fv.Kind() == reflect.Interface {
				fv = reflect.ValueOf(fv.Interface())
			}

			if fv.Kind() != reflect.String {
				np := make([]string, 0, len(path))
				np = append(np, path...)
				np = append(np, ft.Name)
				confReflectMap(filePath, fv, np)
				continue
			}

			ns := confDoReplace(filePath, ft.Name, fv.String(), path)
			rawV.Set(reflect.ValueOf(ns))
		}
	case reflect.Map:
		iter := value.MapRange()
		for iter.Next() {
			v := iter.Value()
			if v.Kind() == reflect.Interface {
				v = reflect.ValueOf(v.Interface())
			}

			if v.Kind() != reflect.String {
				np := make([]string, 0, len(path))
				np = append(np, path...)
				np = append(np, iter.Key().String())

				confReflectMap(filePath, v, np)
				continue
			}
			ns := confDoReplace(filePath, iter.Key().String(), v.String(), path)
			value.SetMapIndex(iter.Key(), reflect.ValueOf(ns))
		}
	case reflect.Slice:
		for i := 0; i < value.Len(); i++ {
			v := value.Index(i)
			rawV := v
			if v.Kind() == reflect.Interface {
				v = reflect.ValueOf(v.Interface())
			}
			if v.Kind() != reflect.String {
				np := make([]string, 0, len(path))
				np = append(np, path...)
				np = append(np, fmt.Sprintf("$%d", i))
				confReflectMap(filePath, v, np)
				continue
			}
			ns := confDoReplace(filePath, fmt.Sprintf("$%d", i), v.String(), path)
			rawV.Set(reflect.ValueOf(ns))
		}
	}
}

func LoadToml(conf interface{}, fp string, logFileContent bool) error {
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
	if logFileContent {
		log.Printf("sha.utils.config: load from file `%s`\n%s\n", fp, v)
	}
	return nil
}
