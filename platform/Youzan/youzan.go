package Youzan

import (
	"github.com/onmpw/JYGO/model"
	"log"
	"monitor/Tool"
	"monitor/monitor"
)

var yzChan      = make(chan int, 1)

var Order = monitor.MyOrderInfo{
	Platform: "有赞", PlatformKey:"YZ",
}

// 获取有赞的原始数据
func getYouZanOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime

	myT.CalculateTime()

	var trades []*OrderTrade

	num, _ := model.Read(new(OrderTrade)).Filter("modified",">=",myT.Start).Filter("modified","<=",myT.End).GetAll(&trades)

	var youzanJdp monitor.Jdp

	oriChannel := make(chan monitor.Jdp)

	go func() {
		for i:=0; i < int(num); i++{
			Tool.SetOrder(&youzanJdp,trades[i])

			oriChannel <- youzanJdp
		}
		yzChan <- 1
	}()
	return oriChannel, nil
}

func ParseYouZan() {
	oriCha, err := getYouZanOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order, oriCha, yzChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
