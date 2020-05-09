package lurker

import (
	"go.uber.org/zap"
)

var log *zap.SugaredLogger

func init() {
	cfg := zap.NewProductionConfig()
	//cfg.Level = logLvToAtomicLv(Level)
	cfg.OutputPaths = []string{"stdout"}
	cfg.ErrorOutputPaths = []string{"stdout"}
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
