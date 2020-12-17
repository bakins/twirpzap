package twirpzap

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/twitchtv/twirp"
	"github.com/twitchtv/twirp/ctxsetters"
	"github.com/twitchtv/twirp/example"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		hatSize int32
	}{
		{
			name:    "simple 200",
			status:  "200",
			hatSize: 10,
		},
		{
			name:    "invalid size",
			status:  "400",
			hatSize: 0,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			var buf bytes.Buffer
			l := newLogger(&buf)

			server := example.NewHaberdasherServer(&testHaberdasher{}, ServerHooks(l))
			svr := httptest.NewServer(server)
			defer svr.Close()

			client := example.NewHaberdasherProtobufClient(svr.URL, &http.Client{})

			_, err := client.MakeHat(context.Background(), &example.Size{Inches: test.hatSize})

			if test.status != "200" {
				require.Error(t, err)
			}

			data := buf.String()
			require.Contains(t, data, `"twirp_status":"`+test.status+`"`)
			require.Contains(t, data, `"twirp_service":"Haberdasher"`)
			require.Contains(t, data, `"twirp_method":"MakeHat"`)
		})
	}
}

func BenchmarkServerHooks(b *testing.B) {
	hooks := ServerHooks(nullLogger)

	parent := ctxsetters.WithMethodName(context.Background(), "MakeHat")
	parent = ctxsetters.WithServiceName(parent, "Haberdasher")
	parent = ctxsetters.WithPackageName(parent, "twitch.twirp.example")
	parent = ctxsetters.WithStatusCode(parent, http.StatusOK)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		ctx, err := hooks.RequestReceived(parent)
		if err != nil {
			b.Fatal(err)
		}

		ctx, err = hooks.RequestRouted(ctx)
		if err != nil {
			b.Fatal(err)
		}

		hooks.ResponseSent(ctx)
	}
}

func newLogger(w io.Writer) *zap.Logger {
	encoderCfg := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		NameKey:        "logger",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
	}
	core := zapcore.NewCore(zapcore.NewJSONEncoder(encoderCfg), zapcore.AddSync(w), zapcore.DebugLevel)
	return zap.New(core)
}

type testHaberdasher struct{}

func (h *testHaberdasher) MakeHat(ctx context.Context, size *example.Size) (*example.Hat, error) {
	if size.Inches <= 0 {
		return nil, twirp.InvalidArgumentError("Inches", "I can't make a hat that small!")
	}
	return &example.Hat{
		Size:  size.Inches,
		Color: "black",
		Name:  "derby",
	}, nil
}
