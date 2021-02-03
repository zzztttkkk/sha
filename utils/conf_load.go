package utils

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
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

	if strings.HasSuffix(fp, ".toml") {
		e = confFromTomlBytes(conf, v)
	} else {
		e = json.Unmarshal(v, conf)
	}

	if e != nil {
		return e
	}
	confReflectMap(fp, reflect.ValueOf(conf), []string{})

	fp, err := filepath.Abs(fp)
	if err != nil {
		panic(err)
	}
	log.Printf("sha.utils.config: load from file `%s`\n", fp)
	return nil
}

func (cn _ConfNamespace) LoadFromFiles(dist interface{}, fps ...string) {
	if reflect.TypeOf(dist).Kind() != reflect.Ptr || dist == nil || reflect.ValueOf(dist).IsNil() {
		panic(fmt.Errorf("sha.utils.config: bad dist value, `%v`\n", dist))
	}

	t := dist

	ct := reflect.TypeOf(dist).Elem()
	if ct.Kind() != reflect.Struct {
		panic(fmt.Errorf("sha.utils.config: dist is not a struct pointer"))
	}

	for _, fp := range fps {
		ele := reflect.New(ct).Interface()
		if err := cn.LoadFromFile(ele, fp); err != nil {
			panic(err)
		}

		if err := mergo.Merge(t, ele, mergo.WithOverride); err != nil {
			panic(err)
		}
	}
}
