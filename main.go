package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bluekirby1111/rayscan/config"
	"github.com/bluekirby1111/rayscan/connection"
	"github.com/bluekirby1111/rayscan/onchain"
)

func main() {
	var publisherAddr string = ""
	flag.StringVar(&publisherAddr, "publisher-addr", "", "AMPQ connection string")
	flag.Parse()

	fmt.Printf("Starting with args: publisherAddr=%s \n", publisherAddr)

	//TODO: move config to args
	cfg, err := config.LoadConfig(config.DefaultConfigPath)
	if err != nil {
		fmt.Printf("Error loading config: %s\n", err)
		os.Exit(1)
	}

	rpcPool, err := connection.NewRPCClientPool(cfg.Nodes)
	if err != nil {
		fmt.Printf("Error creating rpc pool: %s\n", err)
		os.Exit(1)
	}
	defer rpcPool.Close()

	var pairPublishChannel []chan *onchain.PairInfo
	channel := make(chan *onchain.PairInfo, 100)
	pairPublishChannel = append(pairPublishChannel, channel)

	var pairCollector *onchain.PairCollector
	if len(publisherAddr) != 0 {
		chHandler := onchain.NewChHandler(publisherAddr, pairPublishChannel)
		chHandler.Start()

		pairCollector = onchain.NewPairCollector()
		pairCollector.Start(chHandler.Channels())
	} else {
		pairCollector = onchain.NewPairCollector()
		pairCollector.Start(nil)

	}

	txAnalyzer := onchain.NewTxAnalyzer(rpcPool)
	txAnalyzer.Start(pairCollector.Channel())

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var observers []*onchain.LogObserver
	for _, v := range rpcPool.Connections {
		if !v.ConnectionInfo.Observer {
			continue
		}

		obs := onchain.NewLogObserver(rpcPool, v.ConnectionInfo.Name)
		if err := obs.Start(ctx, txAnalyzer.Channel()); err != nil {
			fmt.Printf("Error starting %s log observer: %s\n", v.ConnectionInfo.Name, err)
			os.Exit(1)
		}

		observers = append(observers, obs)
	}

	var stopChan = make(chan os.Signal, 1)
	signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-stopChan // wait for SIGINT

	fmt.Printf("Interrupted; stopping...\n")
	ctx, cancel = context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	for _, obs := range observers {
		if err := obs.Stop(ctx); err != nil {
			fmt.Printf("Error stopping %s log observer: %s\n", obs.ConnectionName(), err)
		}
	}

	if err := txAnalyzer.Stop(ctx); err != nil {
		fmt.Printf("Error stopping tx analyzer: %s\n", err)
	}

	if err := pairCollector.Stop(ctx); err != nil {
		fmt.Printf("Error stopping pair collector: %s\n", err)
	}
}
