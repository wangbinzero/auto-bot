package client

import (
	"auto-bot/core"
	. "auto-bot/model"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
)

type binanceClient struct {
	*core.WsBuilder
	sync.Once
	wsConn *core.WsCon

	//深度回调事件
	depthCallback func(*BinanceDepthResponse)

	//K线回调事件
	klineCallback func(*BinanceKlineResponse)

	//成交明细回调事件
	tickerCallback func(*BinanceTickerResponse)
}

// create binance client
func NewBinance() *binanceClient {
	baClient := &binanceClient{WsBuilder: core.NewWs()}

	baClient.WsBuilder.
		SetWsUrl("wss://stream.binance.com:9443/ws").
		SetErrorHandle(baClient.errorHandle).
		SetProtoHandleFunc(baClient.protocolHandle)

	return baClient
}

// connected to binance websocket.
func (c *binanceClient) connectWs() {

	c.Do(func() {
		c.wsConn = c.WsBuilder.Build()
		c.wsConn.ReceiveMessage()
	})
}

// this func handle message for binance ws
func (ba *binanceClient) protocolHandle(data []byte) error {
	str := string(data)
	if strings.Contains(str, "ping") {
		fmt.Println("币安心跳数据包")
	}

	if !strings.Contains(str, "result") {
		var binance BinanceBaseResponse
		json.Unmarshal(data, &binance)
		switch binance.Type {
		case "kline":
			var klineRes BinanceKlineResponse
			json.Unmarshal(data, &klineRes)
			ba.klineCallback(&klineRes)
		case "24hrTicker":
			var detailRes BinanceTickerResponse
			json.Unmarshal(data, &detailRes)
			ba.tickerCallback(&detailRes)
		case "depthUpdate":
			var depthRes BinanceDepthResponse
			json.Unmarshal(data, &depthRes)
			ba.depthCallback(&depthRes)
		}
	}
	return nil
}

// subscribe depth response data
func (ba *binanceClient) SubscribeDepth(symbol []string) error {

	return ba.subscribe(map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": symbol,
		"id":     1,
	})
}

// subscribe ticker response data
func (ba *binanceClient) SubscribeTicker(symbol []string) error {
	return ba.subscribe(map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": symbol,
		"id":     1,
	})
}

// subscribe kline response data
func (ba *binanceClient) SubscribeKline(symbol []string) error {

	if ba.klineCallback == nil {
		fmt.Println("please setting kline callback func before subscribe.")
		return errors.New("please setting kline callback")
	}
	var params []string

	for _, v := range symbol {

		params = append(params, v+"@kline_1m")
		params = append(params, v+"@kline_5m")
		params = append(params, v+"@kline_15m")
		params = append(params, v+"@kline_30m")
		params = append(params, v+"@kline_1h")
		params = append(params, v+"@kline_1d")
		params = append(params, v+"@kline_1w")
		params = append(params, v+"@kline_1M")
	}
	return ba.subscribe(map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": params,
		"id":     1,
	})
}

// the main func for subscribe
func (ba *binanceClient) subscribe(data map[string]interface{}) error {
	fmt.Println("start subscribe symbol by binance")
	ba.connectWs()
	return ba.wsConn.Subscribe(data)
}

// error handler
func (ba *binanceClient) errorHandle(err error) {
	fmt.Println("币安异常信息:", err)
}

// Setting callback func
func (ba *binanceClient) SetCallbacks(kline func(*BinanceKlineResponse), depth func(*BinanceDepthResponse), ticker func(*BinanceTickerResponse)) {
	ba.depthCallback = depth
	ba.klineCallback = kline
	ba.tickerCallback = ticker
}

func BinanceRun() {
	client := NewBinance()
	client.SetProxyUrl("socks5://127.0.0.1:1080")
	//binance.SubscribeDepth("btcusdt")
	client.SetCallbacks(func(response *BinanceKlineResponse) {
		fmt.Println("Binance [kline] - : ", response.Kline.Close, response.Kline.Low)
	}, nil, nil)
	client.SubscribeKline([]string{"btcusdt", "eosusdt", "bchusdt"})
}
