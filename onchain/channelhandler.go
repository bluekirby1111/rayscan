package onchain

import (
	"fmt"
	"time"
)

type ChHandler struct {
	channels []chan *PairInfo
}

// TODO: add redis connection to dispatch pair info
func NewChHandler(channels []chan *PairInfo) *ChHandler {
	return &ChHandler{
		channels: channels,
	}
}

func (c *ChHandler) Channels() []chan *PairInfo {
	return c.channels
}

func (a *ChHandler) Start() {
	fmt.Printf("[%v] ChHandler: starting...\n", time.Now().Format("2006-01-02 15:04:05.000"))

	go func() {
		for _, ch := range a.channels {
			value := <-ch
			fmt.Printf("Pair %s", value.TokenInfo.Address)
		}
	}()
}
