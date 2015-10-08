package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsAbs(t *testing.T) {

	a := []string{"/map",
		"data/dd",
		"c:\\windows",
		"sdf\\ss",
		"logs"}

	for _, item := range a {
		t.Logf("%s is abs: %v", item, filepath.IsAbs(item))
	}
}