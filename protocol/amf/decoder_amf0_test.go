package amf

import (
	"bytes"
	"testing"
)

func TestDecodeAmf0Number(t *testing.T) {
	buf := bytes.NewReader([]byte{0x00, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33})
	expect := float64(1.2)

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test number interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0Number(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test number interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0Number(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf0BooleanTrue(t *testing.T) {
	buf := bytes.NewReader([]byte{0x01, 0x01})
	expect := true

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test boolean interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0Boolean(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test boolean interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0Boolean(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf0BooleanFalse(t *testing.T) {
	buf := bytes.NewReader([]byte{0x01, 0x00})
	expect := false

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test boolean interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0Boolean(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test boolean interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0Boolean(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf0String(t *testing.T) {
	buf := bytes.NewReader([]byte{0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f})
	expect := "foo"

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test string interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0String(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test string interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0String(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf0Object(t *testing.T) {
	buf := bytes.NewReader([]byte{0x03, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x02, 0x00, 0x03, 0x62, 0x61, 0x72, 0x00, 0x00, 0x09})

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	obj, ok := got.(Object)
	if ok != true {
		t.Errorf("expected result to cast to object")
	}
	if obj["foo"] != "bar" {
		t.Errorf("expected {'foo'='bar'}, got %v", obj)
	}

	// Test object interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0Object(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	obj, ok = got.(Object)
	if ok != true {
		t.Errorf("expected result to cast to object")
	}
	if obj["foo"] != "bar" {
		t.Errorf("expected {'foo'='bar'}, got %v", obj)
	}

	// Test object interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0Object(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	obj, ok = got.(Object)
	if ok != true {
		t.Errorf("expected result to cast to object")
	}
	if obj["foo"] != "bar" {
		t.Errorf("expected {'foo'='bar'}, got %v", obj)
	}
}

func TestDecodeAmf0Null(t *testing.T) {
	buf := bytes.NewReader([]byte{0x05})

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if got != nil {
		t.Errorf("expect nil got %v", got)
	}

	// Test null interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0Null(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if got != nil {
		t.Errorf("expect nil got %v", got)
	}
}

func TestDecodeAmf0Undefined(t *testing.T) {
	buf := bytes.NewReader([]byte{0x06})

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if got != nil {
		t.Errorf("expect nil got %v", got)
	}

	// Test undefined interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0Undefined(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if got != nil {
		t.Errorf("expect nil got %v", got)
	}
}

/*
func TestDecodeReference(t *testing.T) {
	buf := bytes.NewReader([]byte{0x03, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x07, 0x00, 0x00, 0x00, 0x00, 0x09})

	dec := &Decoder{}

	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	obj, ok := got.(Object)
	if ok != true {
		t.Errorf("expected result to cast to object")
	}

	_, ok2 := obj["foo"].(Object)
	if ok2 != true {
		t.Errorf("expected foo value to cast to object")
	}
}
*/

func TestDecodeAmf0EcmaArray(t *testing.T) {
	buf := bytes.NewReader([]byte{0x08, 0x00, 0x00, 0x00, 0x01, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x02, 0x00, 0x03, 0x62, 0x61, 0x72, 0x00, 0x00, 0x09})

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	obj, ok := got.(Object)
	if ok != true {
		t.Errorf("expected result to cast to object")
	}
	if obj["foo"] != "bar" {
		t.Errorf("expected {'foo'='bar'}, got %v", obj)
	}

	// Test ecma array interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0EcmaArray(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	obj, ok = got.(Object)
	if ok != true {
		t.Errorf("expected result to cast to object")
	}
	if obj["foo"] != "bar" {
		t.Errorf("expected {'foo'='bar'}, got %v", obj)
	}

	// Test ecma array interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0EcmaArray(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	obj, ok = got.(Object)
	if ok != true {
		t.Errorf("expected result to cast to ecma array")
	}
	if obj["foo"] != "bar" {
		t.Errorf("expected {'foo'='bar'}, got %v", obj)
	}
}

func TestDecodeAmf0StrictArray(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0a, 0x00, 0x00, 0x00, 0x03, 0x00, 0x40, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x03, 0x66, 0x6f, 0x6f, 0x05})

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	arr, ok := got.(Array)
	if ok != true {
		t.Errorf("expected result to cast to strict array")
	}
	if arr[0] != float64(5) {
		t.Errorf("expected array[0] to be 5, got %v", arr[0])
	}
	if arr[1] != "foo" {
		t.Errorf("expected array[1] to be 'foo', got %v", arr[1])
	}
	if arr[2] != nil {
		t.Errorf("expected array[2] to be nil, got %v", arr[2])
	}

	// Test strict array interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0StrictArray(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	arr, ok = got.(Array)
	if ok != true {
		t.Errorf("expected result to cast to strict array")
	}
	if arr[0] != float64(5) {
		t.Errorf("expected array[0] to be 5, got %v", arr[0])
	}
	if arr[1] != "foo" {
		t.Errorf("expected array[1] to be 'foo', got %v", arr[1])
	}
	if arr[2] != nil {
		t.Errorf("expected array[2] to be nil, got %v", arr[2])
	}

	// Test strict array interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0StrictArray(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	arr, ok = got.(Array)
	if ok != true {
		t.Errorf("expected result to cast to strict array")
	}
	if arr[0] != float64(5) {
		t.Errorf("expected array[0] to be 5, got %v", arr[0])
	}
	if arr[1] != "foo" {
		t.Errorf("expected array[1] to be 'foo', got %v", arr[1])
	}
	if arr[2] != nil {
		t.Errorf("expected array[2] to be nil, got %v", arr[2])
	}
}

func TestDecodeAmf0Date(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0b, 0x40, 0x14, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
	expect := float64(5)

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test date interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0Date(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test date interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0Date(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf0LongString(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0c, 0x00, 0x00, 0x00, 0x03, 0x66, 0x6f, 0x6f})
	expect := "foo"

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test long string interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0LongString(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test long string interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0LongString(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf0Unsupported(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0d})

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if got != nil {
		t.Errorf("expect nil got %v", got)
	}

	// Test unsupported interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0Unsupported(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if got != nil {
		t.Errorf("expect nil got %v", got)
	}
}

func TestDecodeAmf0XmlDocument(t *testing.T) {
	buf := bytes.NewReader([]byte{0x0f, 0x00, 0x00, 0x00, 0x03, 0x66, 0x6f, 0x6f})
	expect := "foo"

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test long string interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0XmlDocument(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	// Test long string interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0XmlDocument(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf0TypedObject(t *testing.T) {

	buf := bytes.NewReader([]byte{
		0x10, 0x00, 0x0F, 'o', 'r', 'g',
		'.', 'a', 'm', 'f', '.', 'A',
		'S', 'C', 'l', 'a', 's', 's',
		0x00, 0x03, 'b', 'a', 'z', 0x05,
		0x00, 0x03, 'f', 'o', 'o', 0x02,
		0x00, 0x03, 'b', 'a', 'r', 0x00,
		0x00, 0x09,
	})

	dec := &Decoder{}

	// Test main interface
	got, err := dec.DecodeAmf0(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	tobj, ok := got.(TypedObject)
	if ok != true {
		t.Errorf("expected result to cast to typed object, got %+v", tobj)
	}
	if tobj.Type != "org.amf.ASClass" {
		t.Errorf("expected typed object type to be 'class', got %v", tobj.Type)
	}
	if tobj.Object["foo"] != "bar" {
		t.Errorf("expected typed object object foo to eql bar, got %v", tobj.Object["foo"])
	}
	if tobj.Object["baz"] != nil {
		t.Errorf("expected typed object object baz to nil, got %v", tobj.Object["baz"])
	}

	// Test typed object interface with marker
	buf.Seek(0, 0)
	got, err = dec.DecodeAmf0TypedObject(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	tobj, ok = got.(TypedObject)
	if ok != true {
		t.Errorf("expected result to cast to typed object, got %+v", tobj)
	}
	if tobj.Type != "org.amf.ASClass" {
		t.Errorf("expected typed object type to be 'class', got %v", tobj.Type)
	}
	if tobj.Object["foo"] != "bar" {
		t.Errorf("expected typed object object foo to eql bar, got %v", tobj.Object["foo"])
	}
	if tobj.Object["baz"] != nil {
		t.Errorf("expected typed object object baz to nil, got %v", tobj.Object["baz"])
	}

	// Test typed object interface without marker
	buf.Seek(1, 0)
	got, err = dec.DecodeAmf0TypedObject(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	tobj, ok = got.(TypedObject)
	if ok != true {
		t.Errorf("expected result to cast to typed object, got %+v", tobj)
	}
	if tobj.Type != "org.amf.ASClass" {
		t.Errorf("expected typed object type to be 'class', got %v", tobj.Type)
	}
	if tobj.Object["foo"] != "bar" {
		t.Errorf("expected typed object object foo to eql bar, got %v", tobj.Object["foo"])
	}
	if tobj.Object["baz"] != nil {
		t.Errorf("expected typed object object baz to nil, got %v", tobj.Object["baz"])
	}
}
