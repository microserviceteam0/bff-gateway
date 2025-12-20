package metrics

import (
	"context"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func GRPCUnaryServerInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		GRPCActiveConnections.WithLabelValues(serviceName).Inc()
		defer GRPCActiveConnections.WithLabelValues(serviceName).Dec()

		resp, err := handler(ctx, req)

		duration := time.Since(start).Seconds()

		statusCode := codes.OK
		if err != nil {
			statusCode = status.Code(err)
		}

		GRPCRequestsTotal.WithLabelValues(
			serviceName,
			info.FullMethod,
			statusCode.String(),
		).Inc()

		GRPCRequestDuration.WithLabelValues(
			serviceName,
			info.FullMethod,
		).Observe(duration)

		return resp, err
	}
}
