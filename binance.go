package main

import "auto-bot/client"

func main() {
	var channel chan struct{}
	binance := client.NewBAClient()
	binance.SetProxyUrl("socks5://127.0.0.1:1086")
	binance.SubscribeDepth("btcusdt")
	<-channel

}
