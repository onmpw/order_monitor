package Tool

import (
	"fmt"
	"monitor/monitor"
	"monitor/monitor/db"
)

func CheckSync(order *monitor.MyOrderInfo, oriChan <-chan monitor.Jdp, quit <-chan int) error {

	var count, unusualCount int64
	var unusualType, response, remarks string
	var err error

	for {
		select {
		case orderInfo := <-oriChan:
			order.TotalCount++ // 记录订单总数

			var succeedOrder monitor.SucceedOrder
			var failedOrder monitor.BadOrder // 临时存储失败订单信息
			count,err = db.Db.Connector().Table("order_info").Where([]interface{}{[]interface{}{"company_id",orderInfo.CompanyId},[]interface{}{"number",orderInfo.Oid},[]interface{}{"source","!=","System"}}...).Count()

			if err != nil {
				panic(err.Error()) // proper error handling instead of panic in your app
			}

			monitor.SafeCompanyOrder.Lock()
			if _, ok := monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId]; !ok {
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId] = &monitor.MyOrderInfo{}
			}
			monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].TotalCount++
			monitor.SafeCompanyOrder.Unlock()
			if count == 0 {
				order.FailedCount++ //  记录失败数量

				// 记录失败订单的信息
				failedOrder.Oid = orderInfo.Oid
				failedOrder.ShopId = orderInfo.ShopId
				failedOrder.CompanyId = orderInfo.CompanyId
				failedOrder.Created = orderInfo.Created
				failedOrder.Modified = orderInfo.Modified
				failedOrder.Response = orderInfo.Response
				failedOrder.OrderType = orderInfo.OrderType

				unusualRow := db.Db.Connector().Table("order_sync_unusual").Select("count(*)","type","response","remarks").Where([]interface{}{[]interface{}{"tid",orderInfo.Oid},[]interface{}{"is_delete","N"}}...).GetOne()
				_ = unusualRow.Scan(&unusualCount, &unusualType, &response, &remarks)
				if unusualCount == 0 {
					failedOrder.Reason = " 订单同步未成功，没有进入异常表，具体原因有待查明！" //订单失败原因
				} else {
					failedOrder.Reason = " 订单同步异常，异常类型：" + unusualType + " 原因：" + remarks // 同步订单异常
				}

				fmt.Printf("%s %c[31;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步失败", 0x1B)
				order.FailedOrder = append(order.FailedOrder, failedOrder)
				monitor.SafeCompanyOrder.Lock()
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedCount++
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder = append(monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder, failedOrder)
				monitor.SafeCompanyOrder.Unlock()
			} else if count > 1 {
				order.FailedCount++ // 订单同步成功，但是重复了，所以标记为失败
				/*monitor.SafeCompanyOrder.Lock()
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedCount++
				monitor.SafeCompanyOrder.Unlock()*/

				// 加入失败订单，提示出来
				failedOrder.Oid = orderInfo.Oid
				failedOrder.ShopId = orderInfo.ShopId
				failedOrder.CompanyId = orderInfo.CompanyId
				failedOrder.Created = orderInfo.Created
				failedOrder.Modified = orderInfo.Modified
				failedOrder.Response = orderInfo.Response
				failedOrder.OrderType = orderInfo.OrderType

				failedOrder.Reason = " 订单同步成功，但是重复进入系统表，需要删除其中的一条！" //订单失败原因

				fmt.Printf("%s %c[32;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步成功", 0x1B)
				order.FailedOrder = append(order.FailedOrder, failedOrder)
				monitor.SafeCompanyOrder.Lock()
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedCount++
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder = append(monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].FailedOrder, failedOrder)
				monitor.SafeCompanyOrder.Unlock()

			} else {
				succeedOrder.Oid = orderInfo.Oid
				succeedOrder.ShopId = orderInfo.ShopId
				succeedOrder.CompanyId = orderInfo.CompanyId
				succeedOrder.Created = orderInfo.Created
				succeedOrder.Modified = orderInfo.Modified
				succeedOrder.Response = orderInfo.Response
				succeedOrder.OrderType = orderInfo.OrderType
				succeedOrder.Reason = "成功!"

				order.SucceedCount++ // 记录成功数量
				fmt.Printf("%s %c[32;40;1m%s%c[0m \n", orderInfo.Oid, 0x1B, "订单同步成功", 0x1B)
				monitor.SafeCompanyOrder.Lock()
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedCount ++
				monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedOrder = append(monitor.SafeCompanyOrder.CompanyOrder[orderInfo.ShopId].SucceedOrder, succeedOrder)
				monitor.SafeCompanyOrder.Unlock()

			}
		case <-quit:
			monitor.C <- 1
			return nil
		}
	}
}

func Close(){

}
