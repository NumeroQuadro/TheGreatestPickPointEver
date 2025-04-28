package logger

import (
	"go.uber.org/zap"
)

var ZapLogger *zap.Logger

func init() {
	atomicLevel := zap.NewAtomicLevel()
	atomicLevel.SetLevel(zap.DebugLevel)

	cfg := zap.Config{
		Level:            atomicLevel,
		Development:      false,
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	var err error
	ZapLogger, err = cfg.Build()
	if err != nil {
		panic("failed to build zap logger: " + err.Error())
	}
}
