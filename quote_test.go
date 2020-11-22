package suna

import (
	"fmt"
	"testing"
)

func TestA(t *testing.T) {
	quotedArgShouldEscapeTable := func() [256]byte {
		// According to RFC 3986 §2.3
		var a [256]byte
		for i := 0; i < 256; i++ {
			a[i] = 1
		}

		// ALPHA
		for i := int('a'); i <= int('z'); i++ {
			a[i] = 0
		}
		for i := int('A'); i <= int('Z'); i++ {
			a[i] = 0
		}

		// DIGIT
		for i := int('0'); i <= int('9'); i++ {
			a[i] = 0
		}

		// Unreserved characters
		for _, v := range `-_.~` {
			a[v] = 0
		}

		return a
	}()

	quotedPathShouldEscapeTable := func() [256]byte {
		// The implementation here equal to net/url shouldEscape(s, encodePath)
		//
		// The RFC allows : @ & = + $ but saves / ; , for assigning
		// meaning to individual path segments. This package
		// only manipulates the path as a whole, so we allow those
		// last three as well. That leaves only ? to escape.
		var a = quotedArgShouldEscapeTable

		for _, v := range `$&+,/:;=@` {
			a[v] = 0
		}

		return a
	}()

	quotedHeaderShouldEscapeTable := func() [256]byte {
		// The implementation here equal to net/url shouldEscape(s, encodePath)
		//
		// The RFC allows : @ & = + $ but saves / ; , for assigning
		// meaning to individual path segments. This package
		// only manipulates the path as a whole, so we allow those
		// last three as well. That leaves only ? to escape.
		var a = quotedArgShouldEscapeTable

		for _, v := range "`~!@#$%^&*()_+-=[]{}\\; '\",./?<>:" {
			a[v] = 0
		}

		return a
	}()

	fmt.Printf("%q\n", quotedArgShouldEscapeTable)
	fmt.Printf("%q\n", quotedPathShouldEscapeTable)
	fmt.Printf("%q\n", quotedHeaderShouldEscapeTable)

}
