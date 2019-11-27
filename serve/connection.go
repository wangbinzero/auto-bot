package serve

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"sync"
)

type Connection struct {
	Conn   *websocket.Conn
	once   sync.Once
	stopCh chan struct{}
	Id     string
	lock   sync.RWMutex
	Event  map[string][]string
}

func (c *Connection) GetId() string {
	c.once.Do(func() {
		uid := uuid.New()
		c.Id = uid.String()
	})
	return c.Id
}

func NewConn(conn *websocket.Conn) *Connection {
	return &Connection{Conn: conn, stopCh: make(chan struct{}), lock: sync.RWMutex{}, Event: map[string][]string{}}
}

func (c *Connection) OnConnect(uuid string) {
	fmt.Sprintf("client %s active\n", c.Conn.RemoteAddr().String())
	bytes, _ := json.Marshal(uuid)
	c.WriteMessage(bytes)
	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			fmt.Sprintf("read message error %v\n", err)
		}
		MsgHandler(c, data)
	}

}

// 回写消息
func (c *Connection) WriteMessage(data []byte) {
	c.Conn.WriteMessage(websocket.TextMessage, data)
}
