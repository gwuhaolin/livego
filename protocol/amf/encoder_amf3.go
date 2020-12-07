package amf

import (
	"encoding/binary"
	"fmt"
	"io"
	"reflect"
	"sort"
	"time"
)

// amf3 polymorphic router

func (e *Encoder) EncodeAmf3(w io.Writer, val interface{}) (int, error) {
	if val == nil {
		return e.EncodeAmf3Null(w, true)
	}

	v := reflect.ValueOf(val)
	if !v.IsValid() {
		return e.EncodeAmf3Null(w, true)
	}

	switch v.Kind() {
	case reflect.String:
		return e.EncodeAmf3String(w, v.String(), true)
	case reflect.Bool:
		if v.Bool() {
			return e.EncodeAmf3True(w, true)
		} else {
			return e.EncodeAmf3False(w, true)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32:
		n := v.Int()
		if n >= 0 && n <= AMF3_INTEGER_MAX {
			return e.EncodeAmf3Integer(w, uint32(n), true)
		} else {
			return e.EncodeAmf3Double(w, float64(n), true)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32:
		n := v.Uint()
		if n <= AMF3_INTEGER_MAX {
			return e.EncodeAmf3Integer(w, uint32(n), true)
		} else {
			return e.EncodeAmf3Double(w, float64(n), true)
		}
	case reflect.Int64:
		return e.EncodeAmf3Double(w, float64(v.Int()), true)
	case reflect.Uint64:
		return e.EncodeAmf3Double(w, float64(v.Uint()), true)
	case reflect.Float32, reflect.Float64:
		return e.EncodeAmf3Double(w, float64(v.Float()), true)
	case reflect.Array, reflect.Slice:
		length := v.Len()
		arr := make(Array, length)
		for i := 0; i < length; i++ {
			arr[i] = v.Index(int(i)).Interface()
		}
		return e.EncodeAmf3Array(w, arr, true)
	case reflect.Map:
		obj, ok := val.(Object)
		if ok != true {
			return 0, fmt.Errorf("encode amf3: unable to create object from map")
		}

		to := *new(TypedObject)
		to.Object = obj

		return e.EncodeAmf3Object(w, to, true)
	}

	if tm, ok := val.(time.Time); ok {
		return e.EncodeAmf3Date(w, tm, true)
	}

	if to, ok := val.(TypedObject); ok {
		return e.EncodeAmf3Object(w, to, true)
	}

	return 0, fmt.Errorf("encode amf3: unsupported type %s", v.Type())
}

// marker: 1 byte 0x00
// no additional data
func (e *Encoder) EncodeAmf3Undefined(w io.Writer, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_UNDEFINED_MARKER); err != nil {
			return
		}
		n += 1
	}

	return
}

// marker: 1 byte 0x01
// no additional data
func (e *Encoder) EncodeAmf3Null(w io.Writer, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_NULL_MARKER); err != nil {
			return
		}
		n += 1
	}

	return
}

// marker: 1 byte 0x02
// no additional data
func (e *Encoder) EncodeAmf3False(w io.Writer, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_FALSE_MARKER); err != nil {
			return
		}
		n += 1
	}

	return
}

// marker: 1 byte 0x03
// no additional data
func (e *Encoder) EncodeAmf3True(w io.Writer, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_TRUE_MARKER); err != nil {
			return
		}
		n += 1
	}

	return
}

// marker: 1 byte 0x04
func (e *Encoder) EncodeAmf3Integer(w io.Writer, val uint32, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_INTEGER_MARKER); err != nil {
			return
		}
		n += 1
	}

	var m int
	m, err = e.encodeAmf3Uint29(w, val)
	if err != nil {
		return
	}
	n += m

	return
}

// marker: 1 byte 0x05
func (e *Encoder) EncodeAmf3Double(w io.Writer, val float64, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_DOUBLE_MARKER); err != nil {
			return
		}
		n += 1
	}

	err = binary.Write(w, binary.BigEndian, &val)
	if err != nil {
		return
	}
	n += 8

	return
}

// marker: 1 byte 0x06
// format:
// - u29 reference int. if reference, no more data. if not reference,
//   length value of bytes to read to complete string.
func (e *Encoder) EncodeAmf3String(w io.Writer, val string, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_STRING_MARKER); err != nil {
			return
		}
		n += 1
	}

	var m int

	m, err = e.encodeAmf3Utf8(w, val)
	if err != nil {
		return
	}
	n += m

	return
}

// marker: 1 byte 0x08
// format:
// - u29 reference int, if reference, no more data
// - timestamp double
func (e *Encoder) EncodeAmf3Date(w io.Writer, val time.Time, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_DATE_MARKER); err != nil {
			return
		}
		n += 1
	}

	if err = WriteMarker(w, 0x01); err != nil {
		return n, fmt.Errorf("amf3 encode: cannot encode u29 for array: %s", err)
	}
	n += 1

	u64 := float64(val.Unix()) * 1000.0
	err = binary.Write(w, binary.BigEndian, &u64)
	if err != nil {
		return n, fmt.Errorf("amf3 encode: unable to write date double: %s", err)
	}
	n += 8

	return
}

// marker: 1 byte 0x09
// format:
// - u29 reference int. if reference, no more data.
// - string representing associative array if present
// - n values (length of u29)
func (e *Encoder) EncodeAmf3Array(w io.Writer, val Array, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_ARRAY_MARKER); err != nil {
			return
		}
		n += 1
	}

	var m int
	length := uint32(len(val))
	u29 := uint32(length<<1) | 0x01

	m, err = e.encodeAmf3Uint29(w, u29)
	if err != nil {
		return n, fmt.Errorf("amf3 encode: cannot encode u29 for array: %s", err)
	}
	n += m

	m, err = e.encodeAmf3Utf8(w, "")
	if err != nil {
		return n, fmt.Errorf("amf3 encode: cannot encode empty string for array: %s", err)
	}
	n += m

	for _, v := range val {
		m, err := e.EncodeAmf3(w, v)
		if err != nil {
			return n, fmt.Errorf("amf3 encode: cannot encode array element: %s", err)
		}
		n += m
	}

	return
}

// marker: 1 byte 0x0a
// format: ugh
func (e *Encoder) EncodeAmf3Object(w io.Writer, val TypedObject, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_OBJECT_MARKER); err != nil {
			return
		}
		n += 1
	}

	m := 0

	trait := *NewTrait()
	trait.Type = val.Type
	trait.Dynamic = false
	trait.Externalizable = false

	for k, _ := range val.Object {
		trait.Properties = append(trait.Properties, k)
	}

	sort.Strings(trait.Properties)

	var u29 uint32 = 0x03
	if trait.Dynamic {
		u29 |= 0x02 << 2
	}

	if trait.Externalizable {
		u29 |= 0x01 << 2
	}

	u29 |= uint32(len(trait.Properties)) << 4

	m, err = e.encodeAmf3Uint29(w, u29)
	if err != nil {
		return n, fmt.Errorf("amf3 encode: cannot encode trait header for object: %s", err)
	}
	n += m

	m, err = e.encodeAmf3Utf8(w, trait.Type)
	if err != nil {
		return n, fmt.Errorf("amf3 encode: cannot encode trait type for object: %s", err)
	}
	n += m

	for _, prop := range trait.Properties {
		m, err = e.encodeAmf3Utf8(w, prop)
		if err != nil {
			return n, fmt.Errorf("amf3 encode: cannot encode trait property for object: %s", err)
		}
		n += m
	}

	if trait.Externalizable {
		return n, fmt.Errorf("amf3 encode: cannot encode externalizable object")
	}

	for _, prop := range trait.Properties {
		m, err = e.EncodeAmf3(w, val.Object[prop])
		if err != nil {
			return n, fmt.Errorf("amf3 encode: cannot encode sealed object value: %s", err)
		}
		n += m
	}

	if trait.Dynamic {
		for k, v := range val.Object {
			var foundProp bool = false
			for _, prop := range trait.Properties {
				if prop == k {
					foundProp = true
					break
				}
			}

			if foundProp != true {
				m, err = e.encodeAmf3Utf8(w, k)
				if err != nil {
					return n, fmt.Errorf("amf3 encode: cannot encode dynamic object property key: %s", err)
				}
				n += m

				m, err = e.EncodeAmf3(w, v)
				if err != nil {
					return n, fmt.Errorf("amf3 encode: cannot encode dynamic object value: %s", err)
				}
				n += m
			}

			m, err = e.encodeAmf3Utf8(w, "")
			if err != nil {
				return n, fmt.Errorf("amf3 encode: cannot encode dynamic object ending marker string: %s", err)
			}
			n += m
		}
	}

	return
}

// marker: 1 byte 0x0c
// format:
// - u29 reference int. if reference, no more data. if not reference,
//   length value of bytes to read .
func (e *Encoder) EncodeAmf3ByteArray(w io.Writer, val []byte, encodeMarker bool) (n int, err error) {
	if encodeMarker {
		if err = WriteMarker(w, AMF3_BYTEARRAY_MARKER); err != nil {
			return
		}
		n += 1
	}

	var m int

	length := uint32(len(val))
	u29 := (length << 1) | 1

	m, err = e.encodeAmf3Uint29(w, u29)
	if err != nil {
		return n, fmt.Errorf("amf3 encode: cannot encode u29 for bytearray: %s", err)
	}
	n += m

	m, err = w.Write(val)
	if err != nil {
		return n, fmt.Errorf("encode amf3: unable to encode bytearray value: %s", err)
	}
	n += m

	return
}

func (e *Encoder) encodeAmf3Utf8(w io.Writer, val string) (n int, err error) {
	length := uint32(len(val))
	u29 := uint32(length<<1) | 0x01

	var m int
	m, err = e.encodeAmf3Uint29(w, u29)
	if err != nil {
		return n, fmt.Errorf("amf3 encode: cannot encode u29 for string: %s", err)
	}
	n += m

	m, err = w.Write([]byte(val))
	if err != nil {
		return n, fmt.Errorf("encode amf3: unable to encode string value: %s", err)
	}
	n += m

	return
}

func (e *Encoder) encodeAmf3Uint29(w io.Writer, val uint32) (n int, err error) {
	if val <= 0x0000007F {
		err = WriteByte(w, byte(val))
		if err == nil {
			n += 1
		}
	} else if val <= 0x00003FFF {
		n, err = w.Write([]byte{byte(val>>7 | 0x80), byte(val & 0x7F)})
	} else if val <= 0x001FFFFF {
		n, err = w.Write([]byte{byte(val>>14 | 0x80), byte(val>>7&0x7F | 0x80), byte(val & 0x7F)})
	} else if val <= 0x1FFFFFFF {
		n, err = w.Write([]byte{byte(val>>22 | 0x80), byte(val>>15&0x7F | 0x80), byte(val>>8&0x7F | 0x80), byte(val)})
	} else {
		return n, fmt.Errorf("amf3 encode: cannot encode u29 with value %d (out of range)", val)
	}

	return
}
