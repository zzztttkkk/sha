package jsonx

import (
	"encoding/json"
	"github.com/zzztttkkk/suna/config"
	"github.com/zzztttkkk/suna/internal"
)

var _UnmarshalImpl = json.Unmarshal
var _MarshalImpl = json.Marshal

func Unmarshal(data []byte, dist interface{}) error {
	return _UnmarshalImpl(data, dist)
}

func Marshal(val interface{}) ([]byte, error) {
	return _MarshalImpl(val)
}

func init() {
	internal.Dig.LazyInvoke(
		func(cfg *config.Suna) {
			_UnmarshalImpl = cfg.Json.Unmarshal
			_MarshalImpl = cfg.Json.Marshal
		},
	)
}
