package amf

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

// amf3 polymorphic router
func (d *Decoder) DecodeAmf3(r io.Reader) (interface{}, error) {
	marker, err := ReadMarker(r)
	if err != nil {
		return nil, err
	}

	switch marker {
	case AMF3_UNDEFINED_MARKER:
		return d.DecodeAmf3Undefined(r, false)
	case AMF3_NULL_MARKER:
		return d.DecodeAmf3Null(r, false)
	case AMF3_FALSE_MARKER:
		return d.DecodeAmf3False(r, false)
	case AMF3_TRUE_MARKER:
		return d.DecodeAmf3True(r, false)
	case AMF3_INTEGER_MARKER:
		return d.DecodeAmf3Integer(r, false)
	case AMF3_DOUBLE_MARKER:
		return d.DecodeAmf3Double(r, false)
	case AMF3_STRING_MARKER:
		return d.DecodeAmf3String(r, false)
	case AMF3_XMLDOC_MARKER:
		return d.DecodeAmf3Xml(r, false)
	case AMF3_DATE_MARKER:
		return d.DecodeAmf3Date(r, false)
	case AMF3_ARRAY_MARKER:
		return d.DecodeAmf3Array(r, false)
	case AMF3_OBJECT_MARKER:
		return d.DecodeAmf3Object(r, false)
	case AMF3_XMLSTRING_MARKER:
		return d.DecodeAmf3Xml(r, false)
	case AMF3_BYTEARRAY_MARKER:
		return d.DecodeAmf3ByteArray(r, false)
	}

	return nil, fmt.Errorf("decode amf3: unsupported type %d", marker)
}

// marker: 1 byte 0x00
// no additional data
func (d *Decoder) DecodeAmf3Undefined(r io.Reader, decodeMarker bool) (result interface{}, err error) {
	err = AssertMarker(r, decodeMarker, AMF3_UNDEFINED_MARKER)
	return
}

// marker: 1 byte 0x01
// no additional data
func (d *Decoder) DecodeAmf3Null(r io.Reader, decodeMarker bool) (result interface{}, err error) {
	err = AssertMarker(r, decodeMarker, AMF3_NULL_MARKER)
	return
}

// marker: 1 byte 0x02
// no additional data
func (d *Decoder) DecodeAmf3False(r io.Reader, decodeMarker bool) (result bool, err error) {
	err = AssertMarker(r, decodeMarker, AMF3_FALSE_MARKER)
	result = false
	return
}

// marker: 1 byte 0x03
// no additional data
func (d *Decoder) DecodeAmf3True(r io.Reader, decodeMarker bool) (result bool, err error) {
	err = AssertMarker(r, decodeMarker, AMF3_TRUE_MARKER)
	result = true
	return
}

// marker: 1 byte 0x04
func (d *Decoder) DecodeAmf3Integer(r io.Reader, decodeMarker bool) (result int32, err error) {
	if err = AssertMarker(r, decodeMarker, AMF3_INTEGER_MARKER); err != nil {
		return
	}

	var u29 uint32
	u29, err = d.decodeU29(r)
	if err != nil {
		return
	}

	result = int32(u29)
	if result > 0xfffffff {
		result = int32(u29 - 0x20000000)
	}

	return
}

// marker: 1 byte 0x05
func (d *Decoder) DecodeAmf3Double(r io.Reader, decodeMarker bool) (result float64, err error) {
	if err = AssertMarker(r, decodeMarker, AMF3_DOUBLE_MARKER); err != nil {
		return
	}

	err = binary.Read(r, binary.BigEndian, &result)
	if err != nil {
		return float64(0), fmt.Errorf("amf3 decode: unable to read double: %s", err)
	}

	return
}

// marker: 1 byte 0x06
// format:
// - u29 reference int. if reference, no more data. if not reference,
//   length value of bytes to read to complete string.
func (d *Decoder) DecodeAmf3String(r io.Reader, decodeMarker bool) (result string, err error) {
	if err = AssertMarker(r, decodeMarker, AMF3_STRING_MARKER); err != nil {
		return
	}

	var isRef bool
	var refVal uint32
	isRef, refVal, err = d.decodeReferenceInt(r)
	if err != nil {
		return "", fmt.Errorf("amf3 decode: unable to decode string reference and length: %s", err)
	}

	if isRef {
		result = d.stringRefs[refVal]
		return
	}

	buf := make([]byte, refVal)
	_, err = r.Read(buf)
	if err != nil {
		return "", fmt.Errorf("amf3 decode: unable to read string: %s", err)
	}

	result = string(buf)
	if result != "" {
		d.stringRefs = append(d.stringRefs, result)
	}

	return
}

// marker: 1 byte 0x08
// format:
// - u29 reference int, if reference, no more data
// - timestamp double
func (d *Decoder) DecodeAmf3Date(r io.Reader, decodeMarker bool) (result time.Time, err error) {
	if err = AssertMarker(r, decodeMarker, AMF3_DATE_MARKER); err != nil {
		return
	}

	var isRef bool
	var refVal uint32
	isRef, refVal, err = d.decodeReferenceInt(r)
	if err != nil {
		return result, fmt.Errorf("amf3 decode: unable to decode date reference and length: %s", err)
	}

	if isRef {
		res, ok := d.objectRefs[refVal].(time.Time)
		if ok != true {
			return result, fmt.Errorf("amf3 decode: unable to extract time from date object references")
		}

		return res, err
	}

	var u64 float64
	err = binary.Read(r, binary.BigEndian, &u64)
	if err != nil {
		return result, fmt.Errorf("amf3 decode: unable to read double: %s", err)
	}

	result = time.Unix(int64(u64/1000), 0).UTC()

	d.objectRefs = append(d.objectRefs, result)

	return
}

// marker: 1 byte 0x09
// format:
// - u29 reference int. if reference, no more data.
// - string representing associative array if present
// - n values (length of u29)
func (d *Decoder) DecodeAmf3Array(r io.Reader, decodeMarker bool) (result Array, err error) {
	if err = AssertMarker(r, decodeMarker, AMF3_ARRAY_MARKER); err != nil {
		return
	}

	var isRef bool
	var refVal uint32
	isRef, refVal, err = d.decodeReferenceInt(r)
	if err != nil {
		return result, fmt.Errorf("amf3 decode: unable to decode array reference and length: %s", err)
	}

	if isRef {
		objRefId := refVal >> 1

		res, ok := d.objectRefs[objRefId].(Array)
		if ok != true {
			return result, fmt.Errorf("amf3 decode: unable to extract array from object references")
		}

		return res, err
	}

	var key string
	key, err = d.DecodeAmf3String(r, false)
	if err != nil {
		return result, fmt.Errorf("amf3 decode: unable to read key for array: %s", err)
	}

	if key != "" {
		return result, fmt.Errorf("amf3 decode: array key is not empty, can't handle associative array")
	}

	for i := uint32(0); i < refVal; i++ {
		tmp, err := d.DecodeAmf3(r)
		if err != nil {
			return result, fmt.Errorf("amf3 decode: array element could not be decoded: %s", err)
		}
		result = append(result, tmp)
	}

	d.objectRefs = append(d.objectRefs, result)

	return
}

// marker: 1 byte 0x09
// format: oh dear god
func (d *Decoder) DecodeAmf3Object(r io.Reader, decodeMarker bool) (result interface{}, err error) {
	if err = AssertMarker(r, decodeMarker, AMF3_OBJECT_MARKER); err != nil {
		return nil, err
	}

	// decode the initial u29
	isRef, refVal, err := d.decodeReferenceInt(r)
	if err != nil {
		return nil, fmt.Errorf("amf3 decode: unable to decode object reference and length: %s", err)
	}

	// if this is a object reference only, grab it and return it
	if isRef {
		objRefId := refVal >> 1

		return d.objectRefs[objRefId], nil
	}

	// each type has traits that are cached, if the peer sent a reference
	// then we'll need to look it up and use it.
	var trait Trait

	traitIsRef := (refVal & 0x01) == 0

	if traitIsRef {
		traitRef := refVal >> 1
		trait = d.traitRefs[traitRef]

	} else {
		// build a new trait from what's left of the given u29
		trait = *NewTrait()
		trait.Externalizable = (refVal & 0x02) != 0
		trait.Dynamic = (refVal & 0x04) != 0

		var cls string
		cls, err = d.DecodeAmf3String(r, false)
		if err != nil {
			return result, fmt.Errorf("amf3 decode: unable to read trait type for object: %s", err)
		}
		trait.Type = cls

		// traits have property keys, encoded as amf3 strings
		propLength := refVal >> 3
		for i := uint32(0); i < propLength; i++ {
			tmp, err := d.DecodeAmf3String(r, false)
			if err != nil {
				return result, fmt.Errorf("amf3 decode: unable to read trait property for object: %s", err)
			}
			trait.Properties = append(trait.Properties, tmp)
		}

		d.traitRefs = append(d.traitRefs, trait)
	}

	d.objectRefs = append(d.objectRefs, result)

	// objects can be externalizable, meaning that the system has no concrete understanding of
	// their properties or how they are encoded. in that case, we need to find and delegate behavior
	// to the right object.
	if trait.Externalizable {
		switch trait.Type {
		case "DSA": // AsyncMessageExt
			result, err = d.decodeAsyncMessageExt(r)
			if err != nil {
				return result, fmt.Errorf("amf3 decode: unable to decode dsa: %s", err)
			}
		case "DSK": // AcknowledgeMessageExt
			result, err = d.decodeAcknowledgeMessageExt(r)
			if err != nil {
				return result, fmt.Errorf("amf3 decode: unable to decode dsk: %s", err)
			}
		case "flex.messaging.io.ArrayCollection":
			result, err = d.decodeArrayCollection(r)
			if err != nil {
				return result, fmt.Errorf("amf3 decode: unable to decode ac: %s", err)
			}

			// store an extra reference to array collection container
			d.objectRefs = append(d.objectRefs, result)

		default:
			fn, ok := d.externalHandlers[trait.Type]
			if ok {
				result, err = fn(d, r)
				if err != nil {
					return result, fmt.Errorf("amf3 decode: unable to call external decoder for type %s: %s", trait.Type, err)
				}
			} else {
				return result, fmt.Errorf("amf3 decode: unable to decode external type %s, no handler", trait.Type)
			}
		}

		return result, err
	}

	var key string
	var val interface{}
	var obj Object

	obj = make(Object)

	// non-externalizable objects have property keys in traits, iterate through them
	// and add the read values to the object
	for _, key = range trait.Properties {
		val, err = d.DecodeAmf3(r)
		if err != nil {
			return result, fmt.Errorf("amf3 decode: unable to decode object property: %s", err)
		}

		obj[key] = val
	}

	// if an object is dynamic, it can have extra key/value data at the end. in this case,
	// read keys until we get an empty one.
	if trait.Dynamic {
		for {
			key, err = d.DecodeAmf3String(r, false)
			if err != nil {
				return result, fmt.Errorf("amf3 decode: unable to decode dynamic key: %s", err)
			}
			if key == "" {
				break
			}
			val, err = d.DecodeAmf3(r)
			if err != nil {
				return result, fmt.Errorf("amf3 decode: unable to decode dynamic value: %s", err)
			}

			obj[key] = val
		}
	}

	result = obj

	return
}

// marker: 1 byte 0x07 or 0x0b
// format:
// - u29 reference int. if reference, no more data. if not reference,
//   length value of bytes to read to complete string.
func (d *Decoder) DecodeAmf3Xml(r io.Reader, decodeMarker bool) (result string, err error) {
	if decodeMarker {
		var marker byte
		marker, err = ReadMarker(r)
		if err != nil {
			return "", err
		}

		if (marker != AMF3_XMLDOC_MARKER) && (marker != AMF3_XMLSTRING_MARKER) {
			return "", fmt.Errorf("decode assert marker failed: expected %v or %v, got %v", AMF3_XMLDOC_MARKER, AMF3_XMLSTRING_MARKER, marker)
		}
	}

	var isRef bool
	var refVal uint32
	isRef, refVal, err = d.decodeReferenceInt(r)
	if err != nil {
		return "", fmt.Errorf("amf3 decode: unable to decode xml reference and length: %s", err)
	}

	if isRef {
		var ok bool
		buf := d.objectRefs[refVal]
		result, ok = buf.(string)
		if ok != true {
			return "", fmt.Errorf("amf3 decode: cannot coerce object reference into xml string")
		}

		return
	}

	buf := make([]byte, refVal)
	_, err = r.Read(buf)
	if err != nil {
		return "", fmt.Errorf("amf3 decode: unable to read xml string: %s", err)
	}

	result = string(buf)

	if result != "" {
		d.objectRefs = append(d.objectRefs, result)
	}

	return
}

// marker: 1 byte 0x0c
// format:
// - u29 reference int. if reference, no more data. if not reference,
//   length value of bytes to read.
func (d *Decoder) DecodeAmf3ByteArray(r io.Reader, decodeMarker bool) (result []byte, err error) {
	if err = AssertMarker(r, decodeMarker, AMF3_BYTEARRAY_MARKER); err != nil {
		return
	}

	var isRef bool
	var refVal uint32
	isRef, refVal, err = d.decodeReferenceInt(r)
	if err != nil {
		return result, fmt.Errorf("amf3 decode: unable to decode byte array reference and length: %s", err)
	}

	if isRef {
		var ok bool
		result, ok = d.objectRefs[refVal].([]byte)
		if ok != true {
			return result, fmt.Errorf("amf3 decode: unable to convert object ref to bytes")
		}

		return
	}

	result = make([]byte, refVal)
	_, err = r.Read(result)
	if err != nil {
		return result, fmt.Errorf("amf3 decode: unable to read bytearray: %s", err)
	}

	d.objectRefs = append(d.objectRefs, result)

	return
}

func (d *Decoder) decodeU29(r io.Reader) (result uint32, err error) {
	var b byte

	for i := 0; i < 3; i++ {
		b, err = ReadByte(r)
		if err != nil {
			return
		}
		result = (result << 7) + uint32(b&0x7F)
		if (b & 0x80) == 0 {
			return
		}
	}

	b, err = ReadByte(r)
	if err != nil {
		return
	}

	result = ((result << 8) + uint32(b))

	return
}

func (d *Decoder) decodeReferenceInt(r io.Reader) (isRef bool, refVal uint32, err error) {
	u29, err := d.decodeU29(r)
	if err != nil {
		return false, 0, fmt.Errorf("amf3 decode: unable to decode reference int: %s", err)
	}

	isRef = u29&0x01 == 0
	refVal = u29 >> 1

	return
}
