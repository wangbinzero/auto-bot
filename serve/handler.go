package serve

import "fmt"

func MsgHandler(conn *Connection, data []byte) {
	fmt.Println("read message ", string(data))
	ConnPool["client1"] = conn
	fmt.Println("加入会话")
}
