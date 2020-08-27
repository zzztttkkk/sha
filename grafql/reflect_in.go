package grafql

import (
	"fmt"
	"github.com/zzztttkkk/suna/validator"
	"reflect"
)

func NewInObject(value reflect.Value) {
	rules := validator.GetRules(value)

	fmt.Println(rules)

}
