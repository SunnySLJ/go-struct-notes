package server

import (
	"go-micro/gen/proto/order"
)

type Server struct {
	// 使用unsafe可以强制让编译器检查是否实现了相关方法
	order.UnsafeOrderServiceServer
}
