package main

import (
	"auto-bot/client"
	"auto-bot/common"
	"auto-bot/serve"
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
			log.Println("深度   ------------   卖 :", response.Asks[0])
		}

		if len(response.Bids) > 0 {
			log.Println("深度   ------------   买 :", response.Bids[0])
		}

	}, func(response *client.KlineResponse) {
		log.Println("K线数据", response)
	}, nil)

	ws.SubscribeKline("btcusdt")
	ws.SubscribeDepth("btcusdt")
	//ws.SubscribeDepth("ethusdt")
	go serve.New().Listen()
	<-channel

}
