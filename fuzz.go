// +build gofuzz

package jsonseq

import "bytes"

/*
Fuzzing function for github.com/dvyukov/go-fuzz.

Generate jsonseq-fuzz.zip:

 	 go generate -tags gofuzz

Fuzz:

    go-fuzz -bin=jsonseq-fuzz.zip -workdir=testdata
*/

//go:generate go-fuzz-build github.com/jmank88/jsonseq

func Fuzz(data []byte) int {
	var i interface{}
	if NewDecoder(bytes.NewReader(data)).Decode(&i) != nil {
		return 0
	}
	return 1
}
