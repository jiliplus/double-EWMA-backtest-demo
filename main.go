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
			log.Println("in main go for", tick)
		}
		wg.Done()
	}()

	interval := time.Hour

	// 启动策略
	log.Println("准备启动 策略")
	strategyService(context.TODO(), pubSub, interval, "BTCUSDT", "BTC", "USDT")
	log.Println("完成启动 策略")

	// 启动帐户服务
	log.Println("准备启动 帐户记录服务")
	prices := make(map[string]float64, 2)

	asset, capital := "BTC", "USDT"
	prices[capital] = 1
	backtest.BalanceService(context.TODO(), pubSub, prices, asset)
	log.Println("完成启动 帐户记录服务")

	// 启动 backtest 交易所
	usdt := exch.NewAsset("USDT", 10000, 0)
	bal := exch.NewBalances(usdt)
	log.Println("初始化的 Balance", bal)
	backtest.NewBackTest(context.TODO(), pubSub, bal)

	// 启动 TickBarService
	backtest.TickBarService(context.TODO(), pubSub, interval)

	srcName := "./btcusdt.sqlite3"
	// srcName := "./binance.sqlite3"
	//
	db := openToMemory(srcName)
	defer db.Close()
	//
	tickPublishService(context.TODO(), pubSub, db)
	//
	wg.Wait()
}
