package twirpzap

import (
	"context"
	"fmt"
	"time"

	"sync"

	"github.com/twitchtv/twirp"
	"go.uber.org/zap"
)

type contextKey struct{}

var logKey = contextKey{}

// ServerHooks creates twirp server hooks for logging
// using zap
func ServerHooks(logger *zap.Logger) *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			return requestReceived(ctx, logger)
		},
		ResponseSent: responseSent,
	}
}

var requestLoggerPool = sync.Pool{
	New: func() interface{} {
		return &requestLogger{}
	},
}

func requestReceived(ctx context.Context, logger *zap.Logger) (context.Context, error) {
	r := requestLoggerPool.Get().(*requestLogger)
	r.startTime = time.Now()
	r.logger = logger

	ctx = context.WithValue(ctx, logKey, r)
	return ctx, nil
}

type requestLogger struct {
	startTime time.Time
	logger    *zap.Logger
}

func responseSent(ctx context.Context) {
	r, ok := ctx.Value(logKey).(*requestLogger)
	if !ok || r == nil {
		fmt.Println(ok, r)
		return
	}
	pkg, _ := twirp.PackageName(ctx)
	svc, _ := twirp.ServiceName(ctx)
	meth, _ := twirp.MethodName(ctx)
	status, _ := twirp.StatusCode(ctx)

	duration := time.Since(r.startTime)

	r.logger.Info("response sent",
		zap.String("twirp.package", pkg),
		zap.String("twirp.service", svc),
		zap.String("twirp.method", meth),
		zap.String("twirp.status", status),
		zap.Duration("duration", duration),
	)

	requestLoggerPool.Put(r)
}
