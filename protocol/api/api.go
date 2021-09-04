package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/configure"
	"github.com/gwuhaolin/livego/protocol/rtmp"
	"github.com/gwuhaolin/livego/protocol/rtmp/rtmprelay"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	log "github.com/sirupsen/logrus"
)

type Response struct {
	w      http.ResponseWriter
	Status int         `json:"status"`
	Data   interface{} `json:"data"`
}

func (r *Response) SendJson() (int, error) {
	resp, _ := json.Marshal(r)
	r.w.Header().Set("Content-Type", "application/json")
	r.w.WriteHeader(r.Status)
	return r.w.Write(resp)
}

type Operation struct {
	Method string `json:"method"`
	URL    string `json:"url"`
	Stop   bool   `json:"stop"`
}

type OperationChange struct {
	Method    string `json:"method"`
	SourceURL string `json:"source_url"`
	TargetURL string `json:"target_url"`
	Stop      bool   `json:"stop"`
}

type ClientInfo struct {
	url              string
	rtmpRemoteClient *rtmp.Client
	rtmpLocalClient  *rtmp.Client
}

type Server struct {
	handler  av.Handler
	session  map[string]*rtmprelay.RtmpRelay
	rtmpAddr string
}

func NewServer(h av.Handler, rtmpAddr string) *Server {
	return &Server{
		handler:  h,
		session:  make(map[string]*rtmprelay.RtmpRelay),
		rtmpAddr: rtmpAddr,
	}
}

func JWTMiddleware(next http.Handler) http.Handler {
	isJWT := len(configure.Config.GetString("jwt.secret")) > 0
	if !isJWT {
		return next
	}

	log.Info("Using JWT middleware")

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var algorithm jwt.SigningMethod
		if len(configure.Config.GetString("jwt.algorithm")) > 0 {
			algorithm = jwt.GetSigningMethod(configure.Config.GetString("jwt.algorithm"))
		}

		if algorithm == nil {
			algorithm = jwt.SigningMethodHS256
		}

		jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
			Extractor: jwtmiddleware.FromFirst(jwtmiddleware.FromAuthHeader, jwtmiddleware.FromParameter("jwt")),
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return []byte(configure.Config.GetString("jwt.secret")), nil
			},
			SigningMethod: algorithm,
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err string) {
				res := &Response{
					w:      w,
					Status: 403,
					Data:   err,
				}
				res.SendJson()
			},
		})

		jwtMiddleware.HandlerWithNext(w, r, next.ServeHTTP)
	})
}

func (s *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()

	mux.Handle("/statics/", http.StripPrefix("/statics/", http.FileServer(http.Dir("statics"))))

	mux.HandleFunc("/control/push", func(w http.ResponseWriter, r *http.Request) {
		s.handlePush(w, r)
	})
	mux.HandleFunc("/control/pull", func(w http.ResponseWriter, r *http.Request) {
		s.handlePull(w, r)
	})
	mux.HandleFunc("/control/get", func(w http.ResponseWriter, r *http.Request) {
		s.handleGet(w, r)
	})
	mux.HandleFunc("/control/reset", func(w http.ResponseWriter, r *http.Request) {
		s.handleReset(w, r)
	})
	mux.HandleFunc("/control/delete", func(w http.ResponseWriter, r *http.Request) {
		s.handleDelete(w, r)
	})
	mux.HandleFunc("/stat/livestat", func(w http.ResponseWriter, r *http.Request) {
		s.GetLiveStatics(w, r)
	})
	http.Serve(l, JWTMiddleware(mux))
	return nil
}

type stream struct {
	Key             string `json:"key"`
	Url             string `json:"url"`
	StreamId        uint32 `json:"stream_id"`
	VideoTotalBytes uint64 `json:"video_total_bytes"`
	VideoSpeed      uint64 `json:"video_speed"`
	AudioTotalBytes uint64 `json:"audio_total_bytes"`
	AudioSpeed      uint64 `json:"audio_speed"`
}

type streams struct {
	Publishers []stream `json:"publishers"`
	Players    []stream `json:"players"`
}

//http://127.0.0.1:8090/stat/livestat
func (server *Server) GetLiveStatics(w http.ResponseWriter, req *http.Request) {
	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}

	defer res.SendJson()

	room := ""

	if err := req.ParseForm(); err == nil {
		room = req.Form.Get("room")
	}

	rtmpStream := server.handler.(*rtmp.RtmpStream)
	if rtmpStream == nil {
		res.Status = 500
		res.Data = "Get rtmp stream information error"
		return
	}

	msgs := new(streams)

	if room == "" {
		rtmpStream.GetStreams().Range(func(key, val interface{}) bool {
			if s, ok := val.(*rtmp.Stream); ok {
				if s.GetReader() != nil {
					switch s.GetReader().(type) {
					case *rtmp.VirReader:
						v := s.GetReader().(*rtmp.VirReader)
						msg := stream{key.(string), v.Info().URL, v.ReadBWInfo.StreamId, v.ReadBWInfo.VideoDatainBytes, v.ReadBWInfo.VideoSpeedInBytesperMS,
							v.ReadBWInfo.AudioDatainBytes, v.ReadBWInfo.AudioSpeedInBytesperMS}
						msgs.Publishers = append(msgs.Publishers, msg)
					}
				}
			}
			return true
		})

		rtmpStream.GetStreams().Range(func(key, val interface{}) bool {
			ws := val.(*rtmp.Stream).GetWs()
			ws.Range(func(k, v interface{}) bool {
				if pw, ok := v.(*rtmp.PackWriterCloser); ok {
					if pw.GetWriter() != nil {
						switch pw.GetWriter().(type) {
						case *rtmp.VirWriter:
							v := pw.GetWriter().(*rtmp.VirWriter)
							msg := stream{key.(string), v.Info().URL, v.WriteBWInfo.StreamId, v.WriteBWInfo.VideoDatainBytes, v.WriteBWInfo.VideoSpeedInBytesperMS,
								v.WriteBWInfo.AudioDatainBytes, v.WriteBWInfo.AudioSpeedInBytesperMS}
							msgs.Players = append(msgs.Players, msg)
						}
					}
				}
				return true
			})
			return true
		})
	} else {
        // Warning: The room should be in the "live/stream" format!
		roomInfo, exists := (rtmpStream.GetStreams()).Load(room)
		if exists == false {
			res.Status = 404
			res.Data = "room not found or inactive"
			return
		}

		if s, ok := roomInfo.(*rtmp.Stream); ok {
			if s.GetReader() != nil {
				switch s.GetReader().(type) {
				case *rtmp.VirReader:
					v := s.GetReader().(*rtmp.VirReader)
					msg := stream{room, v.Info().URL, v.ReadBWInfo.StreamId, v.ReadBWInfo.VideoDatainBytes, v.ReadBWInfo.VideoSpeedInBytesperMS,
						v.ReadBWInfo.AudioDatainBytes, v.ReadBWInfo.AudioSpeedInBytesperMS}
					msgs.Publishers = append(msgs.Publishers, msg)
				}
			}

			s.GetWs().Range(func(k, v interface{}) bool {
				if pw, ok := v.(*rtmp.PackWriterCloser); ok {
					if pw.GetWriter() != nil {
						switch pw.GetWriter().(type) {
						case *rtmp.VirWriter:
							v := pw.GetWriter().(*rtmp.VirWriter)
							msg := stream{room, v.Info().URL, v.WriteBWInfo.StreamId, v.WriteBWInfo.VideoDatainBytes, v.WriteBWInfo.VideoSpeedInBytesperMS,
								v.WriteBWInfo.AudioDatainBytes, v.WriteBWInfo.AudioSpeedInBytesperMS}
							msgs.Players = append(msgs.Players, msg)
						}
					}
				}
				return true
			})
		}
	}

	//resp, _ := json.Marshal(msgs)
	res.Data = msgs
}

//http://127.0.0.1:8090/control/pull?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456
func (s *Server) handlePull(w http.ResponseWriter, req *http.Request) {
	var retString string
	var err error

	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}

	defer res.SendJson()

	if req.ParseForm() != nil {
		res.Status = 400
		res.Data = "url: /control/pull?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456"
		return
	}

	oper := req.Form.Get("oper")
	app := req.Form.Get("app")
	name := req.Form.Get("name")
	url := req.Form.Get("url")

	log.Debugf("control pull: oper=%v, app=%v, name=%v, url=%v", oper, app, name, url)
	if (len(app) <= 0) || (len(name) <= 0) || (len(url) <= 0) {
		res.Status = 400
		res.Data = "control push parameter error, please check them."
		return
	}

	remoteurl := "rtmp://127.0.0.1" + s.rtmpAddr + "/" + app + "/" + name
	localurl := url

	keyString := "pull:" + app + "/" + name
	if oper == "stop" {
		pullRtmprelay, found := s.session[keyString]

		if !found {
			retString = fmt.Sprintf("session key[%s] not exist, please check it again.", keyString)
			res.Status = 400
			res.Data = retString
			return
		}
		log.Debugf("rtmprelay stop push %s from %s", remoteurl, localurl)
		pullRtmprelay.Stop()

		delete(s.session, keyString)
		retString = fmt.Sprintf("<h1>push url stop %s ok</h1></br>", url)
		res.Status = 400
		res.Data = retString
		log.Debugf("pull stop return %s", retString)
	} else {
		pullRtmprelay := rtmprelay.NewRtmpRelay(&localurl, &remoteurl)
		log.Debugf("rtmprelay start push %s from %s", remoteurl, localurl)
		err = pullRtmprelay.Start()
		if err != nil {
			retString = fmt.Sprintf("push error=%v", err)
		} else {
			s.session[keyString] = pullRtmprelay
			retString = fmt.Sprintf("<h1>push url start %s ok</h1></br>", url)
		}
		res.Status = 400
		res.Data = retString
		log.Debugf("pull start return %s", retString)
	}
}

//http://127.0.0.1:8090/control/push?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456
func (s *Server) handlePush(w http.ResponseWriter, req *http.Request) {
	var retString string
	var err error

	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}

	defer res.SendJson()

	if req.ParseForm() != nil {
		res.Data = "url: /control/push?&oper=start&app=live&name=123456&url=rtmp://192.168.16.136/live/123456"
		return
	}

	oper := req.Form.Get("oper")
	app := req.Form.Get("app")
	name := req.Form.Get("name")
	url := req.Form.Get("url")

	log.Debugf("control push: oper=%v, app=%v, name=%v, url=%v", oper, app, name, url)
	if (len(app) <= 0) || (len(name) <= 0) || (len(url) <= 0) {
		res.Data = "control push parameter error, please check them."
		return
	}

	localurl := "rtmp://127.0.0.1" + s.rtmpAddr + "/" + app + "/" + name
	remoteurl := url

	keyString := "push:" + app + "/" + name
	if oper == "stop" {
		pushRtmprelay, found := s.session[keyString]
		if !found {
			retString = fmt.Sprintf("<h1>session key[%s] not exist, please check it again.</h1>", keyString)
			res.Data = retString
			return
		}
		log.Debugf("rtmprelay stop push %s from %s", remoteurl, localurl)
		pushRtmprelay.Stop()

		delete(s.session, keyString)
		retString = fmt.Sprintf("<h1>push url stop %s ok</h1></br>", url)
		res.Data = retString
		log.Debugf("push stop return %s", retString)
	} else {
		pushRtmprelay := rtmprelay.NewRtmpRelay(&localurl, &remoteurl)
		log.Debugf("rtmprelay start push %s from %s", remoteurl, localurl)
		err = pushRtmprelay.Start()
		if err != nil {
			retString = fmt.Sprintf("push error=%v", err)
		} else {
			retString = fmt.Sprintf("<h1>push url start %s ok</h1></br>", url)
			s.session[keyString] = pushRtmprelay
		}

		res.Data = retString
		log.Debugf("push start return %s", retString)
	}
}

//http://127.0.0.1:8090/control/reset?room=ROOM_NAME
func (s *Server) handleReset(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}
	defer res.SendJson()

	if err := r.ParseForm(); err != nil {
		res.Status = 400
		res.Data = "url: /control/reset?room=<ROOM_NAME>"
		return
	}
	room := r.Form.Get("room")

	if len(room) == 0 {
		res.Status = 400
		res.Data = "url: /control/reset?room=<ROOM_NAME>"
		return
	}

	msg, err := configure.RoomKeys.SetKey(room)

	if err != nil {
		msg = err.Error()
		res.Status = 400
	}

	res.Data = msg
}

//http://127.0.0.1:8090/control/get?room=ROOM_NAME
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}
	defer res.SendJson()

	if err := r.ParseForm(); err != nil {
		res.Status = 400
		res.Data = "url: /control/get?room=<ROOM_NAME>"
		return
	}

	room := r.Form.Get("room")

	if len(room) == 0 {
		res.Status = 400
		res.Data = "url: /control/get?room=<ROOM_NAME>"
		return
	}

	msg, err := configure.RoomKeys.GetKey(room)
	if err != nil {
		msg = err.Error()
		res.Status = 400
	}
	res.Data = msg
}

//http://127.0.0.1:8090/control/delete?room=ROOM_NAME
func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	res := &Response{
		w:      w,
		Data:   nil,
		Status: 200,
	}
	defer res.SendJson()

	if err := r.ParseForm(); err != nil {
		res.Status = 400
		res.Data = "url: /control/delete?room=<ROOM_NAME>"
		return
	}

	room := r.Form.Get("room")

	if len(room) == 0 {
		res.Status = 400
		res.Data = "url: /control/delete?room=<ROOM_NAME>"
		return
	}

	if configure.RoomKeys.DeleteChannel(room) {
		res.Data = "Ok"
		return
	}
	res.Status = 404
	res.Data = "room not found"
}
