package lurker

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.SugaredLogger

func init() {
	cfg := zap.NewProductionConfig()
	//cfg.Level = logLvToAtomicLv(Level)
	cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	cfg.OutputPaths = []string{"lk.log"}
	cfg.ErrorOutputPaths = []string{"lk_err.log"}
	logger, e := cfg.Build(
		zap.AddCaller(),
		//zap.AddCallerSkip(1),
	)
	if e != nil {
		panic(e)
	}
	log = logger.Sugar()
	log.Debugw("log init")
}
