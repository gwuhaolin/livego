package rtmp

import (
	"fmt"
	"strings"
	"time"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/rtmp/cache"
	"github.com/gwuhaolin/livego/protocol/rtmp/rtmprelay"

	cmap "github.com/orcaman/concurrent-map"
	log "github.com/sirupsen/logrus"
)

var (
	EmptyID = ""
)

type RtmpStream struct {
	streams cmap.ConcurrentMap //key
}

func NewRtmpStream() *RtmpStream {
	ret := &RtmpStream{
		streams: cmap.New(),
	}
	go ret.CheckAlive()
	return ret
}

func (rs *RtmpStream) HandleReader(r av.ReadCloser) {
	info := r.Info()
	log.Debugf("HandleReader: info[%v]", info)

	var stream *Stream
	i, ok := rs.streams.Get(info.Key)
	if stream, ok = i.(*Stream); ok {
		stream.TransStop()
		id := stream.ID()
		if id != EmptyID && id != info.UID {
			ns := NewStream()
			stream.Copy(ns)
			stream = ns
			rs.streams.Set(info.Key, ns)
		}
	} else {
		stream = NewStream()
		rs.streams.Set(info.Key, stream)
		stream.info = info
	}

	stream.AddReader(r)
}

func (rs *RtmpStream) HandleWriter(w av.WriteCloser) {
	info := w.Info()
	log.Debugf("HandleWriter: info[%v]", info)

	var s *Stream
	ok := rs.streams.Has(info.Key)
	if !ok {
		s = NewStream()
		rs.streams.Set(info.Key, s)
		s.info = info
	} else {
		item, ok := rs.streams.Get(info.Key)
		if ok {
			s = item.(*Stream)
			s.AddWriter(w)
		}
	}
}

func (rs *RtmpStream) GetStreams() cmap.ConcurrentMap {
	return rs.streams
}

func (rs *RtmpStream) CheckAlive() {
	for {
		<-time.After(5 * time.Second)
		for item := range rs.streams.IterBuffered() {
			v := item.Val.(*Stream)
			if v.CheckAlive() == 0 {
				rs.streams.Remove(item.Key)
			}
		}
	}
}

type Stream struct {
	isStart bool
	cache   *cache.Cache
	r       av.ReadCloser
	ws      cmap.ConcurrentMap
	info    av.Info
}

type PackWriterCloser struct {
	init bool
	w    av.WriteCloser
}

func (p *PackWriterCloser) GetWriter() av.WriteCloser {
	return p.w
}

func NewStream() *Stream {
	return &Stream{
		cache: cache.NewCache(),
		ws:    cmap.New(),
	}
}

func (s *Stream) ID() string {
	if s.r != nil {
		return s.r.Info().UID
	}
	return EmptyID
}

func (s *Stream) GetReader() av.ReadCloser {
	return s.r
}

func (s *Stream) GetWs() cmap.ConcurrentMap {
	return s.ws
}

func (s *Stream) Copy(dst *Stream) {
	for item := range s.ws.IterBuffered() {
		v := item.Val.(*PackWriterCloser)
		s.ws.Remove(item.Key)
		v.w.CalcBaseTimestamp()
		dst.AddWriter(v.w)
	}
}

func (s *Stream) AddReader(r av.ReadCloser) {
	s.r = r
	go s.TransStart()
}

func (s *Stream) AddWriter(w av.WriteCloser) {
	info := w.Info()
	pw := &PackWriterCloser{w: w}
	s.ws.Set(info.UID, pw)
}

/*检测本application下是否配置static_push,
如果配置, 启动push远端的连接*/
func (s *Stream) StartStaticPush() {
	key := s.info.Key

	dscr := strings.Split(key, "/")
	if len(dscr) < 1 {
		return
	}

	index := strings.Index(key, "/")
	if index < 0 {
		return
	}

	streamname := key[index+1:]
	appname := dscr[0]

	log.Debugf("StartStaticPush: current streamname=%s， appname=%s", streamname, appname)
	pushurllist, err := rtmprelay.GetStaticPushList(appname)
	if err != nil || len(pushurllist) < 1 {
		log.Debugf("StartStaticPush: GetStaticPushList error=%v", err)
		return
	}

	for _, pushurl := range pushurllist {
		pushurl := pushurl + "/" + streamname
		log.Debugf("StartStaticPush: static pushurl=%s", pushurl)

		staticpushObj := rtmprelay.GetAndCreateStaticPushObject(pushurl)
		if staticpushObj != nil {
			if err := staticpushObj.Start(); err != nil {
				log.Debugf("StartStaticPush: staticpushObj.Start %s error=%v", pushurl, err)
			} else {
				log.Debugf("StartStaticPush: staticpushObj.Start %s ok", pushurl)
			}
		} else {
			log.Debugf("StartStaticPush GetStaticPushObject %s error", pushurl)
		}
	}
}

func (s *Stream) StopStaticPush() {
	key := s.info.Key

	log.Debugf("StopStaticPush......%s", key)
	dscr := strings.Split(key, "/")
	if len(dscr) < 1 {
		return
	}

	index := strings.Index(key, "/")
	if index < 0 {
		return
	}

	streamname := key[index+1:]
	appname := dscr[0]

	log.Debugf("StopStaticPush: current streamname=%s， appname=%s", streamname, appname)
	pushurllist, err := rtmprelay.GetStaticPushList(appname)
	if err != nil || len(pushurllist) < 1 {
		log.Debugf("StopStaticPush: GetStaticPushList error=%v", err)
		return
	}

	for _, pushurl := range pushurllist {
		pushurl := pushurl + "/" + streamname
		log.Debugf("StopStaticPush: static pushurl=%s", pushurl)

		staticpushObj, err := rtmprelay.GetStaticPushObject(pushurl)
		if (staticpushObj != nil) && (err == nil) {
			staticpushObj.Stop()
			rtmprelay.ReleaseStaticPushObject(pushurl)
			log.Debugf("StopStaticPush: staticpushObj.Stop %s ", pushurl)
		} else {
			log.Debugf("StopStaticPush GetStaticPushObject %s error", pushurl)
		}
	}
}

func (s *Stream) IsSendStaticPush() bool {
	key := s.info.Key

	dscr := strings.Split(key, "/")
	if len(dscr) < 1 {
		return false
	}

	appname := dscr[0]

	//log.Debugf("SendStaticPush: current streamname=%s， appname=%s", streamname, appname)
	pushurllist, err := rtmprelay.GetStaticPushList(appname)
	if err != nil || len(pushurllist) < 1 {
		//log.Debugf("SendStaticPush: GetStaticPushList error=%v", err)
		return false
	}

	index := strings.Index(key, "/")
	if index < 0 {
		return false
	}

	streamname := key[index+1:]

	for _, pushurl := range pushurllist {
		pushurl := pushurl + "/" + streamname
		//log.Debugf("SendStaticPush: static pushurl=%s", pushurl)

		staticpushObj, err := rtmprelay.GetStaticPushObject(pushurl)
		if (staticpushObj != nil) && (err == nil) {
			return true
			//staticpushObj.WriteAvPacket(&packet)
			//log.Debugf("SendStaticPush: WriteAvPacket %s ", pushurl)
		} else {
			log.Debugf("SendStaticPush GetStaticPushObject %s error", pushurl)
		}
	}
	return false
}

func (s *Stream) SendStaticPush(packet av.Packet) {
	key := s.info.Key

	dscr := strings.Split(key, "/")
	if len(dscr) < 1 {
		return
	}

	index := strings.Index(key, "/")
	if index < 0 {
		return
	}

	streamname := key[index+1:]
	appname := dscr[0]

	//log.Debugf("SendStaticPush: current streamname=%s， appname=%s", streamname, appname)
	pushurllist, err := rtmprelay.GetStaticPushList(appname)
	if err != nil || len(pushurllist) < 1 {
		//log.Debugf("SendStaticPush: GetStaticPushList error=%v", err)
		return
	}

	for _, pushurl := range pushurllist {
		pushurl := pushurl + "/" + streamname
		//log.Debugf("SendStaticPush: static pushurl=%s", pushurl)

		staticpushObj, err := rtmprelay.GetStaticPushObject(pushurl)
		if (staticpushObj != nil) && (err == nil) {
			staticpushObj.WriteAvPacket(&packet)
			//log.Debugf("SendStaticPush: WriteAvPacket %s ", pushurl)
		} else {
			log.Debugf("SendStaticPush GetStaticPushObject %s error", pushurl)
		}
	}
}

func (s *Stream) TransStart() {
	s.isStart = true
	var p av.Packet

	log.Debugf("TransStart: %v", s.info)

	s.StartStaticPush()

	for {
		if !s.isStart {
			s.closeInter()
			return
		}
		err := s.r.Read(&p)
		if err != nil {
			s.closeInter()
			s.isStart = false
			return
		}

		if s.IsSendStaticPush() {
			s.SendStaticPush(p)
		}

		s.cache.Write(p)

		for item := range s.ws.IterBuffered() {
			v := item.Val.(*PackWriterCloser)
			if !v.init {
				//log.Debugf("cache.send: %v", v.w.Info())
				if err = s.cache.Send(v.w); err != nil {
					log.Debugf("[%s] send cache packet error: %v, remove", v.w.Info(), err)
					s.ws.Remove(item.Key)
					continue
				}
				v.init = true
			} else {
				new_packet := p
				//writeType := reflect.TypeOf(v.w)
				//log.Debugf("w.Write: type=%v, %v", writeType, v.w.Info())
				if err = v.w.Write(&new_packet); err != nil {
					log.Debugf("[%s] write packet error: %v, remove", v.w.Info(), err)
					s.ws.Remove(item.Key)
				}
			}
		}
	}
}

func (s *Stream) TransStop() {
	log.Debugf("TransStop: %s", s.info.Key)

	if s.isStart && s.r != nil {
		s.r.Close(fmt.Errorf("stop old"))
	}

	s.isStart = false
}

func (s *Stream) CheckAlive() (n int) {
	if s.r != nil && s.isStart {
		if s.r.Alive() {
			n++
		} else {
			s.r.Close(fmt.Errorf("read timeout"))
		}
	}
	for item := range s.ws.IterBuffered() {
		v := item.Val.(*PackWriterCloser)
		if v.w != nil {
			if !v.w.Alive() && s.isStart {
				s.ws.Remove(item.Key)
				v.w.Close(fmt.Errorf("write timeout"))
				continue
			}
			n++
		}

	}
	return
}

func (s *Stream) closeInter() {
	if s.r != nil {
		s.StopStaticPush()
		log.Debugf("[%v] publisher closed", s.r.Info())
	}

	for item := range s.ws.IterBuffered() {
		v := item.Val.(*PackWriterCloser)
		if v.w != nil {
			if v.w.Info().IsInterval() {
				v.w.Close(fmt.Errorf("closed"))
				s.ws.Remove(item.Key)
				log.Debugf("[%v] player closed and remove\n", v.w.Info())
			}
		}
	}
}
