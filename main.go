package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/jujili/exch"
	"github.com/jujili/exch/backtest"
)

func main() {
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{
			// 没有下面这一行，会无法从 Payload 的数据是乱的
			BlockPublishUntilSubscriberAck: true,
		},
		watermill.NewStdLogger(false, false),
	)

	var wg sync.WaitGroup

	go func() {
		wg.Add(1)
		ticks, err := pubSub.Subscribe(context.TODO(), "tick")
		if err != nil {
			panic(err)
		}
		decTick := exch.DecTickFunc()
		for msg := range ticks {
			tick := decTick(msg.Payload)
			msg.Ack()
			log.Println("dst", tick)
		}
		wg.Done()
	}()

	// 启动策略
	strategyService(context.TODO(), pubSub, time.Hour, "BTCUSDT", "BTC", "USDT")

	// 启动帐户服务
	prices := make(map[string]float64, 2)
	asset := "USDT"
	prices[asset] = 1
	backtest.BalanceService(context.TODO(), pubSub, prices, asset)

	// 启动 backtest 交易所
	// exch.NewBacktest(context.TODO(), pubSub)

	// srcName := "./btcusdt.sqlite3"
	srcName := "./binance.sqlite3"
	//
	db := openToMemory(srcName)
	defer db.Close()
	//
	tickPublishService(context.TODO(), pubSub, db)
	//
	wg.Wait()
}
