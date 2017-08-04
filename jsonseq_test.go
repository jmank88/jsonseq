package jsonseq

import (
	"bufio"
	"bytes"
	"math"
	"regexp"
	"testing"
)

func TestDecode(t *testing.T) {

	type Coord struct {
		X int
		Y int
	}

	br := bytes.NewReader([]byte("\u001e{\"x\":1,\"y\":2}\n\u001e { \"x\":3, \"y\":4 } \n\u001e{\"x\":5,\"y\":6}\n"))
	d := NewDecoder(br)
	for i := 0; i <= 2; i++ {
		xx := 2*i + 1
		xy := xx + 1
		c := &Coord{}
		err := d.Decode(c)
		if err != nil {
			t.Errorf("decode obj %d failed: %s", i, err)
		}
		if c.X != xx || c.Y != xy {
			t.Errorf("decode obj %d failed, expected (%d,%d), got (%d,%d)", i, xx, xy, c.X, c.Y)
		}
	}

}

func TestDecodeTypes(t *testing.T) {

	type TestData struct {
		B bool
		F float64
		S string
		A []interface{}
		O map[string]interface{}
	}

	br := bytes.NewReader([]byte("\u001e{\"B\":true,\"F\":3.14159}\n\u001e { \"S\":\"covfefe\", \"A\": [1,2,3] } \n\u001e{\"O\":{\"a\":1,\"b\":\"two\"} }\n"))

	d := NewDecoder(br)
	for i := 0; i <= 2; i++ {
		obj := &TestData{}
		err := d.Decode(obj)
		if err != nil {
			t.Errorf("decode obj %d failed: %s", i, err)
		}
		switch i {
		case 0:
			if !obj.B {
				t.Errorf("decode obj 1 boolean failed: expected true, got %b", obj.B)
			}
			if math.Abs(obj.F-3.14159) > 0.000001 {
				t.Errorf("decode obj 1 float failed: expected 3.14159, got %f", obj.F)
			}
		case 1:
			if obj.S != "covfefe" {
				t.Errorf("decode obj 2 string failed: expected covfefe, got %v", obj.S)
			}
			a1, ok1 := obj.A[0].(float64)
			a2, ok2 := obj.A[1].(float64)
			a3, ok3 := obj.A[2].(float64)
			if !(ok1 && ok2 && ok3 && a1 == 1 && a2 == 2 && a3 == 3) {
				t.Errorf("decode obj 2 array failed: expected [1 2 3] got %v", obj.A)
			}
		case 2:
			m := obj.O
			a, ok := m["a"].(float64)
			if !ok || a != 1 {
				t.Errorf("decode obj 3 nested object field \"a\" failed: expected {\"a\": 1}, got %v", m["a"])
			}
			b, ok := m["b"].(string)
			if !ok || b != "two" {
				t.Errorf("decode obj 3 nested object field \"b\" failed: expected {\"b\": \"two\"}, got %v", m["b"])
			}
		}
	}
}

// Independent implementation of record slicing using regexp, just for testing.
var slicer = regexp.MustCompile("\u001e([^\u001e\n]*)\n")

func sliceup(buf *bytes.Buffer) [][]byte {
	var result [][]byte
	txt := buf.Bytes()
	tuples := slicer.FindAllIndex(txt, -1)
	for _, se := range tuples {
		s := se[0] + 1
		e := se[1] - 1
		result = append(result, txt[s:e])
	}
	return result
}

func TestWriteRecord(t *testing.T) {

	sjson := "{\"s\":\"trivial\"}"
	json := []byte(sjson)

	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	for i := 0; i < 3; i++ {
		err := WriteRecord(w, json)
		if err != nil {
			t.Errorf("failed to write JSON record: %s", err)
		}
	}
	err := w.Flush()
	if err != nil {
		t.Errorf("error flushing JSON records: %s", err)
	}

	js := sliceup(&buf)
	if len(js) != 3 {
		t.Errorf("record write failed, expected 3 records got %d", len(js))
	}

	for i := 0; i < 2; i++ {
		s := string(js[i])
		if s != sjson {
			t.Errorf("record write failed, expected record 3 to be %s got %s", json, s)
		}
	}
}
