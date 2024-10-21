package onchain

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/wagslane/go-rabbitmq"
)

type ChHandler struct {
	ReceiverChannels []chan *PairInfo
	Publisher        *rabbitmq.Publisher
}

func NewChHandler(publisherAddr string, channels []chan *PairInfo) *ChHandler {
	conn, err := rabbitmq.NewConn(
		publisherAddr,
		rabbitmq.WithConnectionOptionsLogging,
	)
	if err != nil {
		log.Fatal(err)
	}

	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName("pairs"),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
	)
	if err != nil {
		log.Fatal(err)
	}

	return &ChHandler{
		ReceiverChannels: channels,
		Publisher:        publisher,
	}
}

func (c *ChHandler) Channels() []chan *PairInfo {
	return c.ReceiverChannels
}

func (a *ChHandler) Start() {
	fmt.Printf("[%v] ChHandler: starting...\n", time.Now().Format("2006-01-02 15:04:05.000"))

	go func() {
		for _, ch := range a.ReceiverChannels {
			value := <-ch
			jsonValue, err := json.Marshal(value)
			if err != nil {
				fmt.Println(err)
				continue
			}

			a.Publisher.Publish(
				[]byte(jsonValue),
				[]string{"rayscan"},
				rabbitmq.WithPublishOptionsContentType("application/json"),
				rabbitmq.WithPublishOptionsExchange("events"),
			)

		}
	}()
}
