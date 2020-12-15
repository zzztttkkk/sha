package sqlx

import (
	"encoding/json"
	"fmt"
	"testing"
)

func TestBytes_UnmarshalJSON(t *testing.T) {
	a := JsonBytes("\"\\")
	fmt.Println(len(a))
	j, _ := json.Marshal(a)
	fmt.Println(string(j), len(j))
	var b JsonBytes
	_ = json.Unmarshal(j, &b)
	fmt.Println(string(b), len(b))

	s := "\"\\"
	j, _ = json.Marshal(s)
	fmt.Println(string(j), len(j))
	var bs string
	_ = json.Unmarshal(j, &bs)
	fmt.Println(string(b), len(b))
}
