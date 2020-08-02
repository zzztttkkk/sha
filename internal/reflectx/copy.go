package reflectx

import (
	"fmt"
	"reflect"
)

func Copy(dist, src reflect.Value) {
	dT := dist.Type()
	sT := src.Type()

	if dT != sT {
		panic(fmt.Errorf(""))
	}



}
