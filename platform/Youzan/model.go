package Youzan

type OrderTrade struct {
	Id				int
	Oid				string
	Response		string
	Cid				int
	Created			string
	Modified		string
	Type 			string
	Sid 			int
}

func (o *OrderTrade) TableName() string {
	return "jdp_youzan_order_trade"
}
