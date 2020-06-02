package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/jujili/exch"
	"github.com/jujili/exch/backtest"
	"github.com/jujili/ta"
)

// strategyService 会从 pubsub 的
//   bar 话题中获取最新的交易数据。
//   "tick" 话题中获取最新的成交数据。
//   "balance" 话题中获取即时更新的账户信息。
// 在下达新订单的时候，会先在 "cancelAllOrders" 里面发出通知，取消所有的未成交订单。
// 然后再利用 "order" 话题下单。
func strategyService(ctx context.Context, ps backtest.Pubsub, interval time.Duration, symbol, asset, capital string) {
	log.Println("进入 策略 服务...")
	// TODO: 把 topic 封装起来
	topic := fmt.Sprintf("%sBar", interval)
	bars, err := ps.Subscribe(ctx, topic)
	if err != nil {
		panic(err)
	}
	decBar := exch.DecBarFunc()
	//
	balances, err := ps.Subscribe(ctx, "balance")
	if err != nil {
		panic(err)
	}
	decBal := exch.DecBalanceFunc()
	go func() {
		log.Println("策略服务 go func ...")
		short := ta.NewEWMA(10)
		long := ta.NewEWMA(30)
		var balance exch.Balance
		orderTamplate := exch.NewOrder(symbol, asset, capital)
		enc := exch.EncFunc()
		for {
			log.Println("策略  for 循环")
			select {
			case <-ctx.Done():
				log.Fatalln("strategy service end: ", ctx.Err())
			case msg := <-bars:
				// log.Println("策略  <-bars")
				bar := decBar(msg.Payload)
				msg.Ack()
				short.Update(bar.Close)
				long.Update(bar.Close)
				if !short.IsInited() || !long.IsInited() {
					continue
				}
				s, l := short.Value(), long.Value()
				if s > l { // 市场开始上扬
					free := balance[capital].Free
					if free > 0 {
						order := orderTamplate.With(exch.Market(exch.BUY, free))
						message := message.NewMessage(watermill.NewUUID(), enc(order))
						go ps.Publish("order", message)
						log.Println("下市价买单", order)
					}
				} else if s < l { // 市场开始下调
					free := balance[asset].Free
					if free > 0 {
						order := orderTamplate.With(exch.Market(exch.SELL, free))
						message := message.NewMessage(watermill.NewUUID(), enc(order))
						go ps.Publish("order", message)
						log.Println("下市价卖单", order)
					}
				}
			case msg := <-balances:
				// log.Println("策略  <-balances")
				balance = *decBal(msg.Payload)
				msg.Ack()
			}
		}
	}()
}

// 因为这里全部采用市价单，所以，基本不会有未成交的订单
// func cancelAllOrders(pub backtest.Publisher) {
// 	pub.Publish("cancelAllOrders", nil)
// }
