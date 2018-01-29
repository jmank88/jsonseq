package jsonseq

import (
	"fmt"
	"io"
	"os"
	"strings"
)

func ExampleWriteRecord() {
	_ = WriteRecord(os.Stdout, []byte(`{"id":1}`))
	_ = WriteRecord(os.Stdout, []byte(`{"id":2}`))
	_ = WriteRecord(os.Stdout, []byte(`{"id":3}`))

	// Output:
	// {"id":1}
	// {"id":2}
	// {"id":3}
	//
}

func ExampleNewEncoder() {
	encoder := NewEncoder(os.Stdout)
	_ = encoder.Encode("Test")
	_ = encoder.Encode(123.456)
	_ = encoder.Encode(struct{ Id int }{Id: 1})

	// Output:
	// "Test"
	// 123.456
	// {"Id":1}
	//
}

func ExampleDecoder_Decode() {
	d := NewDecoder(strings.NewReader(`{"id":1} 12341234 true discarded junk`))
	for {
		var i interface{}
		if err := d.Decode(&i); err != nil {
			if err == io.EOF {
				break
			}
			fmt.Println(err)
		} else {
			fmt.Println(i)
		}
	}

	// Output:
	// map[id:1]
	// invalid record: "1234"
	// 1234
	// true
}
