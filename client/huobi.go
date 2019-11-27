package client

import (
	"auto-bot/common"
	. "auto-bot/core"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

type WsResponse struct {
	Ch   string
	Ts   int64
	Tick json.RawMessage
}

type DepthResponse struct {
	Bids [][]float64
	Asks [][]float64
}

//火币客户端
type HClient struct {
	*WsBuilder
	sync.Once
	wsConn        *WsCon
	depthCallback func(*DepthResponse)
	Follow        bool
	Rate          float64
}

//初始化火币客户端
func NewClient() *HClient {
	hclient := &HClient{WsBuilder: NewWs()}
	hclient.WsBuilder.
		SetWsUrl("wss://api.huobi.pro/ws").
		SetErrorHandle(func(err error) {
			log.Println("火币异常处理器", err)

		}).
		SetReconnectIntervalTime(20 * time.Minute).
		SetUnCompressFunc(common.GzipUnCompress).
		SetProtoHandleFunc(hclient.protocolHandle).SetProxyUrl("socks5://127.0.0.1:1086")
	return hclient
}

//火币数据协议处理器
func (hc *HClient) protocolHandle(data []byte) error {

	if strings.Contains(string(data), "ping") {
		var ping struct {
			Ping int64
		}
		json.Unmarshal(data, &ping)
		pong := struct {
			Pong int64 `json:"pong"`
		}{ping.Ping}
		hc.wsConn.SendJsonMessage(pong)
		hc.wsConn.UpdateActiveTime()
		return nil
	}

	var res WsResponse
	err := json.Unmarshal(data, &res)
	if err != nil {
		return err
	}

	var depthRes DepthResponse
	json.Unmarshal(res.Tick, &depthRes)
	hc.depthCallback(&depthRes)

	return nil
}

//******************************订阅***************************************//
func (hc *HClient) SubscribeDepth(symbol string) error {
	if hc.depthCallback == nil {
		return errors.New("请设置深度回调事件")
	}
	return hc.subscribe(map[string]interface{}{
		"id":  "client1",
		"sub": fmt.Sprintf("market.%s.depth.step1", symbol),
	})
}

func (hc *HClient) subscribe(sub map[string]interface{}) error {
	log.Println("订阅交易对:", sub)
	hc.connectWs()
	return hc.wsConn.Subscribe(sub)
}

func (hc *HClient) connectWs() {
	hc.Do(func() {
		hc.wsConn = hc.WsBuilder.Build()
		hc.wsConn.ReceiveMessage()
	})
}

//深度回调
func (hc *HClient) SetCallbacks(depthCallback func(*DepthResponse)) {
	hc.depthCallback = depthCallback
}

func (hc *HClient) SetFollow(rate float64) {
	hc.Follow = true
	hc.Rate = rate
}

//解析深度数据
func (hc *HClient) parseDepth() {

}
