// Code generated by gormer. DO NOT EDIT.
package model

import "time"

// Table Level Info
const OrderTableName = "orders"

// Field Level Info
type OrderField string

const (
	OrderFieldId           OrderField = "id"
	OrderFieldName         OrderField = "name"
	OrderFieldPrice        OrderField = "price"
	OrderFieldCreateTime   OrderField = "create_time"
	OrderFieldUpdateTime   OrderField = "update_time"
	OrderFieldDeleteStatus OrderField = "delete_status"
)

var OrderFieldAll = []OrderField{"id", "name", "price", "create_time", "update_time", "delete_status"}

// Kernel struct for table for one row
type Order struct {
	Id           int64     `gorm:"column:id"`            // 主键
	Name         string    `gorm:"column:name"`          // 名称，建议唯一
	Price        float64   `gorm:"column:price"`         // 订单价格
	CreateTime   time.Time `gorm:"column:create_time"`   // 创建时间
	UpdateTime   time.Time `gorm:"column:update_time"`   // 更新时间
	DeleteStatus int       `gorm:"column:delete_status"` // 删除状态，1表示软删除
}

// Kernel struct for table operation
type OrderOptions struct {
	Order  *Order
	Fields []string
}

// Match: case insensitive
var ordersFieldMap = map[string]string{
	"Id": "id", "id": "id",
	"Name": "name", "name": "name",
	"Price": "price", "price": "price",
	"CreateTime": "create_time", "create_time": "create_time",
	"UpdateTime": "update_time", "update_time": "update_time",
	"DeleteStatus": "delete_status", "delete_status": "delete_status",
}

func NewOrderOptions(target *Order, fields ...OrderField) *OrderOptions {
	options := &OrderOptions{
		Order:  target,
		Fields: make([]string, len(fields)),
	}
	for index, field := range fields {
		options.Fields[index] = string(field)
	}
	return options
}

func NewOrderOptionsAll(target *Order) *OrderOptions {
	return NewOrderOptions(target, OrderFieldAll...)
}

func NewOrderOptionsRawString(target *Order, fields ...string) *OrderOptions {
	options := &OrderOptions{
		Order: target,
	}
	for _, field := range fields {
		if f, ok := ordersFieldMap[field]; ok {
			options.Fields = append(options.Fields, f)
		}
	}
	return options
}
