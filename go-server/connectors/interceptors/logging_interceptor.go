package interceptors

import (
	"context"
	"fmt"
	"time"

	"connector-recruitment/go-server/connectors/logger"

	"google.golang.org/grpc"
)

func LoggingUnaryInterceptor(logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()

		// Process the request
		resp, err := handler(ctx, req)

		elapsed := time.Since(startTime)
		elapsedStr := fmt.Sprintf("%.2fms", elapsed.Seconds()*1000) // convert it to 

		// Log the method name, elapsed time, and any error
		logger.Info("Unary RPC completed",
			"method", info.FullMethod,
			"elapsed", elapsedStr,
			"error", err,
		)

		return resp, err
	}
}
