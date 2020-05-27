package main

import (
	"context"
	"log"
	"sync"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/jujili/exch"
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
	wg.Add(1)
	go func() {
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

	srcName := "./btcusdt.sqlite3"
	// srcName := "./binance.sqlite3"
	//
	db := openToMemory(srcName)
	defer db.Close()
	//
	tickPublishService(context.TODO(), pubSub, db)
	wg.Wait()
}
