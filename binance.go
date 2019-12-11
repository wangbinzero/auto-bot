package main

import (
	"auto-bot/config"
	"auto-bot/middleware"
)

func main() {
	var channel chan struct{}
	//client := client.NewBinance()
	//client.SetProxyUrl("socks5://127.0.0.1:1080")
	////binance.SubscribeDepth("btcusdt")
	//client.SetCallbacks(func(response *model.BinanceKlineResponse) {
	//	fmt.Println("Binance [kline] - : ", response.Kline.Close, response.Kline.Low)
	//}, nil, nil)
	//client.SubscribeKline([]string{"btcusdt", "eosusdt", "bchusdt"})

	config.InitEnvironment()
	middleware.InitAmqpConfig()
	middleware.InitAmqpConn()
	str := "hello"
	byte := []byte(str)
	middleware.PublishMessageWithRouteKey("bot.kline", "kline", "text/plain", &byte, nil, 1)

	chann, _ := middleware.RabbitMqConnect.Channel()
	//msg,_:=chann.Consume("queue.kline")
	<-channel

}
