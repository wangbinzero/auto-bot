package serve

import "fmt"

func MsgHandler(data []byte) {
	fmt.Println("read message ", string(data))
}
