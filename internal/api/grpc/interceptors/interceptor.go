package interceptors

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitlab.ozon.dev/dimabelunin7/homework/hw-4/internal/api/grpc/monitoring"
)

func MetricsAndLoggingInterceptor(logger *zap.Logger, tracer trace.Tracer) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod)
		defer span.End()

		startTime := time.Now()
		monitoring.GrpcRequestsTotal.Inc()
		logger.Info("Starting RPC call",
			zap.String("method", info.FullMethod),
			zap.String("traceID", span.SpanContext().TraceID().String()),
		)

		resp, err := handler(ctx, req)
		duration := time.Since(startTime)
		durationSec := duration.Seconds()

		monitoring.RequestDuration.Observe(durationSec)
		monitoring.ResponseTimeSummary.Observe(durationSec)

		var code string
		if err != nil {
			s, _ := status.FromError(err)
			code = s.Code().String()
			monitoring.GrpcErrorsTotal.Inc()
			logger.Error("RPC call failed",
				zap.String("method", info.FullMethod),
				zap.String("traceID", span.SpanContext().TraceID().String()),
				zap.Error(err),
				zap.Duration("duration", duration),
			)
		} else {
			code = codes.OK.String()
			logger.Info("RPC call succeeded",
				zap.String("method", info.FullMethod),
				zap.String("traceID", span.SpanContext().TraceID().String()),
				zap.Duration("duration", duration),
			)
		}

		monitoring.GrpcRequestCountByStatus.WithLabelValues(code).Inc()
		monitoring.RpcCounter.WithLabelValues(info.FullMethod, code).Inc()

		return resp, err
	}
}
