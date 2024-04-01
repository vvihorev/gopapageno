package main

import (
	"testing"
)

func TestMainProgram(t *testing.T) {
	for i := 0; i < 1; i++ {
		if err := run(); err != nil {
			t.Fail()
		}
	}
}
