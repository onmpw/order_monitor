package Alibb

import (
	"database/sql"
	"log"
	"monitor/Tool"
	"monitor/monitor"
)


var alibbChan	= make(chan int, 1)
var Order = monitor.MyOrderInfo{
	Platform: "1688", PlatformKey:"1688",
}

/**
 * 获取1688的原始数据
 */
func getAlibbOriginData() (<-chan monitor.Jdp, error) {
	var myT monitor.MyTime
	alibbDb , err := sql.Open(monitor.DriverName,monitor.DataSourceName)

	if err != nil {
		return nil,err
	}

	stmtOut , err := alibbDb.Prepare("select id,oid,response,cid,created,modified,`type`,sid from jdp_alibb_order_trade where modified>=? and modified<?")
	if err != nil {
		return nil,err
	}

	defer monitor.CloseStmt(stmtOut)

	// 计算时间
	myT.CalculateTime()

	var alibbJdp monitor.Jdp
	var inter = monitor.RowData{&alibbJdp.Id,&alibbJdp.Oid,&alibbJdp.Response,&alibbJdp.CompanyId,&alibbJdp.Created,&alibbJdp.Modified,&alibbJdp.OrderType,&alibbJdp.ShopId}

	rows, err := stmtOut.Query(myT.Start,myT.End)

	if err != nil {
		return nil,err
	}

	oriChannel := make(chan monitor.Jdp)

	go func(){
		for rows.Next() {
			err = rows.Scan(inter...)
			if err != nil {
				panic(err.Error())
			}
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
