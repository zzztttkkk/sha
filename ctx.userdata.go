// copied from `https://raw.githubusercontent.com/valyala/fasthttp/master/userdata.go`

package sha

import (
	"io"
)

type userDataKV struct {
	key   interface{}
	value interface{}
}

type userData []userDataKV

func (d *userData) Set(key interface{}, value interface{}) {
	args := *d
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if kv.key == key {
			kv.value = value
			return
		}
	}

	c := cap(args)
	if c > n {
		args = args[:n+1]
		kv := &args[n]
		kv.key = key
		kv.value = value
		*d = args
		return
	}
	*d = append(args, userDataKV{key: key, value: value})
}

func (d *userData) Get(key interface{}) interface{} {
	args := *d
	n := len(args)
	for i := 0; i < n; i++ {
		kv := &args[i]
		if kv.key == key {
			return kv.value
		}
	}
	return nil
}

func (d *userData) Visit(fn func(k interface{}, v interface{}) bool) {
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
