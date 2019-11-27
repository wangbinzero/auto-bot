package serve

import (
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
)

var upgrade = websocket.Upgrader{CheckOrigin: func(r *http.Request) bool {
	return true
}, ReadBufferSize: 1024, WriteBufferSize: 1024}

type Serve struct {
}

type wsHandle struct {
	*Serve
}

func (h *wsHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrade.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("serve start error", err)
		return
	}
	c := NewConn(conn)
	uuid := c.GetId()
	c.OnConnect(uuid)
}

func (s *Serve) handler() http.Handler {
	return &wsHandle{s}
}

func New() *Serve {
	return &Serve{}
}

func (s *Serve) Listen() {
	http.Handle("/", s.handler())
	err := http.ListenAndServe(":12345", nil)
	if err != nil {
		fmt.Println("serve Listen error", err)
		return
	} else {
		fmt.Println("serve Listen success")
	}
}
