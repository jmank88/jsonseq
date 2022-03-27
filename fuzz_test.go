// go:build go1.18
package jsonseq

import (
	"bytes"
	"strings"
	"testing"
)

func FuzzDecEnc(f *testing.F) {
	f.Add(true, "", "", `{"id":1} 123412.34 true discarded junk`)
	f.Add(false, "", " ", `{"id":"1"} "1234"1234 true discarded junk`)
	f.Add(true, "", "\t", `{"id":1.0} 12341234 true discarded junk`)
	f.Add(false, "\t", "", `{"id":1,"foo":[{"bar":3},{}]} [1,2,3,4]1234 true discarded junk`)
	f.Fuzz(func(t *testing.T, encEscapeHTML bool, encPrefix, encIndent, data string) {
		var i interface{}
		if err := NewDecoder(strings.NewReader(data)).Decode(&i); err != nil {
			return // invalid
		}

		var b bytes.Buffer
		enc := NewEncoder(&b)
		enc.SetEscapeHTML(encEscapeHTML)
		enc.SetIndent(encPrefix, encIndent)
		if err := enc.Encode(i); err != nil {
			t.Error(err)
		}
	})
}
