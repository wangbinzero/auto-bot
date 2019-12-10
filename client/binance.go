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
	detailCallback func(*BinanceDetailResponse)
}

// 创建币安客户端
func NewBinance() *binanceClient {
	baClient := &binanceClient{WsBuilder: core.NewWs()}
	baClient.WsBuilder.
		SetWsUrl("wss://stream.binance.com:9443/ws").
		SetErrorHandle(baClient.errorHandle).
		SetProtoHandleFunc(baClient.protocolHandle)
	return baClient
}

func (c *binanceClient) connectWs() {
	c.Do(func() {
		c.wsConn = c.WsBuilder.Build()
		c.wsConn.ReceiveMessage()
	})
}

// 消息协议处理
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
			var detailRes BinanceDetailResponse
			json.Unmarshal(data, &detailRes)
			ba.detailCallback(&detailRes)
		case "depthUpdate":
			var depthRes BinanceDepthResponse
			json.Unmarshal(data, &depthRes)
			ba.depthCallback(&depthRes)
		}
	}
	return nil
}

// 币安深度订阅
func (ba *binanceClient) SubscribeDepth(symbol []string) error {

	return ba.subscribe(map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": symbol,
		"id":     1,
	})
}

// 币安详情订阅
func (ba *binanceClient) SubscribeDetail(symbol []string) error {
	return ba.subscribe(map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": symbol,
		"id":     1,
	})
}

// K线订阅
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

func (ba *binanceClient) subscribe(data map[string]interface{}) error {
	fmt.Println("订阅交易对:", data)
	ba.connectWs()
	return ba.wsConn.Subscribe(data)
}

// 异常处理器
func (ba *binanceClient) errorHandle(err error) {
	fmt.Println("币安异常信息:", err)
}

func (ba *binanceClient) SetCallbacks(kline func(response *BinanceDepthResponse)) {

}
