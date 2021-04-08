package flv

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/configure"
	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/gwuhaolin/livego/utils/pio"
	"github.com/gwuhaolin/livego/utils/uid"

	log "github.com/sirupsen/logrus"
)

var (
	flvHeader = []byte{0x46, 0x4c, 0x56, 0x01, 0x05, 0x00, 0x00, 0x00, 0x09}
)

/*
func NewFlv(handler av.Handler, info av.Info) {
	patths := strings.SplitN(info.Key, "/", 2)

	if len(patths) != 2 {
		log.Warning("invalid info")
		return
	}

	w, err := os.OpenFile(*flvFile, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Error("open file error: ", err)
	}

	writer := NewFLVWriter(patths[0], patths[1], info.URL, w)

	handler.HandleWriter(writer)

	writer.Wait()
	// close flv file
	log.Debug("close flv file")
	writer.ctx.Close()
}
*/

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
	closedWriter    bool
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
	if writer.closedWriter {
		return
	}
	writer.closedWriter = true
	writer.ctx.Close()
	close(writer.closed)
}

func (writer *FLVWriter) Info() (ret av.Info) {
	ret.UID = writer.Uid
	ret.URL = writer.url
	ret.Key = writer.app + "/" + writer.title
	return
}

type FlvDvr struct{}

func (f *FlvDvr) GetWriter(info av.Info) av.WriteCloser {
	paths := strings.SplitN(info.Key, "/", 2)
	if len(paths) != 2 {
		log.Warning("invalid info")
		return nil
	}

	flvDir := configure.Config.GetString("flv_dir")

	err := os.MkdirAll(path.Join(flvDir, paths[0]), 0755)
	if err != nil {
		log.Error("mkdir error: ", err)
		return nil
	}

	fileName := fmt.Sprintf("%s_%d.%s", path.Join(flvDir, info.Key), time.Now().Unix(), "flv")
	log.Debug("flv dvr save stream to: ", fileName)
	w, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		log.Error("open file error: ", err)
		return nil
	}

	writer := NewFLVWriter(paths[0], paths[1], info.URL, w)
	log.Debug("new flv dvr: ", writer.Info())
	return writer
}
