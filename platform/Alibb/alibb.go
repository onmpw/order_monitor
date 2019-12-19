package Alibb

import (
	"github.com/onmpw/JYGO/model"
	"log"
	"monitor/Tool"
	"monitor/monitor"
)


var alibbChan	= make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "1688", PlatformKey:"1688",
}


// 获取1688的原始数据
func getAlibbOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime

	myT.CalculateTime()

	var trades []*OrderTrade

	num, _ := model.Read(new(OrderTrade)).Filter("modified",">=",myT.Start).Filter("modified","<=",myT.End).GetAll(&trades)

	var alibbJdp monitor.Jdp

	oriChannel := make(chan monitor.Jdp)

	go func(){
		for i:=0; i < int(num); i++{
			Tool.SetOrder(&alibbJdp,trades[i])
			oriChannel <- alibbJdp
		}
		alibbChan <- 1
	}()
	return oriChannel,nil
}

func ParseAlibb() {
	oriCha, err := getAlibbOriginData()
	if err != nil {
		log.Panic(err.Error())
	}
	go func() {
		err := Tool.CheckSync(&Order,oriCha, alibbChan)
		if err != nil {
			log.Panic(err.Error())
		}
	}()
}
