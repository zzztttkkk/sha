package internal

import (
	"fmt"
	"testing"
)

func TestStatus16_Add(t *testing.T) {
	var s Status16
	for i := 1; i < 16; i++ {
		s.Add(uint8(i))
	}
	for i := 1; i < 16; i++ {
		fmt.Println(i, s.Has(uint8(i)))
	}

	for i := 1; i < 16; i++ {
		if i%5 == 0 {
			s.Del(uint8(i))
		}
	}
	for i := 1; i < 16; i++ {
		fmt.Println(i, s.Has(uint8(i)))
	}
}
