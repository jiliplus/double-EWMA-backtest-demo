package main

import (
	"context"
	"log"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
	"github.com/jujili/exch"
)

func main() {
	// srcName := "./btcusdt.sqlite3"
	srcName := "./binance.sqlite3"
	//
	db := openToMemory(srcName)
	defer db.Close()
	//
	// tickSrc(db, nil)
	//
	pubSub := gochannel.NewGoChannel(
		gochannel.Config{},
		watermill.NewStdLogger(false, false),
	)

	go func() {
		decTick := exch.DecTickFunc()
		ticks, err := pubSub.Subscribe(context.TODO(), "tick")
		if err != nil {
			panic(err)
		}
		for msg := range ticks {
			tick := decTick(msg.Payload)
			msg.Ack()
			log.Println(tick)
		}
	}()

	tickPublishService(context.TODO(), pubSub, db)
}
