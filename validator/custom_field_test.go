package validator

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
)

type AInt int64

func (a *AInt) FromBytes(v []byte) error {
	n, e := strconv.ParseInt(string(v), 10, 64)
	if e != nil {
		return e
	}
	*a = AInt(n)
	return nil
}

func (a *AInt) Validate() error {
	if *a < 100 {
		return errors.New("AInt must greater than 100")
	}
	return nil
}

func TestField(t *testing.T) {
	type Form struct {
		A    AInt
		ALst []AInt `validator:",size=2-"`
	}

	var f Form
	f.A = 199
	f.ALst = append(f.ALst, 134, 1000)

	fmt.Println(ValidateStruct(&f))
}
