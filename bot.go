package main

import (
	"auto-bot/client"
	"auto-bot/common"
	"fmt"
	"log"
)

func main() {

	var channel chan struct{}
	fmt.Print(common.Prompt)
	ws := client.NewClient()
	ws.SetFollow(0.0001)
	ws.SetCallbacks(func(response *client.DepthResponse) {
		if len(response.Asks) > 0 {
			log.Println("深度   ------------   卖 :", response.Asks[0][0]*ws.Rate)
		}

		if len(response.Bids) > 0 {
			log.Println("深度   ------------   买 :", response.Bids[0][0]*ws.Rate)
		}

	})

	ws.SubscribeDepth("btcusdt")
	<-channel

	//t := Test{
	//	closeChan: make(chan struct{}, 1),
	//	recvChan:  make(chan struct{}, 1),
	//}
	//t.closeTest()

}

type Test struct {
	closeChan chan struct{}
	recvChan  chan struct{}
}

func (t *Test) closeTest() {
	t.cleanChan(t.closeChan)
	t.cleanChan(t.recvChan)
	t.closeChan <- struct{}{}
	t.recvChan <- struct{}{}

}

func (t *Test) cleanChan(c chan struct{}) {
	for {
		if len(c) > 0 {
			data := <-c
			fmt.Println("通道读取数据: ", data)
		} else {
			fmt.Println("没有收到数据")
			break
		}
	}
}
