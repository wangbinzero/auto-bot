package middleware

import (
	"fmt"
	"github.com/streadway/amqp"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"path/filepath"
	"time"
)

type Amqp struct {
	Connect struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		UserName string `yaml:"username"`
		Password string `yaml:"password"`
	} `yaml:"connect"`

	Exchange struct {
		Default map[string]string `yaml:"default"`
		Kline   map[string]string `yaml:"kline"`
		Ticker  map[string]string `yaml:"ticker"`
	} `yaml:"exchange"`

	Queue struct {
		Kline  map[string]string `yaml:"kline"`
		Ticker map[string]string `yaml:"ticker"`
	} `yaml:"queue"`
}

var (
	AmqpGlobalConfig Amqp
	RabbitMqConnect  *amqp.Connection
)

// initialize RabbitMQ config
func InitAmqpConfig() {
	path, _ := filepath.Abs("./config/amqp.yaml")
	content, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
		return
	}
	err = yaml.Unmarshal(content, &AmqpGlobalConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	InitAmqpConn()
}

// initialize RabbitMQ connection
func InitAmqpConn() {
	var err error
	RabbitMqConnect, err = amqp.Dial("amqp://" +
		AmqpGlobalConfig.Connect.UserName + ":" +
		AmqpGlobalConfig.Connect.Password + "@" +
		AmqpGlobalConfig.Connect.Host + ":" +
		AmqpGlobalConfig.Connect.Port + "/")

	if err != nil {
		fmt.Println(AmqpGlobalConfig)
		fmt.Println(err)
		time.Sleep(5000)
		InitAmqpConn()
		return
	}

	go func() {
		<-RabbitMqConnect.NotifyClose(make(chan *amqp.Error))
		InitAmqpConn()
	}()

	//declare exchange
	DeclareExchange()
}

// publish message to RabbitMQ
func PublishMessageWithRouteKey(exchange, routeKey, contentType string,
	message *[]byte, arguments amqp.Table, deliveryMode uint8) error {
	channel, err := RabbitMqConnect.Channel()
	defer channel.Close()
	if err != nil {
		return fmt.Errorf("channel: %s", err)
	}

	if err = channel.Publish(
		exchange,
		routeKey,
		false,
		false,
		amqp.Publishing{
			Headers:         amqp.Table{},
			ContentType:     contentType,
			ContentEncoding: "",
			DeliveryMode:    deliveryMode,
			Priority:        0,
			Body:            *message,
		},
	); err != nil {
		return fmt.Errorf("queue publish: %s", err)
	}
	return nil
}

// declare RabbitMQ exchange
func DeclareExchange() error {
	channel, err := RabbitMqConnect.Channel()
	if err != nil {
		return fmt.Errorf("get channel error: %s", err)
	}

	err = channel.ExchangeDeclare(AmqpGlobalConfig.Exchange.Default["key"],
		AmqpGlobalConfig.Exchange.Default["type"],
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		return fmt.Errorf("declare exchange default error: %s", err)
	}
	fmt.Printf("declare exchange [%s] success\n", AmqpGlobalConfig.Exchange.Default["key"])

	err = channel.ExchangeDeclare(AmqpGlobalConfig.Exchange.Kline["key"],
		AmqpGlobalConfig.Exchange.Kline["type"],
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		return fmt.Errorf("declare exchange kline error: %s", err)
	}

	fmt.Printf("declare exchange [%s] success\n", AmqpGlobalConfig.Exchange.Kline["key"])

	err = channel.ExchangeDeclare(AmqpGlobalConfig.Exchange.Ticker["key"],
		AmqpGlobalConfig.Exchange.Ticker["type"],
		true,
		false,
		false,
		false,
		nil)
	if err != nil {
		return fmt.Errorf("declare exchange ticker error: %s", err)
	}
	fmt.Printf("declare exchange [%s] success\n", AmqpGlobalConfig.Exchange.Ticker["key"])
	DeclareQueue(channel)
	return nil
}

func DeclareQueue(channel *amqp.Channel) error {
	var err error
	_, err = channel.QueueDeclare(AmqpGlobalConfig.Queue.Kline["key"], true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue kline error: %s", err)
	}

	fmt.Printf("declare queue [%s] success\n", AmqpGlobalConfig.Queue.Kline["key"])

	_, err = channel.QueueDeclare(AmqpGlobalConfig.Queue.Ticker["key"], true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("declare queue ticker error: %s", err)
	}
	fmt.Printf("declare queue [%s] success\n", AmqpGlobalConfig.Queue.Ticker["key"])
	BindQueue(channel)
	return nil
}

//  bind queue with exchange by routeKey
func BindQueue(channel *amqp.Channel) {
	channel.QueueBind(AmqpGlobalConfig.Queue.Kline["key"], "kline", AmqpGlobalConfig.Exchange.Kline["key"], false, nil)
	channel.QueueBind(AmqpGlobalConfig.Queue.Ticker["key"], "ticker", AmqpGlobalConfig.Exchange.Ticker["key"], false, nil)
}
