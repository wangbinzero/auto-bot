package client

import (
	"auto-bot/core"
	"fmt"
	"sync"
)

type BAClient struct {
	*core.WsBuilder
	sync.Once
	wsConn         *core.WsCon
	depthCallback  func()
	klineCallback  func()
	detailCallback func()
}

// 创建币安客户端
func NewBAClient() *BAClient {
	baClient := &BAClient{WsBuilder: core.NewWs()}
	baClient.WsBuilder.
		SetWsUrl("wss://stream.binance.com:9443/ws").
		SetErrorHandle(func(err error) {
			fmt.Println("币安异常信息处理:", err)
		}).SetProtoHandleFunc(baClient.protocolHandle)
	return baClient
}

func (ba *BAClient) connectWs() {
	ba.Do(func() {
		ba.wsConn = ba.WsBuilder.Build()
		ba.wsConn.ReceiveMessage()
	})
}

// 消息协议处理
func (ba *BAClient) protocolHandle(data []byte) error {
	fmt.Println("币安消息:", string(data))
	return nil
}

// 币安深度订阅
func (ba *BAClient) SubscribeDepth(symbol string) error {

	return ba.subscribe(map[string]interface{}{
		"method": "SUBSCRIBE",
		"params": []string{"btcusdt@depth"},
		"id":     1,
	})
	//TODO 设定深度回调函数
	//return ba.subscribe(endPoint)
}

func (ba *BAClient) subscribe(data map[string]interface{}) error {
	fmt.Println("订阅交易对:", data)
	ba.connectWs()
	//data = ba.wsConn.WsUrl + data
	return ba.wsConn.Subscribe(data)
}
