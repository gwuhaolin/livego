package rtmp

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/rtmp/cache"
	"github.com/gwuhaolin/livego/protocol/rtmp/rtmprelay"

	log "github.com/sirupsen/logrus"
)

var (
	EmptyID = ""
)

// 流媒体服务器。包含N个流媒体服务。
// 每个单独的流媒体服务都可以有多种输入流，多个输出流。
type StreamServer struct {
	services *sync.Map //key
}

// 创建流媒体服务器。并启动状态监测协程。
func NewStreamServers() *StreamServer {
	ss := &StreamServer{
		services: &sync.Map{},
	}
	go ss.CheckAlive(5)
	return ss
}

// 注册源流处理逻辑。并启动源流读取。
// 若已有源流，则重新启动；否则创建流媒体服务。
func (ss *StreamServer) HandleReader(r av.ReadCloser) {
	info := r.Info()
	log.Debugf("HandleReader: info[%v]", info)

	var service *StreamService
	i, ok := ss.services.Load(info.Key)
	if service, ok = i.(*StreamService); ok {
		service.TransStop()
		id := service.ID()
		if id != EmptyID && id != info.UID {
			ns := NewStreamService()
			service.Copy(ns)
			service = ns
			ss.services.Store(info.Key, ns)
		}
	} else {
		service = NewStreamService()
		ss.services.Store(info.Key, service)
		service.info = info
	}

	service.AddReader(r)
}

// 注册目标流处理逻辑。
// 若没有则新创建流媒体服务；若有则新增流媒体服务writer。
func (ss *StreamServer) HandleWriter(w av.WriteCloser) {
	info := w.Info()
	log.Debugf("HandleWriter: info[%v]", info)

	var service *StreamService
	item, ok := ss.services.Load(info.Key)
	if !ok {
		log.Debugf("HandleWriter: not found create new info[%v]", info)
		service = NewStreamService()
		ss.services.Store(info.Key, service)
		service.info = info
	} else {
		service = item.(*StreamService)
		service.AddWriter(w)
	}
}

// 获取所有流媒体服务
func (ss *StreamServer) GetServices() *sync.Map {
	return ss.services
}

// 定时遍历检测所有媒体服务器
func (ss *StreamServer) CheckAlive(ttl uint) {

	if ttl <= 1 {
		ttl = 1
	}

	for {
		<-time.After(time.Duration(ttl) * time.Second)
		ss.services.Range(func(key, val interface{}) bool {
			v := val.(*StreamService)
			if v.CheckAlive() == 0 {
				ss.services.Delete(key)
			}
			return true
		})
	}
}

// 流媒体服务结构信息
type StreamService struct {
	isStart bool
	info    av.Info       // 源流信息
	cache   *cache.Cache  // 源流视频数据
	r       av.ReadCloser // 读源流handler
	ws      *sync.Map     // 推流目标地址集合
}

// 写流数据
type PackWriterCloser struct {
	init bool
	w    av.WriteCloser // 写目标流handler
}

func (p *PackWriterCloser) GetWriter() av.WriteCloser {
	return p.w
}

// 实例化创建媒体服务。
func NewStreamService() *StreamService {
	return &StreamService{
		cache: cache.NewCache(),
		ws:    &sync.Map{},
	}
}

func (s *StreamService) ID() string {
	if s.r != nil {
		return s.r.Info().UID
	}
	return EmptyID
}

func (s *StreamService) GetReader() av.ReadCloser {
	return s.r
}

func (s *StreamService) GetWs() *sync.Map {
	return s.ws
}

// 复制流媒体服务。
func (s *StreamService) Copy(dst *StreamService) {
	dst.info = s.info
	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		s.ws.Delete(key)
		v.w.CalcBaseTimestamp()
		dst.AddWriter(v.w)
		return true
	})
}

// 新增源流处理。并启动读取流数据协程。
func (s *StreamService) AddReader(r av.ReadCloser) {
	s.r = r
	go s.TransStart()
}

func (s *StreamService) AddWriter(w av.WriteCloser) {
	info := w.Info()
	pw := &PackWriterCloser{w: w}
	s.ws.Store(info.UID, pw)
}

/*检测本application下是否配置static_push,
如果配置, 启动push远端的连接*/
func (s *StreamService) StartStaticPush() {
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

func (s *StreamService) StopStaticPush() {
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

func (s *StreamService) IsSendStaticPush() bool {
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

func (s *StreamService) SendStaticPush(packet av.Packet) {
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

// 流媒体服务读取源流数据协程
func (s *StreamService) TransStart() {
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

		s.ws.Range(func(key, val interface{}) bool {
			v := val.(*PackWriterCloser)
			if !v.init {
				//log.Debugf("cache.send: %v", v.w.Info())
				if err = s.cache.Send(v.w); err != nil {
					log.Debugf("[%s] send cache packet error: %v, remove", v.w.Info(), err)
					s.ws.Delete(key)
					return true
				}
				v.init = true
			} else {
				newPacket := p
				//writeType := reflect.TypeOf(v.w)
				//log.Debugf("w.Write: type=%v, %v", writeType, v.w.Info())
				if err = v.w.Write(&newPacket); err != nil {
					log.Debugf("[%s] write packet error: %v, remove", v.w.Info(), err)
					s.ws.Delete(key)
				}
			}
			return true
		})
	}
}

// 停止读取源流，并重置状态。
func (s *StreamService) TransStop() {
	log.Debugf("TransStop: %s", s.info.Key)

	if s.isStart && s.r != nil {
		s.r.Close(fmt.Errorf("stop old"))
	}

	s.isStart = false
}

// 检测某个媒体服务器状态
func (s *StreamService) CheckAlive() (n int) {
	if s.r != nil && s.isStart {
		if s.r.Alive() {
			n++
		} else {
			s.r.Close(fmt.Errorf("read timeout"))
		}
	}

	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		if v.w != nil {
			//Alive from RWBaser, check last frame now - timestamp, if > timeout then Remove it
			if !v.w.Alive() && s.isStart {
				log.Infof("write timeout remove")
				s.ws.Delete(key)
				v.w.Close(fmt.Errorf("write timeout"))
				return true
			}
			n++
		}
		return true
	})

	return
}

func (s *StreamService) closeInter() {
	if s.r != nil {
		s.StopStaticPush()
		log.Debugf("[%v] publisher closed", s.r.Info())
	}

	s.ws.Range(func(key, val interface{}) bool {
		v := val.(*PackWriterCloser)
		if v.w != nil {
			if v.w.Info().IsInterval() {
				v.w.Close(fmt.Errorf("closed"))
				s.ws.Delete(key)
				log.Debugf("[%v] player closed and remove\n", v.w.Info())
			}
		}
		return true
	})
}
