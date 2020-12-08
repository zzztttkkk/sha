package suna

// https://github.com/valyala/fasthttp/blob/c2542e5acf973cb1a2ab82d74dcb66f7afcb968b/args.go#L527
const hex2intTable = "\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x00\x01\x02\x03\x04\x05\x06\a\b\t\x10\x10\x10\x10\x10\x10\x10\n\v\f\r\x0e\x0f\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\n\v\f\r\x0e\x0f\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10\x10"

func decodeURI(src []byte) []byte {
	cursor := 0
	var end = len(src)
	for i := 0; i < len(src); i++ {
		c := src[i]
		if c == '%' {
			if i+2 >= len(src) {
				src[cursor] = c
				cursor++
				continue
			}

			x2 := hex2intTable[src[i+2]]
			x1 := hex2intTable[src[i+1]]
			if x1 == 16 || x2 == 16 {
				src[cursor] = '%'
				cursor++
				end -= 2
			} else {
				src[cursor] = x1<<4 | x2
				cursor++
				i += 2
				end -= 2
			}
		} else {
			src[cursor] = c
			cursor++
		}
	}
	return src[:end]
}

func decodeURIFormed(src []byte) []byte {
	cursor := 0
	var end = len(src)
	for i := 0; i < len(src); i++ {
		c := src[i]
		if c == '%' {
			if i+2 >= len(src) {
				src[cursor] = c
				cursor++
				continue
			}

			x2 := hex2intTable[src[i+2]]
			x1 := hex2intTable[src[i+1]]
			if x1 == 16 || x2 == 16 {
				src[cursor] = '%'
				cursor++
				end -= 2
			} else {
				src[cursor] = x1<<4 | x2
				cursor++
				i += 2
				end -= 2
			}
		} else if c == '+' {
			src[cursor] = ' '
			cursor++
		} else {
			src[cursor] = c
			cursor++
		}
	}
	return src[:end]
}

const upperhex = "0123456789ABCDEF"

var noEscapedURI [256]bool

func init() {
	for b := 'A'; b < 'z'; b++ {
		noEscapedURI[b] = true
	}
	for b := '0'; b < '9'; b++ {
		noEscapedURI[b] = true
	}
	for _, b := range ";,/?:@&=+$-_.!~*'()#" {
		noEscapedURI[b] = true
	}
}

func encodeURI(v []byte, buf *[]byte) {
	for _, b := range v {
		if noEscapedURI[b] {
			*buf = append(*buf, b)
			continue
		}
		*buf = append(*buf, '%', upperhex[b>>4], upperhex[b&0xf])
	}
}

var noEscapedURIComponent [256]bool

func init() {
	for b := 'A'; b < 'z'; b++ {
		noEscapedURIComponent[b] = true
	}
	for b := '0'; b < '9'; b++ {
		noEscapedURIComponent[b] = true
	}
	for _, b := range "-_.!~*'()" {
		noEscapedURIComponent[b] = true
	}
}

func encodeURIComponent(v []byte, buf *[]byte) {
	for _, b := range v {
		if noEscapedURIComponent[b] {
			*buf = append(*buf, b)
			continue
		}
		*buf = append(*buf, '%', upperhex[b>>4], upperhex[b&0xf])
	}
}
