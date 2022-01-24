package main

import (
	"context"
	"flag"
	"fmt"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/uber/jaeger-client-go"
	jaegerconfig "github.com/uber/jaeger-client-go/config"
	"go-micro/gen/proto/order"
	"go-micro/internal/config"
	"go-micro/internal/logx"
	"go-micro/internal/mysql"
	"go-micro/internal/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
	"net/http"
)

func main() {
	var configFilePath = flag.String("c", "./", "config file path")
	flag.Parse()

	if err := config.Load(*configFilePath); err != nil {
		panic(err)
	}

	logx.Init(config.Viper.GetString("log.path"))
	defer logx.Sync()
	logx.Sugar.Info("server is running")

	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		http.ListenAndServe(fmt.Sprintf(":%d", config.Viper.GetInt("server.prometheus.port")), mux)
	}()

	traceCfg := &jaegerconfig.Configuration{
		ServiceName: "MyService",
		Sampler: &jaegerconfig.SamplerConfig{
			Type:  jaeger.SamplerTypeConst,
			Param: 1,
		},
		Reporter: &jaegerconfig.ReporterConfig{
			LocalAgentHostPort: "127.0.0.1:6831",
			LogSpans:           true,
		},
	}
	tracer, closer, err := traceCfg.NewTracer(jaegerconfig.Logger(jaeger.StdLogger))
	if err != nil {
		panic(err)
	}
	defer closer.Close()
	opentracing.SetGlobalTracer(tracer)

	// mysql初始化失败的话，不要继续运行程序
	if err := mysql.Init(config.Viper.GetString("mysql.user"),
		config.Viper.GetString("mysql.password"),
		config.Viper.GetString("mysql.ip"),
		config.Viper.GetInt("mysql.port"),
		config.Viper.GetString("mysql.dbname")); err != nil {
		logx.Sugar.Fatalf("init mysql error %v", err)
	}

	go func() {
		lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Viper.GetInt("server.grpc.port")))
		if err != nil {
			panic(err)
		}

		s := grpc.NewServer(
			grpc.ChainUnaryInterceptor(
				grpc_opentracing.UnaryServerInterceptor(
					grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
				),
				ServerValidationUnaryInterceptor,
			),
		)
		order.RegisterOrderServiceServer(s, &server.Server{})

		if err = s.Serve(lis); err != nil {
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

	// Register gRPC server endpoint
	// Note: Make sure the gRPC server is running properly and accessible
	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(
			grpc_opentracing.UnaryClientInterceptor(
				grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
			),
		),
	}

	if err := order.RegisterOrderServiceHandlerFromEndpoint(ctx, mux, fmt.Sprintf(":%d", config.Viper.GetInt("server.grpc.port")), opts); err != nil {
		return errors.Wrap(err, "RegisterOrderServiceHandlerFromEndpoint error")
	}

	// Start HTTP server (and proxy calls to gRPC server endpoint)
	return http.ListenAndServe(":8081", tracingWrapper(mux))
}

// ValidateAll 对应 protoc-gen-validate 生成的 *.pb.validate.go 中的代码
type Validator interface {
	ValidateAll() error
}

func ServerValidationUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	logx.Sugar.Infof("%+v", req)
	if r, ok := req.(Validator); ok {
		if err := r.ValidateAll(); err != nil {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
	}

	return handler(ctx, req)
}

var grpcGatewayTag = opentracing.Tag{Key: string(ext.Component), Value: "grpc-gateway"}

func tracingWrapper(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		parentSpanContext, err := opentracing.GlobalTracer().Extract(
			opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(r.Header))
		if err == nil || err == opentracing.ErrSpanContextNotFound {
			serverSpan := opentracing.GlobalTracer().StartSpan(
				"ServeHTTP",
				// this is magical, it attaches the new span to the parent parentSpanContext, and creates an unparented one if empty.
				ext.RPCServerOption(parentSpanContext),
				grpcGatewayTag,
			)
			r = r.WithContext(opentracing.ContextWithSpan(r.Context(), serverSpan))

			trace, ok := serverSpan.Context().(jaeger.SpanContext)
			if ok {
				w.Header().Set(jaeger.TraceContextHeaderName, fmt.Sprint(trace.TraceID()))
			}

			defer serverSpan.Finish()
		}
		h.ServeHTTP(w, r)
	})
}
