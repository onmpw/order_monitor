package Jd

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
	return "jdp_jd_order_trade"
}

func (o *OrderTrade) Connection() string {
	return "jd_production"
}
