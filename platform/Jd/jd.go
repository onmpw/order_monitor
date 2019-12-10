package Jd

import (
	"log"
	"monitor/Tool"
	"monitor/monitor"
	"monitor/monitor/model"
)

var jdChan = make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "京东", PlatformKey:"JD",
}

// 获取京东的原始数据
func getJdOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime
	// 计算时间
	myT.CalculateTime()

	var trades []*OrderTrade

	num, _ := model.Read(new(OrderTrade)).Filter("modified",">=",myT.Start).Filter("modified","<=",myT.End).GetAll(&trades)
	var jdJdp monitor.Jdp

	oriChannel := make(chan monitor.Jdp)

	go func() {
		for i:=0; i < int(num); i++{
			Tool.SetOrder(&jdJdp,trades[i])
			oriChannel <- jdJdp
		}
		jdChan <- 1
	}()

	return oriChannel, nil
}

func ParseJd() {
	oriChan, err := getJdOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order, oriChan, jdChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
