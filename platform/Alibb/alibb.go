package Alibb

import (
	"log"
	"monitor/Tool"
	"monitor/monitor"
	"monitor/monitor/model"
)


var alibbChan	= make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "1688", PlatformKey:"1688",
}


// 获取1688的原始数据
func getAlibbOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime

	myT.CalculateTime()
	//var err error

	var trades []*OrderTrade

	num, _ := model.Read(new(OrderTrade)).Filter("modified",">=",myT.Start).Filter("modified","<=",myT.End).GetAll(&trades)



	/*where := []interface{}{
		[]interface{}{"modified",">=",myT.Start},
		[]interface{}{"modified","<=",myT.End},
	}
	fields := []string{
		"id","oid","response","cid","created","modified","type","sid",
	}
	rows := db.Db.Connector().Table("jdp_alibb_order_trade").Select(fields...).Where(where...).Get()*/

	var alibbJdp monitor.Jdp
	//var inter = monitor.RowData{&alibbJdp.Id,&alibbJdp.Oid,&alibbJdp.Response,&alibbJdp.CompanyId,&alibbJdp.Created,&alibbJdp.Modified,&alibbJdp.OrderType,&alibbJdp.ShopId}

	oriChannel := make(chan monitor.Jdp)

	go func(){
		for i:=0; i < int(num); i++{
			alibbJdp.Id = trades[i].Id
			alibbJdp.Oid = trades[i].Oid
			alibbJdp.Response = trades[i].Response
			alibbJdp.CompanyId = trades[i].Cid
			alibbJdp.Created = trades[i].Created
			alibbJdp.Modified = trades[i].Modified
			alibbJdp.OrderType = trades[i].Type
			alibbJdp.ShopId = trades[i].Sid
			oriChannel <- alibbJdp
		}
		/*for rows.Next() {
			err = rows.Scan(inter...)
			if err != nil {
				panic(err.Error())
			}
			oriChannel <- alibbJdp
		}*/
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
