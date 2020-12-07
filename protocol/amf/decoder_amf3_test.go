package amf

import (
	"bytes"
	"testing"
)

type u29TestCase struct {
	value  uint32
	expect []byte
}

var u29TestCases = []u29TestCase{
	{1, []byte{0x01}},
	{2, []byte{0x02}},
	{127, []byte{0x7F}},
	{128, []byte{0x81, 0x00}},
	{255, []byte{0x81, 0x7F}},
	{256, []byte{0x82, 0x00}},
	{0x3FFF, []byte{0xFF, 0x7F}},
	{0x4000, []byte{0x81, 0x80, 0x00}},
	{0x7FFF, []byte{0x81, 0xFF, 0x7F}},
	{0x8000, []byte{0x82, 0x80, 0x00}},
	{0x1FFFFF, []byte{0xFF, 0xFF, 0x7F}},
	{0x200000, []byte{0x80, 0xC0, 0x80, 0x00}},
	{0x3FFFFF, []byte{0x80, 0xFF, 0xFF, 0xFF}},
	{0x400000, []byte{0x81, 0x80, 0x80, 0x00}},
	{0x0FFFFFFF, []byte{0xBF, 0xFF, 0xFF, 0xFF}},
}

func TestDecodeAmf3Undefined(t *testing.T) {
	buf := bytes.NewReader([]byte{0x00})

	dec := new(Decoder)

	got, err := dec.DecodeAmf3(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if got != nil {
		t.Errorf("expect nil got %v", got)
	}
}

func TestDecodeAmf3Null(t *testing.T) {
	buf := bytes.NewReader([]byte{0x01})

	dec := new(Decoder)

	got, err := dec.DecodeAmf3(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if got != nil {
		t.Errorf("expect nil got %v", got)
	}
}

func TestDecodeAmf3False(t *testing.T) {
	buf := bytes.NewReader([]byte{0x02})
	expect := false

	dec := new(Decoder)

	got, err := dec.DecodeAmf3(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf3True(t *testing.T) {
	buf := bytes.NewReader([]byte{0x03})
	expect := true

	dec := new(Decoder)

	got, err := dec.DecodeAmf3(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeU29(t *testing.T) {
	dec := new(Decoder)

	for _, tc := range u29TestCases {
		buf := bytes.NewBuffer(tc.expect)
		n, err := dec.decodeU29(buf)
		if err != nil {
			t.Errorf("DecodeAmf3Integer error: %s", err)
		}
		if n != tc.value {
			t.Errorf("DecodeAmf3Integer expect n %x got %x", tc.value, n)
		}
	}
}

func TestDecodeAmf3Integer(t *testing.T) {
	dec := new(Decoder)

	buf := bytes.NewReader([]byte{0x04, 0xFF, 0xFF, 0x7F})
	expect := int32(2097151)

	got, err := dec.DecodeAmf3(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	buf.Seek(0, 0)
	got, err = dec.DecodeAmf3Integer(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}

	buf.Seek(1, 0)
	got, err = dec.DecodeAmf3Integer(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf3Double(t *testing.T) {
	buf := bytes.NewReader([]byte{0x05, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33})
	expect := float64(1.2)

	dec := new(Decoder)

	got, err := dec.DecodeAmf3(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf3String(t *testing.T) {
	buf := bytes.NewReader([]byte{0x06, 0x07, 'f', 'o', 'o'})
	expect := "foo"

	dec := new(Decoder)

	got, err := dec.DecodeAmf3(buf)
	if err != nil {
		t.Errorf("%s", err)
	}
	if expect != got {
		t.Errorf("expect %v got %v", expect, got)
	}
}

func TestDecodeAmf3Array(t *testing.T) {
	buf := bytes.NewReader([]byte{0x09, 0x13, 0x01,
		0x06, 0x03, '1',
		0x06, 0x03, '2',
		0x06, 0x03, '3',
		0x06, 0x03, '4',
		0x06, 0x03, '5',
		0x06, 0x03, '6',
		0x06, 0x03, '7',
		0x06, 0x03, '8',
		0x06, 0x03, '9',
	})

	dec := new(Decoder)
	expect := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	got, err := dec.DecodeAmf3Array(buf, true)
	if err != nil {
		t.Errorf("err: %s", err)
	}

	for i, v := range expect {
		if got[i] != v {
			t.Errorf("expected array element %d to be %v, got %v", i, v, got[i])
		}
	}
}

func TestDecodeAmf3Object(t *testing.T) {
	buf := bytes.NewReader([]byte{
		0x0a, 0x23, 0x1f, 'o', 'r', 'g', '.', 'a',
		'm', 'f', '.', 'A', 'S', 'C', 'l', 'a',
		's', 's', 0x07, 'b', 'a', 'z', 0x07, 'f',
		'o', 'o', 0x01, 0x06, 0x07, 'b', 'a', 'r',
	})

	dec := new(Decoder)
	got, err := dec.DecodeAmf3(buf)
	if err != nil {
		t.Errorf("err: %s", err)
	}

	to, ok := got.(Object)
	if ok != true {
		t.Error("unable to cast object as typed object")
	}

	if to["foo"] != "bar" {
		t.Errorf("expected foo to be bar, got: %+v", to["foo"])
	}

	if to["baz"] != nil {
		t.Errorf("expected baz to be nil, got: %+v", to["baz"])
	}
}
