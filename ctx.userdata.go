// copied from `https://raw.githubusercontent.com/valyala/fasthttp/master/userdata.go`

package sha

import (
	"github.com/zzztttkkk/sha/utils"
	"io"
)

type userDataKV struct {
	key   []byte
	value interface{}
}

type userData []userDataKV

func (d *userData) Set(key string, value interface{}) {
	args := *d
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if string(kv.key) == key {
			kv.value = value
			return
		}
	}

	c := cap(args)
	if c > n {
		args = args[:n+1]
		kv := &args[n]
		kv.key = append(kv.key[:0], key...)
		kv.value = value
		*d = args
		return
	}

	kv := userDataKV{}
	kv.key = append(kv.key[:0], key...)
	kv.value = value
	*d = append(args, kv)
}

func (d *userData) SetBytes(key []byte, value interface{}) {
	d.Set(utils.S(key), value)
}

func (d *userData) Get(key string) interface{} {
	args := *d
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if string(kv.key) == key {
			return kv.value
		}
	}
	return nil
}

func (d *userData) GetBytes(key []byte) interface{} {
	return d.Get(utils.S(key))
}

func (d *userData) Visit(fn func(k []byte, v interface{}) bool) {
	for _, item := range *d {
		if !fn(item.key, item.value) {
			break
		}
	}
}

func (d *userData) Reset() {
	args := *d
	n := len(args)
	for i := 0; i < n; i++ {
		v := args[i].value
		if vc, ok := v.(io.Closer); ok {
			_ = vc.Close()
		}
	}
	*d = (*d)[:0]
}
