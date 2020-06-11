package utils

import (
	"errors"
	"fmt"
)

type Bitmap struct {
	v     []uint64
	count uint32
}

func NewBitmap(count uint32) *Bitmap {
	size := count / 64
	if count%64 != 0 {
		size++
	}
	return &Bitmap{
		v:     make([]uint64, size, size),
		count: count,
	}
}

var bitmapOverflowError = errors.New("snow.utils.bitmap: value overflow")

func (m *Bitmap) Has(ind uint32) bool {
	if ind > m.count {
		panic(bitmapOverflowError)
	}
	a := ind / 64
	b := ind % 64
	return m.v[a]&(1<<b) > 0
}

func (m *Bitmap) Add(ind uint32) {
	if ind > m.count {
		panic(bitmapOverflowError)
	}
	a := ind / 64
	b := ind % 64
	m.v[a] = m.v[a] | (1 << b)
}

func (m *Bitmap) Del(ind uint32) {
	if ind > m.count {
		panic(bitmapOverflowError)
	}
	a := ind / 64
	b := ind % 64
	m.v[a] = m.v[a] & (^(1 << b))
}

func (m *Bitmap) Reset() {
	size := m.count / 64
	if m.count%64 != 0 {
		size++
	}
	m.v = make([]uint64, size, size)
}

func (m *Bitmap) String(ind uint32) string {
	return fmt.Sprintf("%064b", m.v[ind])
}

func (m *Bitmap) Or(other *Bitmap) *Bitmap {
	c := m.count
	l := m
	s := other
	if c < other.count {
		c = other.count
		l = other
		s = m
	}
	nm := NewBitmap(c)
	copy(nm.v, l.v)

	for i, v := range s.v {
		nm.v[i] = nm.v[i] | v
	}
	return nm
}

func (m *Bitmap) And(other *Bitmap) *Bitmap {
	c := m.count
	s := m
	l := other
	if c > other.count {
		c = other.count
		s = other
		l = m
	}
	nm := NewBitmap(c)

	for i, v := range s.v {
		nm.v[i] = l.v[i] & v
	}
	return nm
}
