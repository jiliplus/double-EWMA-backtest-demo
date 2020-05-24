package main

import (
	"context"
	"fmt"
	"time"

	"github.com/jujili/exch"
	"github.com/jujili/ta"
)

// strategyService 会从 pubsub 的
//   bar 话题中获取最新的交易数据。
//   "tick" 话题中获取最新的成交数据。
//   "balance" 话题中获取即时更新的账户信息。
// 在下达新订单的时候，会先在 "cancelAllOrders" 里面发出通知，取消所有的未成交订单。
// 然后再利用 "order" 话题下单。
func strategyService(ctx context.Context, ps exch.Pubsub, interval time.Duration) {
	topic := fmt.Sprintf("%sBar", interval)
	bars, err := ps.Subscribe(ctx, topic)
	if err != nil {
		panic(err)
	}
	decBar := exch.DecBarFunc()
	//
	go func() {
		short := ta.NewEWMA(10)
		long := ta.NewEWMA(30)
		for {
			select {
			case msg := <-bars:
				bar := decBar(msg.Payload)
				msg.Ack()
				short.Update(bar.Close)
				long.Update(bar.Close)
			}
		}

	}()
}

func cancelAllOrders(pub exch.Publisher) {
	pub("cancelAllOrders", nil)
}
