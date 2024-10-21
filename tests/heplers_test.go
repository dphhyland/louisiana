package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateStreamID(t *testing.T) {
	streamID := GenerateStreamID() // Make sure the function is exported in your main code
	assert.Regexp(t, `^stream-\d+$`, streamID)
}

func newString(s string) *string {
	return &s
}
