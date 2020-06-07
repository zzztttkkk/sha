package utils

import (
	"errors"
	"fmt"
)

type Status uint64

var statusOutOfRangeError = errors.New("snow.status: out of range. [0, 64]")

// n 0 - 64
func NewStatus(ns ...uint8) Status {
	s := Status(0)
	for _, v := range ns {
		s = s.Add(v)
	}
	return s
}

func (m Status) Has(status uint8) bool {
	if status == 0 {
		return true
	}

	if status > 64 {
		panic(statusOutOfRangeError)
	}

	return m&(1<<(status-1)) != 0
}

func (m Status) Add(status uint8) Status {
	if status == 0 {
		return m
	}

	if status > 64 {
		panic(statusOutOfRangeError)
	}

	return m | (1 << (status - 1))
}

func (m Status) Del(status uint8) Status {
	if status == 0 {
		return m
	}

	if status > 64 {
		panic(statusOutOfRangeError)
	}
	return m & (^(1 << (status - 1)))
}

func (m Status) String() string {
	return fmt.Sprintf("%064b", m)
}
