package imo

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/containers/image/v5/copy"
	"github.com/containers/image/v5/signature"
	"github.com/containers/image/v5/transports/alltransports"
	"github.com/containers/image/v5/types"
	"github.com/google/uuid"
)

// Authentications holds the all the necessary authentications for the incremental
// operations. BaseAuth is the authentication for the base image, FinalAuth is the
// authentication for the final image and PushAuth is the authentication for the
// destination registry. For example, let's suppose we want to get an incremental
// difference between an imaged hosted on X registry and an image hosted on Y
// registry and late on we want to push the difference to Z registry.
// In this case:
// - BaseAuth is the authentication for the X registry.
// - FinalAuth is the authentication for the Y registry.
// - PushAuth is the authentication for the Z registry.
type Authentications struct {
	BaseAuth  *types.DockerAuthConfig
	FinalAuth *types.DockerAuthConfig
	PushAuth  *types.DockerAuthConfig
}

// Incremental provides tooling about getting (Pull) or sending (Push) the difference
// between two images. The difference is calculated between the base and final images.
// When Pushing the difference to a destination registry it is important to note that
// the other layers (the ones not included in the 'difference') exist.
type Incremental struct {
	tmpdir    string
	polctx    *signature.PolicyContext
	report    io.Writer
	auths     Authentications
	selection copy.ImageListSelection
}

// PushVet verifies if all the layers not included in the incremental difference exist
// in the destination registry. If not, it returns an error.
func (inc *Incremental) PushVet(ctx context.Context, src, dst string) error {
	dst = fmt.Sprintf("docker://%s", dst)
	dstref, err := alltransports.ParseImageName(dst)
	if err != nil {
		return fmt.Errorf("error parsing destination reference: %w", err)
	}
	sysctx := &types.SystemContext{DockerAuthConfig: inc.auths.PushAuth}
	mans, err := FetchManifests(ctx, dstref, sysctx)
	if err != nil {
		return fmt.Errorf("error fetching destination manifests: %w", err)
	}
	dict := BuildLayersDictionary(mans...)
	srcref, err := alltransports.ParseImageName(fmt.Sprintf("oci-archive:%s", src))
	if err != nil {
		return fmt.Errorf("error parsing source reference: %w", err)
	}
	srcimage, err := srcref.NewImageSource(ctx, &types.SystemContext{})
	if err != nil {
		return fmt.Errorf("error creating source image: %w", err)
	}
	defer srcimage.Close()
	srcmans, err := FetchManifests(ctx, srcref, &types.SystemContext{})
	if err != nil {
		return fmt.Errorf("error fetching source manifests: %w", err)
	}
	for _, srcman := range srcmans {
		for _, layer := range srcman.LayerInfos() {
			binfo := types.BlobInfo{Digest: layer.Digest}
			blob, _, err := srcimage.GetBlob(ctx, binfo, nil)
			if err == nil {
				blob.Close()
				continue
			}
			if _, ok := dict[layer.Digest]; ok {
				continue
			}
			return fmt.Errorf("%s not found in destination", layer.Digest)
		}
	}
	return nil
}

// Push pushes the incremental difference stored in the oci-archive tarball pointed by
// src to the destination registry pointed by to. Be aware that if the remote registry
// does not contain one or more of the layers not included in the incremental difference
// the push will fail.
func (inc *Incremental) Push(ctx context.Context, src, dst string) error {
	dst = fmt.Sprintf("docker://%s", dst)
	dstref, err := alltransports.ParseImageName(dst)
	if err != nil {
		return fmt.Errorf("error parsing destination reference: %w", err)
	}
	srcref, err := alltransports.ParseImageName(fmt.Sprintf("oci-archive:%s", src))
	if err != nil {
		return fmt.Errorf("error parsing source reference: %w", err)
	}
	polctx, err := PolicyContext()
	if err != nil {
		return fmt.Errorf("error creating policy context: %w", err)
	}
	if _, err := copy.Image(
		ctx,
		polctx,
		dstref,
		srcref,
		&copy.Options{
			ReportWriter:       inc.report,
			SourceCtx:          &types.SystemContext{},
			ImageListSelection: inc.selection,
			DestinationCtx: &types.SystemContext{
				DockerAuthConfig: inc.auths.PushAuth,
			},
		},
	); err != nil {
		return fmt.Errorf("failed copying layers: %w", err)
	}
	return nil
}

// Pull pulls the incremental difference between two images. Returns an ReaderCloser from
// where can be read as an oci-archive tarball. The caller is responsible for closing the
// reader.
func (inc *Incremental) Pull(ctx context.Context, base, final string) (io.ReadCloser, error) {
	base = fmt.Sprintf("docker://%s", base)
	baseref, err := alltransports.ParseImageName(base)
	if err != nil {
		return nil, fmt.Errorf("error parsing base reference: %w", err)
	}
	final = fmt.Sprintf("docker://%s", final)
	finalref, err := alltransports.ParseImageName(final)
	if err != nil {
		return nil, fmt.Errorf("error parsing final reference: %w", err)
	}
	fname := fmt.Sprintf("%s.tar", uuid.New().String())
	tpath := path.Join(inc.tmpdir, fname)
	dstref, err := alltransports.ParseImageName(fmt.Sprintf("oci-archive:%s", tpath))
	if err != nil {
		return nil, fmt.Errorf("error parsing destination reference: %w", err)
	}
	sysctx := &types.SystemContext{DockerAuthConfig: inc.auths.BaseAuth}
	destref, err := NewWriter(ctx, baseref, dstref, sysctx)
	if err != nil {
		return nil, fmt.Errorf("error creating incremental writer: %w", err)
	}
	polctx, err := PolicyContext()
	if err != nil {
		return nil, fmt.Errorf("error creating policy context: %w", err)
	}
	if _, err := copy.Image(
		ctx,
		polctx,
		destref,
		finalref,
		&copy.Options{
			ReportWriter:       inc.report,
			DestinationCtx:     &types.SystemContext{},
			ImageListSelection: inc.selection,
			SourceCtx: &types.SystemContext{
				DockerAuthConfig: inc.auths.FinalAuth,
			},
		},
	); err != nil {
		return nil, fmt.Errorf("failed copying layers: %w", err)
	}
	fp, err := os.Open(tpath)
	if err != nil {
		os.Remove(tpath)
		return nil, fmt.Errorf("error opening tarball: %w", err)
	}
	return RemoveOnClose{fp, tpath}, nil
}

// New returns a new Incremental object. With Incremental objects callers can calculate
// the incremental difference between two images (Pull) or send the incremental towards
// a destination (Push).
func New(opts ...Option) *Incremental {
	inc := &Incremental{
		tmpdir:    os.TempDir(),
		report:    io.Discard,
		selection: copy.CopySystemImage,
	}
	for _, opt := range opts {
		opt(inc)
	}
	return inc
}
