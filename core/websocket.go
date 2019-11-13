package core

import (
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"
)

// websocket配置选项
type WsConfig struct {
	WsUrl                 string                       //地址
	ProxyUrl              string                       //代理地址
	RequestHeaders        map[string][]string          //请求头
	HeartbeatIntervalTime time.Duration                //心跳周期
	HeartbeatData         []byte                       //心跳数据
	HeartbeatFunc         func() interface{}           //心跳事件
	ReconnectIntervalTime time.Duration                //重连检测周期
	ProtoHandleFunc       func([]byte) error           //协议处理事件
	UnCompressFunc        func([]byte) ([]byte, error) //解压处理器
	ErrorHandleFunc       func(err error)              //错误处理器
	IsDump                bool
}

// websocket 连接结构体
type WsCon struct {
	*websocket.Conn               //连接对象
	sync.Mutex                    //互斥锁
	WsConfig                      //配置选项
	activeTime      time.Time     //
	activeTimeL     sync.Mutex    //
	mu              chan struct{} //
	closeHeartbeat  chan struct{} //心跳关闭信号通道
	closeReconnect  chan struct{} //关闭连接信号通道
	closeRecv       chan struct{} //关闭消息
	closeCheck      chan struct{} //关闭检查消息
	subs            []interface{} //订阅数据
}

type WsBuilder struct {
	wsConfig *WsConfig
}

// 初始化ws
func NewWs() *WsBuilder {
	return &WsBuilder{&WsConfig{}}
}

// 设置地址
func (w *WsBuilder) SetWsUrl(url string) *WsBuilder {
	w.wsConfig.WsUrl = url
	return w
}

//设置代理
func (w *WsBuilder) SetProxyUrl(url string) *WsBuilder {
	w.wsConfig.ProxyUrl = url
	return w
}

//设置请求头
func (w *WsBuilder) SetReqHeader(key, val string) *WsBuilder {
	w.wsConfig.RequestHeaders[key] = append(w.wsConfig.RequestHeaders[key], val)
	return w
}

//
func (w *WsBuilder) SetDump() *WsBuilder {
	w.wsConfig.IsDump = true
	return w
}

//设置心跳时间周期以及心跳数据
func (w *WsBuilder) SetHeartbeat(data []byte, t time.Duration) *WsBuilder {
	w.wsConfig.HeartbeatIntervalTime = t
	w.wsConfig.HeartbeatData = data
	return w
}

//设置心跳事件以及时间周期
func (w *WsBuilder) SetHeartbeat2(heartbeat func() interface{}, t time.Duration) *WsBuilder {
	w.wsConfig.HeartbeatFunc = heartbeat
	w.wsConfig.HeartbeatIntervalTime = t
	return w
}

//设置重连时间周期
func (w *WsBuilder) SetReconnectIntervalTime(t time.Duration) *WsBuilder {
	w.wsConfig.ReconnectIntervalTime = t
	return w
}

//设置协议处理事件
func (w *WsBuilder) SetProtoHandleFunc(proto func([]byte) error) *WsBuilder {
	w.wsConfig.ProtoHandleFunc = proto
	return w
}

//设置解压事件
func (w *WsBuilder) SetUnCompressFunc(unCompress func([]byte) ([]byte, error)) *WsBuilder {
	w.wsConfig.UnCompressFunc = unCompress
	return w
}

//设置错误处理事件
func (w *WsBuilder) SetErrorHandle(errFunc func(err error)) *WsBuilder {
	w.wsConfig.ErrorHandleFunc = errFunc
	return w
}

//初始化websocketBuilder
func (w *WsBuilder) Build() *WsCon {
	if w.wsConfig.ErrorHandleFunc == nil {
		w.wsConfig.ErrorHandleFunc = func(err error) {
			log.Println("未指定异常处理器，由基类处理:", err)
		}
	}
	wsConn := &WsCon{WsConfig: *w.wsConfig}
	return wsConn.New()
}

//初始化 websocket
func (ws *WsCon) New() *WsCon {
	ws.Lock()
	defer ws.Unlock()
	ws.connect()
	ws.mu = make(chan struct{}, 1)
	ws.closeHeartbeat = make(chan struct{}, 1)
	ws.closeReconnect = make(chan struct{}, 1)
	ws.closeRecv = make(chan struct{}, 1)
	ws.closeCheck = make(chan struct{}, 1)

	ws.HeartbeatTimer()
	ws.ReconnectTimer()
	ws.checkStatusTimer()
	return ws
}

//连接
func (ws *WsCon) connect() {
	dialer := websocket.DefaultDialer
	if ws.ProxyUrl != "" {
		proxy, err := url.Parse(ws.ProxyUrl)
		if err != nil {
			fmt.Println("代理地址为:", proxy)
			dialer.Proxy = http.ProxyURL(proxy)
		} else {
			fmt.Println("代理地址错误:", err)
		}
	}

	wsConn, resp, err := dialer.Dial(ws.WsUrl, http.Header(ws.RequestHeaders))
	if err != nil {
		panic(err)
	}

	ws.Conn = wsConn
	if ws.IsDump {
		dumpData, _ := httputil.DumpResponse(resp, true)
		fmt.Println(string(dumpData))
	}
	ws.UpdateActiveTime()
}

//发送文本消息数据
func (ws *WsCon) SendTextMessage(data []byte) error {
	ws.mu <- struct{}{}
	defer func() {
		<-ws.mu
	}()
	return ws.WriteMessage(websocket.TextMessage, data)
}

//发送JSON消息数据
func (ws *WsCon) SendJsonMessage(json interface{}) error {
	ws.mu <- struct{}{}
	defer func() {
		<-ws.mu
	}()
	return ws.WriteJSON(json)
}

//重连事件
func (ws *WsCon) Reconnect() {
	ws.Lock()
	defer ws.Unlock()

	log.Println("websocket基类关闭失败:", ws.Close())
	time.Sleep(time.Second)

	ws.connect()
	//重新订阅
	for _, sub := range ws.subs {
		log.Println("订阅频道:", sub)
		ws.SendJsonMessage(sub)
	}
}

//更新当前活跃时间
func (ws *WsCon) UpdateActiveTime() {
	ws.activeTimeL.Lock()
	defer ws.activeTimeL.Unlock()
	ws.activeTime = time.Now()
}

//接收处理数据
func (ws *WsCon) ReceiveMessage() {
	ws.clearChannel(ws.closeRecv)

	//开启routine
	go func() {
		for {
			if len(ws.closeRecv) > 0 {
				<-ws.closeRecv
				log.Println("关闭websocket基类连接，退出ReceiveMessage routine")
				return
			}

			t, msg, err := ws.ReadMessage()
			if err != nil {
				ws.ErrorHandleFunc(err)
				time.Sleep(time.Second)
				continue
			}

			switch t {
			case websocket.TextMessage:
				ws.ProtoHandleFunc(msg)
			case websocket.BinaryMessage:
				if ws.UnCompressFunc == nil {
					ws.ProtoHandleFunc(msg)
				} else {
					msg1, err := ws.UnCompressFunc(msg)
					if err != nil {
						ws.ErrorHandleFunc(fmt.Errorf("%s,%s", "消息解压失败", err.Error()))
					} else {
						err := ws.ProtoHandleFunc(msg1)
						if err != nil {
							ws.ErrorHandleFunc(err)
						}
					}
				}
			case websocket.CloseMessage:
				ws.Close()
				return
			default:
				log.Println("消息接收错误，消息为:", string(msg))

			}
		}
	}()
}

//关闭websocket连接
func (ws *WsCon) closeWs() {
	//清理消息通道数据
	ws.clearChannel(ws.closeCheck)
	ws.clearChannel(ws.closeReconnect)
	ws.clearChannel(ws.closeHeartbeat)
	ws.clearChannel(ws.closeRecv)

	ws.closeReconnect <- struct{}{}
	ws.closeHeartbeat <- struct{}{}
	ws.closeRecv <- struct{}{}
	ws.closeCheck <- struct{}{}

	err := ws.Close()
	if err != nil {
		log.Println("websocket基类关闭失败", err)
	}
}

//清除通道数据
func (ws *WsCon) clearChannel(c chan struct{}) {
	for {
		if len(c) > 0 {
			<-c
		} else {
			break
		}
	}
}

//心跳时间周期
func (ws *WsCon) HeartbeatTimer() {
	log.Println("心跳时间周期为:", ws.HeartbeatIntervalTime)
	if ws.HeartbeatIntervalTime == 0 || (ws.HeartbeatFunc == nil && ws.HeartbeatData == nil) {
		return
	}

	timer := time.NewTicker(ws.HeartbeatIntervalTime)
	//开启 routine
	go func() {
		ws.clearChannel(ws.closeHeartbeat)
		for {
			select {
			case <-timer.C:
				var err error
				if ws.HeartbeatFunc != nil {
					err = ws.SendJsonMessage(ws.HeartbeatFunc)
				} else {
					err = ws.SendTextMessage(ws.HeartbeatData)
				}
				if err != nil {
					log.Println("心跳数据发送异常:", err)
					time.Sleep(time.Second)
				}
			case <-ws.closeHeartbeat:
				timer.Stop()
				log.Println("关闭websocket基类连接，退出心跳routine")
				return
			}
		}
	}()
}

//状态检测
func (ws *WsCon) checkStatusTimer() {
	if ws.HeartbeatIntervalTime == 0 {
		return
	}
	timer := time.NewTimer(ws.HeartbeatIntervalTime)

	go func() {
		select {
		case <-timer.C:
			now := time.Now()
			if now.Sub(ws.activeTime) >= 2*ws.HeartbeatIntervalTime {
				log.Println("上一次活跃时间为:[", ws.activeTime, "]，已经过期，开始重新连接")
				ws.Reconnect()
			}
			timer.Reset(ws.HeartbeatIntervalTime)
		case <-ws.closeCheck:
			log.Println("退出状态检测事件")
			return
		}
	}()
}

//重连事件周期检测
func (ws *WsCon) ReconnectTimer() {
	if ws.ReconnectIntervalTime == 0 {
		return
	}

	timer := time.NewTimer(ws.ReconnectIntervalTime)

	go func() {
		ws.clearChannel(ws.closeReconnect)
		for {
			select {
			case <-timer.C:
				log.Println("websocket基类开始重连")
				ws.Reconnect()
				timer.Reset(ws.ReconnectIntervalTime)
			case <-ws.closeReconnect:
				timer.Stop()
				log.Println("websocket基类关闭，退出重连routine")
				return
			}
		}
	}()
}

//订阅
func (ws *WsCon) Subscribe(subEvent interface{}) error {
	err := ws.SendJsonMessage(subEvent)
	if err != nil {
		return err
	}
	ws.subs = append(ws.subs, subEvent)
	return nil
}
