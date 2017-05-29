package core

import (
	"bytes"
	"errors"
	"io"

	"github.com/gwuhaolin/livego/protocol/amf"
	"github.com/gwuhaolin/livego/av"
	"log"
)

var (
	publishLive   = "live"
	publishRecord = "record"
	publishAppend = "append"
)

var (
	ErrReq = errors.New("req error")
)

var (
	cmdConnect       = "connect"
	cmdFcpublish     = "FCPublish"
	cmdReleaseStream = "releaseStream"
	cmdCreateStream  = "createStream"
	cmdPublish       = "publish"
	cmdFCUnpublish   = "FCUnpublish"
	cmdDeleteStream  = "deleteStream"
	cmdPlay          = "play"
)

type ConnectInfo struct {
	App            string `amf:"app" json:"app"`
	Flashver       string `amf:"flashVer" json:"flashVer"`
	SwfUrl         string `amf:"swfUrl" json:"swfUrl"`
	TcUrl          string `amf:"tcUrl" json:"tcUrl"`
	Fpad           bool   `amf:"fpad" json:"fpad"`
	AudioCodecs    int    `amf:"audioCodecs" json:"audioCodecs"`
	VideoCodecs    int    `amf:"videoCodecs" json:"videoCodecs"`
	VideoFunction  int    `amf:"videoFunction" json:"videoFunction"`
	PageUrl        string `amf:"pageUrl" json:"pageUrl"`
	ObjectEncoding int    `amf:"objectEncoding" json:"objectEncoding"`
}

type ConnectResp struct {
	FMSVer       string `amf:"fmsVer"`
	Capabilities int    `amf:"capabilities"`
}

type ConnectEvent struct {
	Level          string `amf:"level"`
	Code           string `amf:"code"`
	Description    string `amf:"description"`
	ObjectEncoding int    `amf:"objectEncoding"`
}

type PublishInfo struct {
	Name string
	Type string
}

type ConnServer struct {
	done          bool
	streamID      int
	isPublisher   bool
	conn          *Conn
	transactionID int
	ConnInfo      ConnectInfo
	PublishInfo   PublishInfo
	decoder       *amf.Decoder
	encoder       *amf.Encoder
	bytesw        *bytes.Buffer
}

func NewConnServer(conn *Conn) *ConnServer {
	return &ConnServer{
		conn:     conn,
		streamID: 1,
		bytesw:   bytes.NewBuffer(nil),
		decoder:  &amf.Decoder{},
		encoder:  &amf.Encoder{},
	}
}

func (self *ConnServer) writeMsg(csid, streamID uint32, args ...interface{}) error {
	self.bytesw.Reset()
	for _, v := range args {
		if _, err := self.encoder.Encode(self.bytesw, v, amf.AMF0); err != nil {
			return err
		}
	}
	msg := self.bytesw.Bytes()
	c := ChunkStream{
		Format:    0,
		CSID:      csid,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  streamID,
		Length:    uint32(len(msg)),
		Data:      msg,
	}
	self.conn.Write(&c)
	return self.conn.Flush()
}

func (self *ConnServer) connect(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			id := int(v.(float64))
			if id != 1 {
				return ErrReq
			}
			self.transactionID = id
		case amf.Object:
			obimap := v.(amf.Object)
			if app, ok := obimap["app"]; ok {
				self.ConnInfo.App = app.(string)
			}
			if flashVer, ok := obimap["flashVer"]; ok {
				self.ConnInfo.Flashver = flashVer.(string)
			}
			if tcurl, ok := obimap["tcUrl"]; ok {
				self.ConnInfo.TcUrl = tcurl.(string)
			}
			if encoding, ok := obimap["objectEncoding"]; ok {
				self.ConnInfo.ObjectEncoding = int(encoding.(float64))
			}
		}
	}
	return nil
}

func (self *ConnServer) releaseStream(vs []interface{}) error {
	return nil
}

func (self *ConnServer) fcPublish(vs []interface{}) error {
	return nil
}

func (self *ConnServer) connectResp(cur *ChunkStream) error {
	c := self.conn.NewWindowAckSize(2500000)
	self.conn.Write(&c)
	c = self.conn.NewSetPeerBandwidth(2500000)
	self.conn.Write(&c)
	c = self.conn.NewSetChunkSize(uint32(1024))
	self.conn.Write(&c)

	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = self.ConnInfo.ObjectEncoding
	return self.writeMsg(cur.CSID, cur.StreamID, "_result", self.transactionID, resp, event)
}

func (self *ConnServer) createStream(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			self.transactionID = int(v.(float64))
		case amf.Object:
		}
	}
	return nil
}

func (self *ConnServer) createStreamResp(cur *ChunkStream) error {
	return self.writeMsg(cur.CSID, cur.StreamID, "_result", self.transactionID, nil, self.streamID)
}

func (self *ConnServer) publishOrPlay(vs []interface{}) error {
	for k, v := range vs {
		switch v.(type) {
		case string:
			if k == 2 {
				self.PublishInfo.Name = v.(string)
			} else if k == 3 {
				self.PublishInfo.Type = v.(string)
			}
		case float64:
			id := int(v.(float64))
			self.transactionID = id
		case amf.Object:
		}
	}

	return nil
}

func (self *ConnServer) publishResp(cur *ChunkStream) error {
	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Publish.Start"
	event["description"] = "Start publising."
	return self.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event)
}

func (self *ConnServer) playResp(cur *ChunkStream) error {
	self.conn.SetRecorded()
	self.conn.SetBegin()

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Play.Reset"
	event["description"] = "Playing and resetting stream."
	if err := self.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.Start"
	event["description"] = "Started playing stream."
	if err := self.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Data.Start"
	event["description"] = "Started playing stream."
	if err := self.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.PublishNotify"
	event["description"] = "Started playing notify."
	if err := self.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}
	return self.conn.Flush()
}

func (self *ConnServer) handleCmdMsg(c *ChunkStream) error {
	amfType := amf.AMF0
	if c.TypeID == 17 {
		c.Data = c.Data[1:]
	}
	r := bytes.NewReader(c.Data)
	vs, err := self.decoder.DecodeBatch(r, amf.Version(amfType))
	if err != nil && err != io.EOF {
		return err
	}
	// log.Printf("rtmp req: %#v", vs)
	switch vs[0].(type) {
	case string:
		switch vs[0].(string) {
		case cmdConnect:
			if err = self.connect(vs[1:]); err != nil {
				return err
			}
			if err = self.connectResp(c); err != nil {
				return err
			}
		case cmdCreateStream:
			if err = self.createStream(vs[1:]); err != nil {
				return err
			}
			if err = self.createStreamResp(c); err != nil {
				return err
			}
		case cmdPublish:
			if err = self.publishOrPlay(vs[1:]); err != nil {
				return err
			}
			if err = self.publishResp(c); err != nil {
				return err
			}
			self.done = true
			self.isPublisher = true
			log.Println("handle publish req done")
		case cmdPlay:
			if err = self.publishOrPlay(vs[1:]); err != nil {
				return err
			}
			if err = self.playResp(c); err != nil {
				return err
			}
			self.done = true
			self.isPublisher = false
			log.Println("handle play req done")
		case cmdFcpublish:
			self.fcPublish(vs)
		case cmdReleaseStream:
			self.releaseStream(vs)
		case cmdFCUnpublish:
		case cmdDeleteStream:
		default:
			log.Println("no support command=", vs[0].(string))
		}
	}

	return nil
}

func (self *ConnServer) ReadMsg() error {
	var c ChunkStream
	for {
		if err := self.conn.Read(&c); err != nil {
			return err
		}
		switch c.TypeID {
		case 20, 17:
			if err := self.handleCmdMsg(&c); err != nil {
				return err
			}
		}
		if self.done {
			break
		}
	}
	return nil
}

func (self *ConnServer) IsPublisher() bool {
	return self.isPublisher
}

func (self *ConnServer) Write(c ChunkStream) error {
	if c.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		c.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if c.Data, err = amf.MetaDataReform(c.Data, amf.DEL); err != nil {
			return err
		}
		c.Length = uint32(len(c.Data))
	}
	return self.conn.Write(&c)
}

func (self *ConnServer) Read(c *ChunkStream) (err error) {
	return self.conn.Read(c)
}

func (self *ConnServer) GetInfo() (app string, name string, url string) {
	app = self.ConnInfo.App
	name = self.PublishInfo.Name
	url = self.ConnInfo.TcUrl + "/" + self.PublishInfo.Name
	return
}

func (self *ConnServer) Close(err error) {
	self.conn.Close()
}
