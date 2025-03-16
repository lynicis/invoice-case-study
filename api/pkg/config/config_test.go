package config

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestConfig_Read(t *testing.T) {
	assert.NotPanics(t, func() {
		config := Read()
		assert.NotNil(t, config)
	})
}
