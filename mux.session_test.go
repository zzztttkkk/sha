package sha

import (
	"fmt"
	"testing"
)

func TestSessionGenerate(t *testing.T) {
	var session Session
	SessionIDGenerator(&session)
	fmt.Println(session.String())
}
