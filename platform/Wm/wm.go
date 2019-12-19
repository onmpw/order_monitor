package Wm

import (
	"github.com/onmpw/JYGO/model"
	"log"
	"monitor/Tool"
	"monitor/monitor"
)

var wmChan  = make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "微盟", PlatformKey:"WM",
}

// 获取微盟的原始数据
func getWmOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime

	myT.CalculateTime()

	var trades []*OrderTrade

	num, _ := model.Read(new(OrderTrade)).Filter("modified",">=",myT.Start).Filter("modified","<=",myT.End).GetAll(&trades)

	var wmJdp monitor.Jdp

	oriChannel := make(chan monitor.Jdp)

	go func() {
		for i:=0; i < int(num); i++{
			Tool.SetOrder(&wmJdp,trades[i])

			oriChannel <- wmJdp
		}
		wmChan <- 1
	}()

	return oriChannel, nil
}

func ParseWm() {
	oriChan, err := getWmOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order, oriChan, wmChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
