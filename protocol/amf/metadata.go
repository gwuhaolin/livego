package amf

import (
	"bytes"
	"fmt"

	log "github.com/sirupsen/logrus"
)

const (
	ADD = 0x0
	DEL = 0x3
)

const (
	SetDataFrame string = "@setDataFrame"
	OnMetaData   string = "onMetaData"
)

var setFrameFrame []byte

func init() {
	b := bytes.NewBuffer(nil)
	encoder := &Encoder{}
	if _, err := encoder.Encode(b, SetDataFrame, AMF0); err != nil {
		log.Fatal(err)
	}
	setFrameFrame = b.Bytes()
}

func MetaDataReform(p []byte, flag uint8) ([]byte, error) {
	r := bytes.NewReader(p)
	decoder := &Decoder{}
	switch flag {
	case ADD:
		v, err := decoder.Decode(r, AMF0)
		if err != nil {
			return nil, err
		}
		switch v.(type) {
		case string:
			vv := v.(string)
			if vv != SetDataFrame {
				tmplen := len(setFrameFrame)
				b := make([]byte, tmplen+len(p))
				copy(b, setFrameFrame)
				copy(b[tmplen:], p)
				p = b
			}
		default:
			return nil, fmt.Errorf("setFrameFrame error")
		}
	case DEL:
		v, err := decoder.Decode(r, AMF0)
		if err != nil {
			return nil, err
		}
		switch v.(type) {
		case string:
			vv := v.(string)
			if vv == SetDataFrame {
				p = p[len(setFrameFrame):]
			}
		default:
			return nil, fmt.Errorf("metadata error")
		}
	default:
		return nil, fmt.Errorf("invalid flag:%d", flag)
	}
	return p, nil
}
