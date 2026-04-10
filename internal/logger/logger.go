package logger

import "go.uber.org/zap"

var Log *zap.Logger

func Init() error {
	logg, err := zap.NewProduction(
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel))
	if err != nil {
		return err
	}
	Log = logg
	return nil
}
