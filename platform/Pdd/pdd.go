package Pdd

import (
	"github.com/onmpw/JYGO/model"
	"log"
	"monitor/Tool"
	"monitor/monitor"
)

var pddChan = make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "拼多多", PlatformKey:"PDD",
}

// 获取拼多多的原始数据
func getPddOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime

	myT.CalculateTime()

	var trades []*OrderTrade

	num, _ := model.Read(new(OrderTrade)).Filter("modified",">=",myT.Start).Filter("modified","<=",myT.End).GetAll(&trades)
	var pddJdp monitor.Jdp

	oriChannel := make(chan monitor.Jdp)

	go func() {
		for i:=0; i < int(num); i++{
			Tool.SetOrder(&pddJdp,trades[i])

			oriChannel <- pddJdp
		}
		pddChan <- 1
	}()
	return oriChannel, nil
}

func ParsePdd() {
	oriChan, err := getPddOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order, oriChan, pddChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
