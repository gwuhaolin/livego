package cache

import (
	"bytes"
	"log"

	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/gwuhaolin/livego/av"
)

const (
	SetDataFrame string = "@setDataFrame"
	OnMetaData   string = "onMetaData"
)

var setFrameFrame []byte

func init() {
	b := bytes.NewBuffer(nil)
	encoder := &amf.Encoder{}
	if _, err := encoder.Encode(b, SetDataFrame, amf.AMF0); err != nil {
		log.Fatal(err)
	}
	setFrameFrame = b.Bytes()
}

type SpecialCache struct {
	full bool
	p    av.Packet
}

func NewSpecialCache() *SpecialCache {
	return &SpecialCache{}
}

func (self *SpecialCache) Write(p av.Packet) {
	self.p = p
	self.full = true
}

func (self *SpecialCache) Send(w av.WriteCloser) error {
	if !self.full {
		return nil
	}
	return w.Write(self.p)
}
