package imo

import (
	"io"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/types"
)

// Option is a functional option for the Incremental type.
type Option func(*Incremental)

// WithReportWriter sets the report writer for the Incremental type. By default the
// report writer is io.Discard.
func WithReporterWriter(reporter io.Writer) Option {
	return func(inc *Incremental) {
		inc.report = reporter
	}
}

// WithBaseAuth sets the authentication for the registry from where we are going to
// pull the "base" image. If we are comparing images v1 and v2 this is the auth for
// v1 registry.
func WithBaseAuth(user, pass string) Option {
	return func(inc *Incremental) {
		inc.auths.BaseAuth = &types.DockerAuthConfig{
			Username: user,
			Password: pass,
		}
	}
}

// WithPushAuth sets the authentication for the registry where we are going to push
// the incremental difference. If we have previously compared images v1 and v2 and
// we are going to push the difference to v3 this is the auth for v3 registry.
func WithPushAuth(user, pass string) Option {
	return func(inc *Incremental) {
		inc.auths.PushAuth = &types.DockerAuthConfig{
			Username: user,
			Password: pass,
		}
	}
}

// WithFinalAuth sets the authentication for the registry from where we are going to
// pull the "latest" image. If we are comparing images v1 and v2 this is the auth for
// v2 registry.
func WithFinalAuth(user, pass string) Option {
	return func(inc *Incremental) {
		inc.auths.FinalAuth = &types.DockerAuthConfig{
			Username: user,
			Password: pass,
		}
	}
}

// WithTempDir sets the temporary directory where we are going to store the diff while
// the user decides what to do with it. By default this is os.TempDir().
func WithTempDir(dir string) Option {
	return func(inc *Incremental) {
		inc.tmpdir = dir
	}
}

// WithInsecurePull sets the pull operation to skip TLS verification.
func WithInsecurePull() Option {
	return func(inc *Incremental) {
		inc.insecurePull = types.OptionalBoolTrue
	}
}

// WithInsecurePush sets the push operation to skip TLS verification.
func WithInsecurePush() Option {
	return func(inc *Incremental) {
		inc.insecurePush = types.OptionalBoolTrue
	}
}

// WithInsecure sets both the pull and push operations to skip TLS verification.
func WithInsecure() Option {
	return func(inc *Incremental) {
		inc.insecurePull = types.OptionalBoolTrue
		inc.insecurePush = types.OptionalBoolTrue
	}
}

// WithAllArchitectures sets the selection to include all the architectures.
func WithAllArchitectures() Option {
	return func(inc *Incremental) {
		inc.selection = copy.CopyAllImages
	}
}
