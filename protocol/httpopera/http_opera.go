package httpopera

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"log"
	"github.com/gwuhaolin/livego/av"
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
		log.Println("post body", result)

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
}

func (s *Server) Pull(uri string, stop bool) error {
	rtmpClient := rtmp.NewRtmpClient(s.handler, nil)
	return rtmpClient.Dial(uri, av.PLAY)
}
