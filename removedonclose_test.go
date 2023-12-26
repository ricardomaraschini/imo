package imo

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveOnClose(t *testing.T) {
	tfile, err := os.CreateTemp("", "imo-test")
	assert.NoError(t, err, "unable to create temp file")
	tpath := tfile.Name()
	roc := RemoveOnClose{File: tfile, path: tpath}
	err = roc.Close()
	assert.NoError(t, err)
	if _, err := os.Stat(tpath); !os.IsNotExist(err) {
		t.Fatalf("File %s was not removed as expected", tpath)
	}
}
