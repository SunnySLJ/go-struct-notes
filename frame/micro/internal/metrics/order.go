package metrics

import "github.com/prometheus/client_golang/prometheus"

func init() {
	prometheus.MustRegister(OrderList)
}

/**
** NewCounterVec 表示这个Counter是一个向量，包括了两块 - opts和labels
   opts包括Name和Help，Name是metrics唯一的名称，Help是metrics的帮助信息
   labels是用来过滤、聚合功能的关键参数，提前声明有利于存储端进行优化（可类比数据库索引）
*/
var OrderList = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Name: "order_list_counter",
		Help: "List Order Count",
	},
	[]string{"service"},
)
