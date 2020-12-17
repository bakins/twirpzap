package twirpzap

import (
	"context"
	"sync"
	"time"

	"github.com/twitchtv/twirp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type contextKey struct{}

var (
	logKey     = contextKey{}
	nullLogger = zap.NewNop()
)

// ServerHooks creates twirp server hooks for logging
// using zap
func ServerHooks(logger *zap.Logger) *twirp.ServerHooks {
	return &twirp.ServerHooks{
		RequestReceived: func(ctx context.Context) (context.Context, error) {
			return requestReceived(ctx, logger)
		},
		RequestRouted: responseRouted,
		ResponseSent:  responseSent,
	}
}

type requestLogger struct {
	startTime time.Time
	logger    *zap.Logger
	fields    []zap.Field
}

var requestLoggerPool = sync.Pool{
	New: func() interface{} {
		return &requestLogger{
			fields: make([]zap.Field, 0, 10),
		}
	},
}

func requestReceived(ctx context.Context, logger *zap.Logger) (context.Context, error) {
	r := requestLoggerPool.Get().(*requestLogger)
	r.startTime = time.Now()
	r.logger = logger
	r.fields = r.fields[:]

	if pkg, ok := twirp.PackageName(ctx); ok {
		r.fields = append(r.fields, zap.String("twirp_package", pkg))
	}

	if svc, ok := twirp.ServiceName(ctx); ok {
		r.fields = append(r.fields, zap.String("twirp_service", svc))
	}

	ctx = context.WithValue(ctx, logKey, r)
	return ctx, nil
}

func responseRouted(ctx context.Context) (context.Context, error) {
	if meth, ok := twirp.MethodName(ctx); ok {
		AddFields(ctx, zap.String("twirp_method", meth))
	}

	return ctx, nil
}

func responseSent(ctx context.Context) {
	r, ok := ctx.Value(logKey).(*requestLogger)
	if !ok || r == nil {
		return
	}

	duration := time.Since(r.startTime)

	r.fields = append(r.fields, zap.Duration("twirp_duration", duration))

	if status, ok := twirp.StatusCode(ctx); ok {
		r.fields = append(r.fields, zap.String("twirp_status", status))
	}

	r.logger.With(r.fields...).Info("response sent")

	r.logger = nullLogger
	r.fields = r.fields[:]

	requestLoggerPool.Put(r)
}

// based on https://github.com/grpc-ecosystem/go-grpc-middleware/blob/master/logging/zap/ctxzap/context.go

// AddFields adds zap fields to the logger.
func AddFields(ctx context.Context, fields ...zapcore.Field) {
	l, ok := ctx.Value(logKey).(*requestLogger)
	if !ok || l == nil {
		return
	}
	l.fields = append(l.fields, fields...)
}

// FromContext returns the request scoped logger.
func FromContext(ctx context.Context) *zap.Logger {
	l, ok := ctx.Value(logKey).(*requestLogger)
	if !ok || l == nil || l.logger == nil {
		return nullLogger
	}
	return l.logger.With(l.fields...)
}
