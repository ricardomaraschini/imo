package imo

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Makes sure we are writing the output to a different writer, not os.Stdout.
func TestIncrementalCustomWriter(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	buf := bytes.NewBuffer(nil)
	inc := New(
		WithReporterWriter(buf),
		WithAllArchitectures(),
	)
	tmpf, err := os.CreateTemp("", "diff-*.tar")
	assert.NoError(t, err, "unable to create temp file")
	defer os.Remove(tmpf.Name())
	diff, err := inc.Pull(ctx, "scratch", "quay.io/prometheus/memcached-exporter")
	assert.NoError(t, err, "unable to pull the whole image")
	defer diff.Close()
	_, err = io.Copy(tmpf, diff)
	assert.NoError(t, err, "unable to copy image to temp file")
	err = tmpf.Close()
	assert.NoError(t, err, "unable to close temp file")
	assert.NotEmpty(t, buf.String(), "no output written to buffer")
}

// TestPullAlpine pulls all alpine images into an oci-archive on disk. copies all
// architecures.
func TestIncrementalPullAlpine(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	inc := New(
		WithReporterWriter(os.Stdout),
		WithAllArchitectures(),
	)
	// pulls alpine:latest into an oci-archive.
	tmpf, err := os.CreateTemp("", "diff-*.tar")
	assert.NoError(t, err, "unable to create temp file")
	defer os.Remove(tmpf.Name())
	diff, err := inc.Pull(ctx, "scratch", "alpine:latest")
	assert.NoError(t, err, "unable to pull the whole image")
	defer diff.Close()
	_, err = io.Copy(tmpf, diff)
	assert.NoError(t, err, "unable to copy image to temp file")
	err = tmpf.Close()
	assert.NoError(t, err, "unable to close temp file")
}

// This test depends on a few environment variables to be set in order for it to
// run. You need to specify the following variables:
// REGISTRY_USERNAME, REGISTRY_PASSWORD, REGISTRY_REPOADDR, and REGISTRY_TAGNAME.
// If these are not set, the test will be skipped. These environment variables
// point to a registry to where we can write and image and read it back from.
// This test copies the tomcat:10.1 image to this location under a random tag
// and then after that it compares tomcat 10.1 with 11.0 and pushes only the
// difference.
func TestIncrementalE2E(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	user := os.Getenv("REGISTRY_USERNAME")
	pass := os.Getenv("REGISTRY_PASSWORD")
	addr := os.Getenv("REGISTRY_REPOADDR")
	tag := os.Getenv("REGISTRY_TAGNAME")
	if user == "" || pass == "" || addr == "" || tag == "" {
		t.Skip("environment variables not set, skipping")
	}
	inc := New(
		WithReporterWriter(os.Stdout),
		WithPushAuth(user, pass),
		WithAllArchitectures(),
	)
	// pulls tomcat:10.1 into an oci-archive.
	tmpf, err := os.CreateTemp("", "diff-*.tar")
	assert.NoError(t, err, "unable to create temp file")
	defer os.Remove(tmpf.Name())
	diff, err := inc.Pull(ctx, "scratch", "tomcat:10.1")
	assert.NoError(t, err, "unable to pull the whole image")
	defer diff.Close()
	_, err = io.Copy(tmpf, diff)
	assert.NoError(t, err, "unable to copy image to temp file")
	err = tmpf.Close()
	assert.NoError(t, err, "unable to close temp file")
	// pushes the pulled tomecat:10.1 to the registry under.
	dst := fmt.Sprintf("%s/e2e:%s", addr, tag)
	err = inc.Push(ctx, tmpf.Name(), dst)
	assert.NoError(t, err, "unable to push whole image")
	// pulls the difference between tomcat:10.1 and tomcat:11.0.
	tmpf, err = os.CreateTemp("", "diff-*.tar")
	assert.NoError(t, err, "unable to create second temp file")
	defer os.Remove(tmpf.Name())
	diff, err = inc.Pull(ctx, "tomcat:10.1", "tomcat:11.0")
	assert.NoError(t, err, "unable to pull only the difference")
	_, err = io.Copy(tmpf, diff)
	assert.NoError(t, err, "unable to copy difference to temp file")
	err = tmpf.Close()
	assert.NoError(t, err, "unable to close temp file")
	// vet push (makes sure all non locally present layers exist remotely).
	err = inc.PushVet(ctx, tmpf.Name(), dst)
	assert.NoError(t, err, "unable to push vet image")
	// pushes the difference between tomcat:10.1 and tomcat:11.0.
	err = inc.Push(ctx, tmpf.Name(), dst)
	assert.NoError(t, err, "unable to push difference")
}
