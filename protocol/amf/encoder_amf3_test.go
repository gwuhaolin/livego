package amf

import (
	"bytes"
	"testing"
)

func TestEncodeAmf3EmptyString(t *testing.T) {
	enc := new(Encoder)

	buf := new(bytes.Buffer)
	expect := []byte{0x01}

	_, err := enc.EncodeAmf3String(buf, "", false)
	if err != nil {
		t.Errorf("%s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf3Undefined(t *testing.T) {
	enc := new(Encoder)

	buf := new(bytes.Buffer)
	expect := []byte{0x00}

	_, err := enc.EncodeAmf3Undefined(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf3Null(t *testing.T) {
	enc := new(Encoder)

	buf := new(bytes.Buffer)
	expect := []byte{0x01}

	_, err := enc.EncodeAmf3(buf, nil)
	if err != nil {
		t.Errorf("%s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf3False(t *testing.T) {
	enc := new(Encoder)

	buf := new(bytes.Buffer)
	expect := []byte{0x02}

	_, err := enc.EncodeAmf3(buf, false)
	if err != nil {
		t.Errorf("%s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf3True(t *testing.T) {
	enc := new(Encoder)

	buf := new(bytes.Buffer)
	expect := []byte{0x03}

	_, err := enc.EncodeAmf3(buf, true)
	if err != nil {
		t.Errorf("%s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf3Integer(t *testing.T) {
	enc := new(Encoder)

	for _, tc := range u29TestCases {
		buf := new(bytes.Buffer)
		_, err := enc.EncodeAmf3Integer(buf, tc.value, false)
		if err != nil {
			t.Errorf("EncodeAmf3Integer error: %s", err)
		}
		got := buf.Bytes()
		if !bytes.Equal(tc.expect, got) {
			t.Errorf("EncodeAmf3Integer expect n %x got %x", tc.value, got)
		}
	}

	buf := new(bytes.Buffer)
	expect := []byte{0x04, 0x80, 0xFF, 0xFF, 0xFF}

	n, err := enc.EncodeAmf3(buf, uint32(4194303))
	if err != nil {
		t.Errorf("%s", err)
	}
	if n != 5 {
		t.Errorf("expected to write 5 bytes, actual %d", n)
	}
	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf3Double(t *testing.T) {
	enc := new(Encoder)

	buf := new(bytes.Buffer)
	expect := []byte{0x05, 0x3f, 0xf3, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}

	_, err := enc.EncodeAmf3(buf, float64(1.2))
	if err != nil {
		t.Errorf("%s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf3String(t *testing.T) {
	enc := new(Encoder)

	buf := new(bytes.Buffer)
	expect := []byte{0x06, 0x07, 'f', 'o', 'o'}

	_, err := enc.EncodeAmf3(buf, "foo")
	if err != nil {
		t.Errorf("%s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf3Array(t *testing.T) {
	enc := new(Encoder)
	buf := new(bytes.Buffer)
	expect := []byte{0x09, 0x13, 0x01,
		0x06, 0x03, '1',
		0x06, 0x03, '2',
		0x06, 0x03, '3',
		0x06, 0x03, '4',
		0x06, 0x03, '5',
		0x06, 0x03, '6',
		0x06, 0x03, '7',
		0x06, 0x03, '8',
		0x06, 0x03, '9',
	}

	arr := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9"}
	_, err := enc.EncodeAmf3(buf, arr)
	if err != nil {
		t.Errorf("err: %s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer: %+v, got: %+v", expect, buf.Bytes())
	}
}

func TestEncodeAmf3Object(t *testing.T) {
	enc := new(Encoder)
	buf := new(bytes.Buffer)
	expect := []byte{
		0x0a, 0x23, 0x1f, 'o', 'r', 'g', '.', 'a',
		'm', 'f', '.', 'A', 'S', 'C', 'l', 'a',
		's', 's', 0x07, 'b', 'a', 'z', 0x07, 'f',
		'o', 'o', 0x01, 0x06, 0x07, 'b', 'a', 'r',
	}

	to := *NewTypedObject()
	to.Type = "org.amf.ASClass"
	to.Object["foo"] = "bar"
	to.Object["baz"] = nil

	_, err := enc.EncodeAmf3(buf, to)
	if err != nil {
		t.Errorf("err: %s", err)
	}

	if bytes.Compare(buf.Bytes(), expect) != 0 {
		t.Errorf("expected buffer:\n%#v\ngot:\n%#v", expect, buf.Bytes())
	}
}
