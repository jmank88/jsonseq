package jsonseq

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func ExampleWriter_WriteRecord() {
	WriteRecord(os.Stdout, []byte(`{"id":1}`))
	WriteRecord(os.Stdout, []byte(`{"id":2}`))
	WriteRecord(os.Stdout, []byte(`{"id":3}`))

	// Output:
	// {"id":1}
	// {"id":2}
	// {"id":3}
	//
}

func ExampleReader_ReadRecord() {
	s := bufio.NewScanner(strings.NewReader(`{"id":1}
12341234
`))
	s.Split(ScanRecord)
	for s.Scan() {
		b, ok := RecordValue(s.Bytes())
		if !ok {
			fmt.Print("partial record: ")
		}
		fmt.Println(strings.TrimSpace(string(b)))
	}
	// Output:
	// {"id":1}
	// partial record: 1234
	// 1234
}
