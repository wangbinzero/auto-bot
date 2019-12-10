package main

import (
	"auto-bot/client"
	"auto-bot/model"
	"fmt"
)

func main() {
	var channel chan struct{}
	client := client.NewBinance()
	client.SetProxyUrl("socks5://127.0.0.1:1080")
	//binance.SubscribeDepth("btcusdt")
	client.SetCallbacks(func(response *model.BinanceKlineResponse) {
		fmt.Println("Binance [kline] - : ", response.Kline.Close, response.Kline.Low)
	}, nil, nil)
	client.SubscribeKline([]string{"btcusdt", "eosusdt", "bchusdt"})
	<-channel

}
