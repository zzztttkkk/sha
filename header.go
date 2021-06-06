package sha

import (
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"strings"
)

type Header struct {
	utils.Kvs
	fromOutSide bool
	buf         strings.Builder
}

const toLowerTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@abcdefghijklmnopqrstuvwxyz[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"
const toUpperTable = "\x00\x01\x02\x03\x04\x05\x06\a\b\t\n\v\f\r\x0e\x0f\x10\x11\x12\x13\x14\x15\x16\x17\x18\x19\x1a\x1b\x1c\x1d\x1e\x1f !\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`ABCDEFGHIJKLMNOPQRSTUVWXYZ{|}~\u007f\x80\x81\x82\x83\x84\x85\x86\x87\x88\x89\x8a\x8b\x8c\x8d\x8e\x8f\x90\x91\x92\x93\x94\x95\x96\x97\x98\x99\x9a\x9b\x9c\x9d\x9e\x9f\xa0\xa1\xa2\xa3\xa4\xa5\xa6\xa7\xa8\xa9\xaa\xab\xac\xad\xae\xaf\xb0\xb1\xb2\xb3\xb4\xb5\xb6\xb7\xb8\xb9\xba\xbb\xbc\xbd\xbe\xbf\xc0\xc1\xc2\xc3\xc4\xc5\xc6\xc7\xc8\xc9\xca\xcb\xcc\xcd\xce\xcf\xd0\xd1\xd2\xd3\xd4\xd5\xd6\xd7\xd8\xd9\xda\xdb\xdc\xdd\xde\xdf\xe0\xe1\xe2\xe3\xe4\xe5\xe6\xe7\xe8\xe9\xea\xeb\xec\xed\xee\xef\xf0\xf1\xf2\xf3\xf4\xf5\xf6\xf7\xf8\xf9\xfa\xfb\xfc\xfd\xfe\xff"

func inPlaceLowercase(d []byte) []byte {
	for i, v := range d {
		d[i] = toLowerTable[v]
	}
	return d
}

func (header *Header) ContentLength() int {
	v, ok := header.Get(HeaderContentLength)
	if !ok {
		return -1
	}
	n, e := strconv.ParseInt(utils.S(v), 10, 64)
	if e == nil {
		return int(n)
	}
	return -1
}

func (header *Header) SetContentLength(v int64) {
	header.Set(HeaderContentLength, utils.B(strconv.FormatInt(v, 10)))
}

func (header *Header) SetContentType(v string) {
	header.Set(HeaderContentType, utils.B(v))
}

func (header *Header) ContentType() []byte {
	v, _ := header.Get(HeaderContentType)
	return v
}

func (header *Header) keyToLower(k string) string {
	header.buf.Reset()
	for _, b := range []byte(k) {
		header.buf.WriteByte(toLowerTable[b])
	}
	return header.buf.String()
}

func (header *Header) Get(key string) ([]byte, bool) {
	if !header.fromOutSide {
		return header.Kvs.Get(key)
	}
	return header.Kvs.Get(header.keyToLower(key))
}

func (header *Header) GetAll(key string) [][]byte {
	if !header.fromOutSide {
		return header.Kvs.GetAll(key)
	}
	return header.Kvs.GetAll(header.keyToLower(key))
}

func (header *Header) Set(key string, v []byte) *utils.KvItem {
	if !header.fromOutSide {
		return header.Kvs.Set(key, v)
	}
	return header.Kvs.Set(header.keyToLower(key), v)
}

func (header *Header) SetString(key, v string) *utils.KvItem {
	if !header.fromOutSide {
		return header.Kvs.SetString(key, v)
	}
	return header.Kvs.SetString(header.keyToLower(key), v)
}

func (header *Header) Del(key string) {
	if !header.fromOutSide {
		header.Kvs.Del(key)
		return
	}
	header.Kvs.Del(header.keyToLower(key))
}

func (header *Header) Reset() {
	header.Kvs.Reset()
	header.fromOutSide = false
}
