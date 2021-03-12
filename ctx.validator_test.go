package sha

import (
	"fmt"
	"github.com/zzztttkkk/sha/utils"
	"strconv"
	"strings"
	"testing"
	"time"
)

type Time time.Time

func (t *Time) FormValue(data []byte) bool {
	v, e := strconv.ParseInt(utils.S(data), 10, 64)
	if e != nil {
		return false
	}
	if v < 6000 {
		return false
	}
	*t = Time(time.Unix(v-6000, 0))
	return true
}

func (t *Time) String() string {
	return (*time.Time)(t).String()
}

type TV01Form struct {
	//NumbersPtr *[]int64
	//Numbers    []int64 `validator:",w=body"`
	//StrPtr     *string
	//Str        string
	CFTime       Time
	CFTimePtr    *Time
	CFTimePtrPtr **Time
}

func (t *TV01Form) String() string {
	sb := strings.Builder{}

	//sb.WriteString("NumbersPtr: ")
	//if t.NumbersPtr == nil {
	//	sb.WriteString("nil\n")
	//} else {
	//	sb.WriteString(fmt.Sprintf("%v\n", *t.NumbersPtr))
	//}
	//
	//sb.WriteString(fmt.Sprintf("Numbers: %v\n", t.Numbers))
	//
	//sb.WriteString("StrPtr: ")
	//if t.StrPtr == nil {
	//	sb.WriteString("nil\n")
	//} else {
	//	sb.WriteString(fmt.Sprintf("%p<%s>\n", t.StrPtr, *t.StrPtr))
	//}
	//
	//sb.WriteString(fmt.Sprintf("Str: %s\n", t.Str))

	sb.WriteString(fmt.Sprintf("CFTime: %s\n", &t.CFTime))

	sb.WriteString("CFTimePtr: ")
	if t.CFTimePtr == nil {
		sb.WriteString("nil\n")
	} else {
		sb.WriteString(fmt.Sprintf("%s\n", t.CFTimePtr))
	}

	return sb.String()
}

func (t *TV01Form) DefaultNumbers() interface{} { return []int64{1, 2, 3} }

func TestValidator(t *testing.T) {
	ListenAndServe(
		"127.0.0.1:5986",
		RequestHandlerFunc(func(ctx *RequestCtx) {
			var form TV01Form

			if err := ctx.Validate(&form); err != nil {
				ctx.SetStatus(err.StatusCode())
				ctx.WriteString(err.Error())
				return
			}

			fmt.Print(&form)
		}),
	)
}
