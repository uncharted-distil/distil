package middleware

import (
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// GenerateUnaryClientInterceptor creates an interceptor function that will log unary grpc calls.
func GenerateUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		startTime := time.Now()
		err := invoker(ctx, method, req, reply, cc, opts...)
		if err != nil {
			err = errors.Wrap(err, "invoker call failed")
		}
		dt := time.Since(startTime)
		newRequestLogger().
			requestType("GRPC.UNARY").
			request(method).
			duration(dt).
			log(true)
		return err
	}
}

// GenerateStreamClientInterceptor creates an interceptor function that will log grpc streaming calls.
func GenerateStreamClientInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		startTime := time.Now()
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		if err != nil {
			err = errors.Wrap(err, "stream call failed")
		}
		dt := time.Since(startTime)
		newRequestLogger().
			requestType("GRPC.UNARY").
			request(method).
			duration(dt).
			log(true)
		return clientStream, err
	}
}
