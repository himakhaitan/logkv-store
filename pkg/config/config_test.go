package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_ReturnsDefaultConfig(t *testing.T) {
	cfg, err := Load()
	assert.NoError(t, err)
	assert.NotNil(t, cfg)

	// default data directory should be set to "data"
	assert.Equal(t, "data", cfg.DataDir)
}

func TestConfigStruct_SetAndGetFields(t *testing.T) {
	// verify that the Config struct behaves as expected
	cfg := &Config{}
	cfg.DataDir = "custom_path"

	assert.Equal(t, "custom_path", cfg.DataDir)
}
