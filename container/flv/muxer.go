package flv

import (
	"flag"
	"log"
	"os"
	"strings"
	"time"

	"livego/av"
	"livego/protocol/amf"
	"livego/utils/pio"
	"livego/utils/uid"
)

var (
	flvHeader = []byte{0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09}
	flvFile   = flag.String("filFile", "./out.flv", "output flv file name")
)

func NewFlv(handler av.Handler, info av.Info) {
	patths := strings.SplitN(info.Key, "/", 2)

	if len(patths) != 2 {
		log.Println("invalid info")
		return
	}

	w, err := os.OpenFile(*flvFile, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Println("open file error: ", err)
	}

	writer := NewFLVWriter(patths[0], patths[1], info.URL, w)

	handler.HandleWriter(writer)

	writer.Wait()
	// close flv file
	log.Println("close flv file")
	writer.ctx.Close()
}

const (
	headerLen = 11
)

type FLVWriter struct {
	Uid string
	av.RWBaser
	app, title, url string
	buf             []byte
	closed          chan struct{}
	ctx             *os.File
}

func NewFLVWriter(app, title, url string, ctx *os.File) *FLVWriter {
	ret := &FLVWriter{
		Uid:     uid.NewId(),
		app:     app,
		title:   title,
		url:     url,
		ctx:     ctx,
		RWBaser: av.NewRWBaser(time.Second * 10),
		closed:  make(chan struct{}),
		buf:     make([]byte, headerLen),
	}

	ret.ctx.Write(flvHeader)
	pio.PutI32BE(ret.buf[:4], 0)
	ret.ctx.Write(ret.buf[:4])

	return ret
}

func (writer *FLVWriter) Write(p *av.Packet) error {
	writer.RWBaser.SetPreTime()
	h := writer.buf[:headerLen]
	typeID := av.TAG_VIDEO
	if !p.IsVideo {
		if p.IsMetadata {
			var err error
			typeID = av.TAG_SCRIPTDATAAMF0
			p.Data, err = amf.MetaDataReform(p.Data, amf.DEL)
			if err != nil {
				return err
			}
		} else {
			typeID = av.TAG_AUDIO
		}
	}
	dataLen := len(p.Data)
	timestamp := p.TimeStamp
	timestamp += writer.BaseTimeStamp()
	writer.RWBaser.RecTimeStamp(timestamp, uint32(typeID))

	preDataLen := dataLen + headerLen
	timestampbase := timestamp & 0xffffff
	timestampExt := timestamp >> 24 & 0xff

	pio.PutU8(h[0:1], uint8(typeID))
	pio.PutI24BE(h[1:4], int32(dataLen))
	pio.PutI24BE(h[4:7], int32(timestampbase))
	pio.PutU8(h[7:8], uint8(timestampExt))

	if _, err := writer.ctx.Write(h); err != nil {
		return err
	}

	if _, err := writer.ctx.Write(p.Data); err != nil {
		return err
	}

	pio.PutI32BE(h[:4], int32(preDataLen))
	if _, err := writer.ctx.Write(h[:4]); err != nil {
		return err
	}

	return nil
}

func (writer *FLVWriter) Wait() {
	select {
	case <-writer.closed:
		return
	}
}

func (writer *FLVWriter) Close(error) {
	writer.ctx.Close()
	close(writer.closed)
}

func (writer *FLVWriter) Info() (ret av.Info) {
	ret.UID = writer.Uid
	ret.URL = writer.url
	ret.Key = writer.app + "/" + writer.title
	return
}
