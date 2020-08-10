package utils

import (
	btoml "github.com/BurntSushi/toml"
	"github.com/imdario/mergo"
	"io/ioutil"
	"os"
	"reflect"
	"time"
)

func TomlFromBytes(conf interface{}, data []byte) error {
	_, err := btoml.Decode(string(data), conf)
	return err
}

func TomlFromFile(conf interface{}, fp string) error {
	f, e := os.Open(fp)
	if e != nil {
		panic(e)
	}
	defer f.Close()

	v, e := ioutil.ReadAll(f)
	if e != nil {
		panic(e)
	}
	return TomlFromBytes(conf, v)
}

func TomlFromFiles(conf interface{}, defaultV interface{}, fps ...string) {
	t := conf
	ct := reflect.TypeOf(conf).Elem()

	for _, fp := range fps {
		ele := reflect.New(ct).Interface()
		err := TomlFromFile(ele, fp)
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

type TomlDuration struct {
	time.Duration
}

func (d *TomlDuration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}
