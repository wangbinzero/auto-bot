package model

// Binance base response
type BinanceBaseResponse struct {
	Type      string `json:"e"` //事件类型
	EventTime int64  `json:"E"` //事件时间
	Symbol    string `json:"s"` //交易对
}

// Binance detail response
type BinanceTickerResponse struct {
	BinanceBaseResponse
	BinanceTicker
}

// Binance kline response
type BinanceKlineResponse struct {
	BinanceBaseResponse
	Kline BinanceKline `json:"k"`
}

// Binance depth response
type BinanceDepthResponse struct {
	BinanceBaseResponse
	BinanceDepth
}

type BinanceDepth struct {
	FirstUpdateID int        `json:"U"` //从上次推送至今新增的第一个 updateID
	LastUpdateID  int        `json:"u"` //从上次推送至今新增的最后一个 updateID
	Bids          [][]string `json:"b"` //变动的买单深度
	Asks          [][]string `json:"a"` //变动的卖单深度
}

type BinanceTicker struct {
	Close  string `json:"c"` //最新成交价格
	Open   string `json:"o"` //24小时前开始第一笔成交价格
	High   string `json:"h"` //24小时内最高成交价
	Low    string `json:"l"` //24小时内最低成交价
	Volume string `json:"v"` //成交量
	Amount string `json:"1"` //成交额
}

type BinanceKline struct {
	StartTime int64  `json:"t"` //这根K线起始时间
	EndTime   int64  `json:"T"` //这根K线结束时间
	Symbol    string `json:"s"` //交易对
	Interval  string `json:"i"` //K线时间间隔
	FirstID   int    `json:"f"` //这根K线第一笔成交ID
	LastID    int    `json:"L"` //这根K线最后一笔成交ID
	Open      string `json:"o"` //这根K线第一笔成交价
	Close     string `json:"c"` //这根K线最后一笔成交价
	High      string `json:"h"` //这根K线最高成交价
	Low       string `json:"l"` //这根K线最低成交价
	Volume    string `json:"v"` //这根K线成交量
	Number    int    `json:"n"` //这根K线成交数量
	End       bool   `json:"x"` //这根K线是否完结
	Amount    string `json:"q"` //这根K线成交额
	V         string `json:"V"` //主动买入成交量
	Q         string `json:"Q"` //主动买入成交额
	B         string `json:"B"` //忽略参数
}
