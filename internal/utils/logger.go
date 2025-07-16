package utils

import (
	"go.uber.org/zap"
	"go.uber.org/zap"
)

func NewLogger(development bool) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error

	if development {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}

	if err != nil {
		return nil, err
	}

	return logger, nil
}