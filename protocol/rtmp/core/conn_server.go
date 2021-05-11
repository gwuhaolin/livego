package core

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/amf"

	log "github.com/sirupsen/logrus"
)

var (
	publishLive   = "live"
	publishRecord = "record"
	publishAppend = "append"
)

var (
	ErrReq = fmt.Errorf("req error")
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

func (connServer *ConnServer) writeMsg(csid, streamID uint32, args ...interface{}) error {
	connServer.bytesw.Reset()
	for _, v := range args {
		if _, err := connServer.encoder.Encode(connServer.bytesw, v, amf.AMF0); err != nil {
			return err
		}
	}
	msg := connServer.bytesw.Bytes()
	c := ChunkStream{
		Format:    0,
		CSID:      csid,
		Timestamp: 0,
		TypeID:    20,
		StreamID:  streamID,
		Length:    uint32(len(msg)),
		Data:      msg,
	}
	connServer.conn.Write(&c)
	return connServer.conn.Flush()
}

func (connServer *ConnServer) connect(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			id := int(v.(float64))
			if id != 1 {
				return ErrReq
			}
			connServer.transactionID = id
		case amf.Object:
			obimap := v.(amf.Object)
			if app, ok := obimap["app"]; ok {
				connServer.ConnInfo.App = app.(string)
			}
			if flashVer, ok := obimap["flashVer"]; ok {
				connServer.ConnInfo.Flashver = flashVer.(string)
			}
			if tcurl, ok := obimap["tcUrl"]; ok {
				connServer.ConnInfo.TcUrl = tcurl.(string)
			}
			if encoding, ok := obimap["objectEncoding"]; ok {
				connServer.ConnInfo.ObjectEncoding = int(encoding.(float64))
			}
		}
	}
	return nil
}

func (connServer *ConnServer) releaseStream(vs []interface{}) error {
	return nil
}

func (connServer *ConnServer) fcPublish(vs []interface{}) error {
	return nil
}

func (connServer *ConnServer) connectResp(cur *ChunkStream) error {
	c := connServer.conn.NewWindowAckSize(2500000)
	connServer.conn.Write(&c)
	c = connServer.conn.NewSetPeerBandwidth(2500000)
	connServer.conn.Write(&c)
	c = connServer.conn.NewSetChunkSize(uint32(1024))
	connServer.conn.Write(&c)

	resp := make(amf.Object)
	resp["fmsVer"] = "FMS/3,0,1,123"
	resp["capabilities"] = 31

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetConnection.Connect.Success"
	event["description"] = "Connection succeeded."
	event["objectEncoding"] = connServer.ConnInfo.ObjectEncoding
	return connServer.writeMsg(cur.CSID, cur.StreamID, "_result", connServer.transactionID, resp, event)
}

func (connServer *ConnServer) createStream(vs []interface{}) error {
	for _, v := range vs {
		switch v.(type) {
		case string:
		case float64:
			connServer.transactionID = int(v.(float64))
		case amf.Object:
		}
	}
	return nil
}

func (connServer *ConnServer) createStreamResp(cur *ChunkStream) error {
	return connServer.writeMsg(cur.CSID, cur.StreamID, "_result", connServer.transactionID, nil, connServer.streamID)
}

func (connServer *ConnServer) publishOrPlay(vs []interface{}) error {
	for k, v := range vs {
		switch v.(type) {
		case string:
			if k == 2 {
				connServer.PublishInfo.Name = v.(string)
			} else if k == 3 {
				connServer.PublishInfo.Type = v.(string)
			}
		case float64:
			id := int(v.(float64))
			connServer.transactionID = id
		case amf.Object:
		}
	}

	return nil
}

func (connServer *ConnServer) publishResp(cur *ChunkStream) error {
	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Publish.Start"
	event["description"] = "Start publishing."
	return connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event)
}

func (connServer *ConnServer) playResp(cur *ChunkStream) error {
	connServer.conn.SetRecorded()
	connServer.conn.SetBegin()

	event := make(amf.Object)
	event["level"] = "status"
	event["code"] = "NetStream.Play.Reset"
	event["description"] = "Playing and resetting stream."
	if err := connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.Start"
	event["description"] = "Started playing stream."
	if err := connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Data.Start"
	event["description"] = "Started playing stream."
	if err := connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}

	event["level"] = "status"
	event["code"] = "NetStream.Play.PublishNotify"
	event["description"] = "Started playing notify."
	if err := connServer.writeMsg(cur.CSID, cur.StreamID, "onStatus", 0, nil, event); err != nil {
		return err
	}
	return connServer.conn.Flush()
}

func (connServer *ConnServer) handleCmdMsg(c *ChunkStream) error {
	amfType := amf.AMF0
	if c.TypeID == 17 {
		c.Data = c.Data[1:]
	}
	r := bytes.NewReader(c.Data)
	vs, err := connServer.decoder.DecodeBatch(r, amf.Version(amfType))
	if err != nil && err != io.EOF {
		return err
	}
	// log.Debugf("rtmp req: %#v", vs)
	switch vs[0].(type) {
	case string:
		switch vs[0].(string) {
		case cmdConnect:
			if err = connServer.connect(vs[1:]); err != nil {
				return err
			}
			if err = connServer.connectResp(c); err != nil {
				return err
			}
		case cmdCreateStream:
			if err = connServer.createStream(vs[1:]); err != nil {
				return err
			}
			if err = connServer.createStreamResp(c); err != nil {
				return err
			}
		case cmdPublish:
			if err = connServer.publishOrPlay(vs[1:]); err != nil {
				return err
			}
			if err = connServer.publishResp(c); err != nil {
				return err
			}
			connServer.done = true
			connServer.isPublisher = true
			log.Debug("handle publish req done")
		case cmdPlay:
			if err = connServer.publishOrPlay(vs[1:]); err != nil {
				return err
			}
			if err = connServer.playResp(c); err != nil {
				return err
			}
			connServer.done = true
			connServer.isPublisher = false
			log.Debug("handle play req done")
		case cmdFcpublish:
			connServer.fcPublish(vs)
		case cmdReleaseStream:
			connServer.releaseStream(vs)
		case cmdFCUnpublish:
		case cmdDeleteStream:
		default:
			log.Debug("no support command=", vs[0].(string))
		}
	}

	return nil
}

func (connServer *ConnServer) ReadMsg() error {
	var c ChunkStream
	for {
		if err := connServer.conn.Read(&c); err != nil {
			return err
		}
		switch c.TypeID {
		case 20, 17:
			if err := connServer.handleCmdMsg(&c); err != nil {
				return err
			}
		}
		if connServer.done {
			break
		}
	}
	return nil
}

func (connServer *ConnServer) IsPublisher() bool {
	return connServer.isPublisher
}

func (connServer *ConnServer) Write(c ChunkStream) error {
	if c.TypeID == av.TAG_SCRIPTDATAAMF0 ||
		c.TypeID == av.TAG_SCRIPTDATAAMF3 {
		var err error
		if c.Data, err = amf.MetaDataReform(c.Data, amf.DEL); err != nil {
			return err
		}
		c.Length = uint32(len(c.Data))
	}
	return connServer.conn.Write(&c)
}

func (connServer *ConnServer) Flush() error {
	return connServer.conn.Flush()
}

func (connServer *ConnServer) Read(c *ChunkStream) (err error) {
	return connServer.conn.Read(c)
}

func (connServer *ConnServer) GetInfo() (app string, name string, url string) {
	app = connServer.ConnInfo.App
	name = connServer.PublishInfo.Name
	url = connServer.ConnInfo.TcUrl + "/" + connServer.PublishInfo.Name
	return
}

func (connServer *ConnServer) Close(err error) {
	connServer.conn.Close()
}
