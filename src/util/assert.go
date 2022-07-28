package util

import "errors"

func Assert(b bool) {
	if !b {
		panic(errors.New("Assert"))
	}
}
