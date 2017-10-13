package middleware

import (
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

// GenerateUnaryClientInterceptor creates an interceptor function that will log unary grpc calls.
func GenerateUnaryClientInterceptor(trace bool) grpc.UnaryClientInterceptor {
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
			message(req.(proto.Message)).
			duration(dt).
			log(true)
		return err
	}
}

// LoggingClientStream implements a GRPC client stream that logs output
type LoggingClientStream struct {
	grpc.ClientStream
	requestType string
	method      string
	trace       bool
}

func newLoggingClientStream(c *grpc.ClientStream, requestType string, request string, trace bool) *LoggingClientStream {
	return &LoggingClientStream{*c, requestType, request, trace}
}

// RecvMsg logs messages recieved over a GRPC stream
func (c *LoggingClientStream) RecvMsg(m interface{}) error {
	err := c.ClientStream.RecvMsg(m)
	if err != nil {
		return err
	}

	if c.trace {
		newRequestLogger().
			requestType(c.requestType).
			request(c.method).
			message(m.(proto.Message)).
			log(true)
	} else {
		newRequestLogger().
			requestType(c.requestType).
			request(c.method).
			log(true)
	}
	return err
}

// SendMsg logs messages sent out over a GRPC stream
func (c *LoggingClientStream) SendMsg(m interface{}) error {
	if c.trace {
		newRequestLogger().
			requestType(c.requestType).
			request(c.method).
			message(m.(proto.Message)).
			log(true)
	} else {
		newRequestLogger().
			requestType(c.requestType).
			request(c.method).
			log(true)
	}
	return c.ClientStream.SendMsg(m)
}

// GenerateStreamClientInterceptor creates an interceptor function that will log grpc streaming calls.
func GenerateStreamClientInterceptor(trace bool) grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		clientStream, err := streamer(ctx, desc, cc, method, opts...)
		loggingClientStream := newLoggingClientStream(&clientStream, "GRPC.STREAM_CLIENT", method, trace)
		if err != nil {
			err = errors.Wrap(err, "stream create call failed")
		}
		return loggingClientStream, err
	}
}
