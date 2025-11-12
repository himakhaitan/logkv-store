package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewLogger_Success(t *testing.T) {
	logger, err := New("test-service")
	assert.NoError(t, err)
	assert.NotNil(t, logger)
	_, ok := interface{}(logger).(*zap.Logger)
	assert.True(t, ok)
	assert.NotPanics(t, func() {
		logger.Info("Logger initialized successfully")
	})
}

func TestNewLogger_IndependentInstances(t *testing.T) {
	logger1, err1 := New("svc1")
	logger2, err2 := New("svc2")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NotNil(t, logger1)
	assert.NotNil(t, logger2)

	// They should be different instances (not pointing to the same logger)
	assert.NotEqual(t, logger1, logger2)
}
