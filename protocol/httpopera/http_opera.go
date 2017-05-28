package httpopera

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"

	"github.com/golang/glog"
	"github.com/gwuhaolin/livego/utils/uid"
	"github.com/gwuhaolin/livego/av"
	"log"
	"github.com/gwuhaolin/livego/protocol/rtmp"
)

type Response struct {
	w       http.ResponseWriter
	Status  int    `json:"status"`
	Message string `json:"message"`
}

func (r *Response) SendJson() (int, error) {
	resp, _ := json.Marshal(r)
	r.w.Header().Set("Content-Type", "application/json")
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

type Server struct {
	handler av.Handler
}

func NewServer(h av.Handler) *Server {
	return &Server{
		handler: h,
	}
}

func (s *Server) Serve(l net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/rtmp/operation", func(w http.ResponseWriter, r *http.Request) {
		s.handleOpera(w, r)
	})
	// mux.HandleFunc("/rtmp/operation/change", func(w http.ResponseWriter, r *http.Request) {
	// 	s.handleOperaChange(w, r)
	// })
	http.Serve(l, mux)
	return nil
}

// handleOpera, 拉流和推流的http api
// @Path: /rtmp/operation
// @Method: POST
// @Param: json
// 		method string, "push" or "pull"
//		url string
// 		stop bool

// @Example,
// curl -v -H "Content-Type: application/json" -X POST --data \
// '{"method":"pull","url":"rtmp://127.0.0.1:1935/live/test"}' \
// http://localhost:8087/rtmp/operation
func (s *Server) handleOpera(w http.ResponseWriter, r *http.Request) {
	rep := &Response{
		w: w,
	}

	if r.Method != "POST" {
		rep.Status = 14000
		rep.Message = "bad request method"
		rep.SendJson()
		return
	} else {
		result, err := ioutil.ReadAll(r.Body)
		if err != nil {
			rep.Status = 15000
			rep.Message = "read request body error"
			rep.SendJson()
			return
		}
		r.Body.Close()
		glog.Infof("post body: %s\n", result)

		var op Operation
		err = json.Unmarshal(result, &op)
		if err != nil {
			rep.Status = 12000
			rep.Message = "parse json body failed"
			rep.SendJson()
			return
		}

		switch op.Method {
		case "push":
			s.Push(op.URL, op.Stop)
		case "pull":
			s.Pull(op.URL, op.Stop)
		}

		rep.Status = 10000
		rep.Message = op.Method + " " + op.URL + " success"
		rep.SendJson()
	}
}

func (s *Server) Push(uri string, stop bool) error {
	rtmpClient := rtmp.NewRtmpClient(s.handler, nil)
	return rtmpClient.Dial(uri, av.PUBLISH)
	// return nil
}

func (s *Server) Pull(uri string, stop bool) error {
	rtmpClient := rtmp.NewRtmpClient(s.handler, nil)
	return rtmpClient.Dial(uri, av.PLAY)
	// return nil
}

// TODO:
// handleOperaChange, 拉流和推流的http api,支持自定义路径
// @Path: /rtmp/operation/change
// @Method: POST
// @Param: json
// 		method string, "push" or "pull"
//		url string
// 		stop bool

// @Example,
// curl -v -H "Content-Type: application/json" -X POST --data \
// '{"method":"pull","url":"rtmp://127.0.0.1:1935/live/test"}' \
// http://localhost:8087/rtmp/operation
// func (s *Server) handleOperaChange(w http.ResponseWriter, r *http.Request) {
// 	rep := &Response{
// 		w: w,
// 	}

// 	if r.Method != "POST" {
// 		rep.Status = 14000
// 		rep.Message = "bad request method"
// 		rep.SendJson()
// 		return
// 	} else {
// 		result, err := ioutil.ReadAll(r.Body)
// 		if err != nil {
// 			rep.Status = 15000
// 			rep.Message = "read request body error"
// 			rep.SendJson()
// 			return
// 		}
// 		r.Body.Close()
// 		glog.Infof("post body: %s\n", result)

// 		var op OperationChange
// 		err = json.Unmarshal(result, &op)
// 		if err != nil {
// 			rep.Status = 12000
// 			rep.Message = "parse json body failed"
// 			rep.SendJson()
// 			return
// 		}

// 		switch op.Method {
// 		case "push":
// 			s.PushChange(op.SourceURL, op.TargetURL, op.Stop)

// 		case "pull":
// 			s.PullChange(op.SourceURL, op.TargetURL, op.Stop)
// 		}

// 		rep.Status = 10000
// 		rep.Message = op.Method + " from" + op.SourceURL + "to " + op.TargetURL + " success"
// 		rep.SendJson()
// 	}
// }

// pushChange suri to turi
// func (s *Server) PushChange(suri, turi string, stop bool) error {
// if !stop {
// 	sinfo := parseURL(suri)
// 	tinfo := parseURL(turi)
// 	rtmpClient := rtmp.NewRtmpClient(s.handler, nil)
// 	return rtmpClient.Dial(turi, av.PUBLISH)
// } else {
// 	sinfo := parseURL(suri)
// 	tinfo := parseURL(turi)
// 	s.delStream(sinfo.Key, true)
// 	return nil
// }
// return nil
// }

// pullChange
// func (s *Server) PullChange(suri, turi string, stop bool) error {
// if !stop {
// 	rtmpStreams, ok := s.handler.(*rtmp.RtmpStream)
// 	if ok {
// 		streams := rtmpStreams.GetStreams()
// 		rtmpClient := rtmp.NewRtmpClient(s.handler, nil)
// 		return rtmpClient.Dial(turi, av.PLAY)
// 	}
// } else {
// 	info := parseURL(suri)
// 	s.delStream(info.Key, false)
// 	return nil
// }
// return nil
// }

func parseURL(URL string) (ret av.Info) {
	ret.UID = uid.NEWID()
	ret.URL = URL
	_url, err := url.Parse(URL)
	if err != nil {
		log.Println(err)
	}
	ret.Key = strings.TrimLeft(_url.Path, "/")
	ret.Inter = true
	return
}
