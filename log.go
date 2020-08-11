package lurker

import (
	golog "github.com/goextension/log"
)

type dummyLog struct {
}

var log golog.Logger

var _ golog.Logger = &dummyLog{}

func init() {
	log = &dummyLog{}
}

// Debug ...
func (d dummyLog) Debug(args ...interface{}) {

}

// Info ...
func (d dummyLog) Info(args ...interface{}) {
}

// Warn ...
func (d dummyLog) Warn(args ...interface{}) {
}

// Error ...
func (d dummyLog) Error(args ...interface{}) {
}

// DPanic ...
func (d dummyLog) DPanic(args ...interface{}) {
}

// Panic ...
func (d dummyLog) Panic(args ...interface{}) {
}

// Fatal ...
func (d dummyLog) Fatal(args ...interface{}) {
}

// Debugf ...
func (d dummyLog) Debugf(template string, args ...interface{}) {
}

// Infof ...
func (d dummyLog) Infof(template string, args ...interface{}) {
}

// Warnf ...
func (d dummyLog) Warnf(template string, args ...interface{}) {
}

// Errorf ...
func (d dummyLog) Errorf(template string, args ...interface{}) {
}

// DPanicf ...
func (d dummyLog) DPanicf(template string, args ...interface{}) {
}

// Panicf ...
func (d dummyLog) Panicf(template string, args ...interface{}) {
}

// Fatalf ...
func (d dummyLog) Fatalf(template string, args ...interface{}) {
}

// Debugw ...
func (d dummyLog) Debugw(msg string, keysAndValues ...interface{}) {
}

// Infow ...
func (d dummyLog) Infow(msg string, keysAndValues ...interface{}) {
}

// Warnw ...
func (d dummyLog) Warnw(msg string, keysAndValues ...interface{}) {
}

// Errorw ...
func (d dummyLog) Errorw(msg string, keysAndValues ...interface{}) {
}

// DPanicw ...
func (d dummyLog) DPanicw(msg string, keysAndValues ...interface{}) {
}

// Panicw ...
func (d dummyLog) Panicw(msg string, keysAndValues ...interface{}) {
}

// Fatalw ...
func (d dummyLog) Fatalw(msg string, keysAndValues ...interface{}) {
}
