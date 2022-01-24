package main

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pkg/errors"
	"go-frame/examples/proto/order"
	"go-frame/examples/registry/etcd/server"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
)

func main() {
	go func() {
		l, err := net.Listen("tcp", ":9091")
		if err != nil {
			log.Panic(err)
		}
		s := grpc.NewServer()
		order.RegisterOrderServiceServer(s, &server.Server{})

		//client, err := etcdclient.New(etcdclient.Config{
		//	Endpoints: []string{"127.0.0.1:2379"},
		//})
		//if err != nil {
		//	log.Fatal(err)
		//}

		//r := etcd.New(client)
		//app := kratos.New(
		//	kratos.Name("helloworld"),
		//	kratos.Server(
		//		httpSrv,
		//		grpcSrv,
		//	),
		//	kratos.Registrar(r),
		//)

		if err = s.Serve(l); err != nil {
			panic(err)
		}
	}()

	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := order.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, ":8082", opts); err != nil {
		return errors.Wrap(err, "RegisterDemoServiceHandlerFromEndpoint error")
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(":8081", mux)
}
