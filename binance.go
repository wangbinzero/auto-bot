package main

import "auto-bot/client"

func main() {
	var channel chan struct{}
	client := client.NewBinance()
	client.SetProxyUrl("socks5://127.0.0.1:1086")
	//binance.SubscribeDepth("btcusdt")

	client.SubscribeKline([]string{"btcusdt"})
	<-channel

}
