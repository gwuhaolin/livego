package amf

import (
	"encoding/binary"
	"fmt"
	"io"
)

// amf0 polymorphic router
func (d *Decoder) DecodeAmf0(r io.Reader) (interface{}, error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return nil, err
	}

	switch marker {
	case AMF0_NUMBER_MARKER:
		return d.DecodeAmf0Number(r, false)
	case AMF0_BOOLEAN_MARKER:
		return d.DecodeAmf0Boolean(r, false)
	case AMF0_STRING_MARKER:
		return d.DecodeAmf0String(r, false)
	case AMF0_OBJECT_MARKER:
		return d.DecodeAmf0Object(r, false)
	case AMF0_MOVIECLIP_MARKER:
		return nil, fmt.Errorf("decode amf0: unsupported type movieclip")
	case AMF0_NULL_MARKER:
		return d.DecodeAmf0Null(r, false)
	case AMF0_UNDEFINED_MARKER:
		return d.DecodeAmf0Undefined(r, false)
	case AMF0_REFERENCE_MARKER:
		return nil, fmt.Errorf("decode amf0: unsupported type reference")
	case AMF0_ECMA_ARRAY_MARKER:
		return d.DecodeAmf0EcmaArray(r, false)
	case AMF0_STRICT_ARRAY_MARKER:
		return d.DecodeAmf0StrictArray(r, false)
	case AMF0_DATE_MARKER:
		return d.DecodeAmf0Date(r, false)
	case AMF0_LONG_STRING_MARKER:
		return d.DecodeAmf0LongString(r, false)
	case AMF0_UNSUPPORTED_MARKER:
		return d.DecodeAmf0Unsupported(r, false)
	case AMF0_RECORDSET_MARKER:
		return nil, fmt.Errorf("decode amf0: unsupported type recordset")
	case AMF0_XML_DOCUMENT_MARKER:
		return d.DecodeAmf0XmlDocument(r, false)
	case AMF0_TYPED_OBJECT_MARKER:
		return d.DecodeAmf0TypedObject(r, false)
	case AMF0_ACMPLUS_OBJECT_MARKER:
		return d.DecodeAmf3(r)
	}

	return nil, fmt.Errorf("decode amf0: unsupported type %d", marker)
}

// marker: 1 byte 0x00
// format: 8 byte big endian float64
func (d *Decoder) DecodeAmf0Number(r io.Reader, decodeMarker bool) (result float64, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0_NUMBER_MARKER); err != nil {
		return
	}

	err = binary.Read(r, binary.BigEndian, &result)
	if err != nil {
		return float64(0), fmt.Errorf("amf0 decode: unable to read number: %s", err)
	}

	return
}

// marker: 1 byte 0x01
// format: 1 byte, 0x00 = false, 0x01 = true
func (d *Decoder) DecodeAmf0Boolean(r io.Reader, decodeMarker bool) (result bool, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0_BOOLEAN_MARKER); err != nil {
		return
	}

	var b byte
	if b, err = ReadByte(r); err != nil {
		return
	}

	if b == AMF0_BOOLEAN_FALSE {
		return false, nil
	} else if b == AMF0_BOOLEAN_TRUE {
		return true, nil
	}

	return false, fmt.Errorf("decode amf0: unexpected value %v for boolean", b)
}

// marker: 1 byte 0x02
// format:
// - 2 byte big endian uint16 header to determine size
// - n (size) byte utf8 string
func (d *Decoder) DecodeAmf0String(r io.Reader, decodeMarker bool) (result string, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0_STRING_MARKER); err != nil {
		return
	}

	var length uint16
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return "", fmt.Errorf("decode amf0: unable to decode string length: %s", err)
	}

	var bytes = make([]byte, length)
	if bytes, err = ReadBytes(r, int(length)); err != nil {
		return "", fmt.Errorf("decode amf0: unable to decode string value: %s", err)
	}

	return string(bytes), nil
}

// marker: 1 byte 0x03
// format:
// - loop encoded string followed by encoded value
// - terminated with empty string followed by 1 byte 0x09
func (d *Decoder) DecodeAmf0Object(r io.Reader, decodeMarker bool) (Object, error) {
	if err := AssertMarker(r, decodeMarker, AMF0_OBJECT_MARKER); err != nil {
		return nil, err
	}

	result := make(Object)
	d.refCache = append(d.refCache, result)

	for {
		key, err := d.DecodeAmf0String(r, false)
		if err != nil {
			return nil, err
		}

		if key == "" {
			if err = AssertMarker(r, true, AMF0_OBJECT_END_MARKER); err != nil {
				return nil, fmt.Errorf("decode amf0: expected object end marker: %s", err)
			}

			break
		}

		value, err := d.DecodeAmf0(r)
		if err != nil {
			return nil, fmt.Errorf("decode amf0: unable to decode object value: %s", err)
		}

		result[key] = value
	}

	return result, nil

}

// marker: 1 byte 0x05
// no additional data
func (d *Decoder) DecodeAmf0Null(r io.Reader, decodeMarker bool) (result interface{}, err error) {
	err = AssertMarker(r, decodeMarker, AMF0_NULL_MARKER)
	return
}

// marker: 1 byte 0x06
// no additional data
func (d *Decoder) DecodeAmf0Undefined(r io.Reader, decodeMarker bool) (result interface{}, err error) {
	err = AssertMarker(r, decodeMarker, AMF0_UNDEFINED_MARKER)
	return
}

// marker: 1 byte 0x07
// format: 2 byte big endian uint16
/*
func (d *Decoder) DecodeAmf0Reference(r io.Reader, decodeMarker bool) (interface{}, error) {
	if err := AssertMarker(r, decodeMarker, AMF0_REFERENCE_MARKER); err != nil {
		return nil, err
	}

	var err error
	var ref uint16

	err = binary.Read(r, binary.BigEndian, &ref)
	if err != nil {
		return nil, fmt.Errorf("decode amf0: unable to decode reference id: %s", err)
	}

	if int(ref) > len(d.refCache) {
		return nil, fmt.Errorf("decode amf0: bad reference %d (current length %d)", ref, len(d.refCache))
	}

	result := d.refCache[ref]

	return result, nil
}
*/

// marker: 1 byte 0x08
// format:
// - 4 byte big endian uint32 with length of associative array
// - normal object format:
//   - loop encoded string followed by encoded value
//   - terminated with empty string followed by 1 byte 0x09
func (d *Decoder) DecodeAmf0EcmaArray(r io.Reader, decodeMarker bool) (Object, error) {
	if err := AssertMarker(r, decodeMarker, AMF0_ECMA_ARRAY_MARKER); err != nil {
		return nil, err
	}

	var length uint32
	err := binary.Read(r, binary.BigEndian, &length)

	result, err := d.DecodeAmf0Object(r, false)
	if err != nil {
		return nil, fmt.Errorf("decode amf0: unable to decode ecma array object: %s", err)
	}

	return result, nil
}

// marker: 1 byte 0x0a
// format:
// - 4 byte big endian uint32 to determine length of associative array
// - n (length) encoded values
func (d *Decoder) DecodeAmf0StrictArray(r io.Reader, decodeMarker bool) (result Array, err error) {
	if err := AssertMarker(r, decodeMarker, AMF0_STRICT_ARRAY_MARKER); err != nil {
		return nil, err
	}

	var length uint32
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return nil, fmt.Errorf("decode amf0: unable to decode strict array length: %s", err)
	}

	d.refCache = append(d.refCache, result)

	for i := uint32(0); i < length; i++ {
		tmp, err := d.DecodeAmf0(r)
		if err != nil {
			return nil, fmt.Errorf("decode amf0: unable to decode strict array object: %s", err)
		}
		result = append(result, tmp)
	}

	return result, nil
}

// marker: 1 byte 0x0b
// format:
// - normal number format:
//   - 8 byte big endian float64
// - 2 byte unused
func (d *Decoder) DecodeAmf0Date(r io.Reader, decodeMarker bool) (result float64, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0_DATE_MARKER); err != nil {
		return
	}

	if result, err = d.DecodeAmf0Number(r, false); err != nil {
		return float64(0), fmt.Errorf("decode amf0: unable to decode float in date: %s", err)
	}

	if _, err = ReadBytes(r, 2); err != nil {
		return float64(0), fmt.Errorf("decode amf0: unable to read 2 trail bytes in date: %s", err)
	}

	return
}

// marker: 1 byte 0x0c
// format:
// - 4 byte big endian uint32 header to determine size
// - n (size) byte utf8 string
func (d *Decoder) DecodeAmf0LongString(r io.Reader, decodeMarker bool) (result string, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0_LONG_STRING_MARKER); err != nil {
		return
	}

	var length uint32
	err = binary.Read(r, binary.BigEndian, &length)
	if err != nil {
		return "", fmt.Errorf("decode amf0: unable to decode long string length: %s", err)
	}

	var bytes = make([]byte, length)
	if bytes, err = ReadBytes(r, int(length)); err != nil {
		return "", fmt.Errorf("decode amf0: unable to decode long string value: %s", err)
	}

	return string(bytes), nil
}

// marker: 1 byte 0x0d
// no additional data
func (d *Decoder) DecodeAmf0Unsupported(r io.Reader, decodeMarker bool) (result interface{}, err error) {
	err = AssertMarker(r, decodeMarker, AMF0_UNSUPPORTED_MARKER)
	return
}

// marker: 1 byte 0x0f
// format:
// - normal long string format
//   - 4 byte big endian uint32 header to determine size
//   - n (size) byte utf8 string
func (d *Decoder) DecodeAmf0XmlDocument(r io.Reader, decodeMarker bool) (result string, err error) {
	if err = AssertMarker(r, decodeMarker, AMF0_XML_DOCUMENT_MARKER); err != nil {
		return
	}

	return d.DecodeAmf0LongString(r, false)
}

// marker: 1 byte 0x10
// format:
// - normal string format:
//   - 2 byte big endian uint16 header to determine size
//   - n (size) byte utf8 string
// - normal object format:
//   - loop encoded string followed by encoded value
//   - terminated with empty string followed by 1 byte 0x09
func (d *Decoder) DecodeAmf0TypedObject(r io.Reader, decodeMarker bool) (TypedObject, error) {
	result := *new(TypedObject)

	err := AssertMarker(r, decodeMarker, AMF0_TYPED_OBJECT_MARKER)
	if err != nil {
		return result, err
	}

	d.refCache = append(d.refCache, result)

	result.Type, err = d.DecodeAmf0String(r, false)
	if err != nil {
		return result, fmt.Errorf("decode amf0: typed object unable to determine type: %s", err)
	}

	result.Object, err = d.DecodeAmf0Object(r, false)
	if err != nil {
		return result, fmt.Errorf("decode amf0: typed object unable to determine object: %s", err)
	}

	return result, nil
}
