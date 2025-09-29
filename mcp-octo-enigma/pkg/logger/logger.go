package logger

import (
	"go.uber.org/zap"
)

var current *zap.Logger

func Init(env string) (*zap.Logger, error) {
	var (
		lg  *zap.Logger
		err error
	)
	if env == "production" {
		lg, err = zap.NewProduction()
	} else {
		lg, err = zap.NewDevelopment()
	}
	if err != nil {
		return nil, err
	}
	current = lg
	return lg, nil
}

func L() *zap.Logger { return current }
