package websocketflv

import (
	"strings"
	"net"
	"net/http"
	"log"
	"github.com/gwuhaolin/livego/av"
	"github.com/gwuhaolin/livego/protocol/httpflv"
	"github.com/gorilla/websocket"
)

type Server struct {
	handler av.Handler
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewServer(h av.Handler) *Server {
	return &Server{
		handler: h,
	}
}

func (server *Server) Serve(listener net.Listener) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleConn)
	http.Serve(listener, mux)
	return nil
}

func (server *Server) handleConn(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("websocket flv handleConn panic: ", r)
		}
	}()

	url := r.URL.String()
	u := r.URL.Path
	if pos := strings.LastIndex(u, "."); pos < 0 || u[pos:] != ".flv" {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	path := strings.TrimSuffix(strings.TrimLeft(u, "/"), ".flv")
	paths := strings.SplitN(path, "/", 2)
	if len(paths) != 2 {
		http.Error(w, "invalid path", http.StatusBadRequest)
		return
	}
	log.Println("url:", u, "path:", path, "paths:", paths)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		conn.Close()
		return
	}

	writer, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		conn.Close()
		return
	}

	flvWriter := httpflv.NewFLVWriter(paths[0], paths[1], url, writer)
	server.handler.HandleWriter(flvWriter)
	flvWriter.Wait()
}
