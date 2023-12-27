package imo

import (
	"context"
	"fmt"

	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
)

// Writer provides a tool to copy only the layers that are not already
// present in a different version of the same image.
type Writer struct {
	types.ImageReference
	dest *destwrap
}

// NewImageDestination returns a handler used to write.
func (i *Writer) NewImageDestination(ctx context.Context, sys *types.SystemContext) (types.ImageDestination, error) {
	return i.dest, nil
}

// destwrap wraps an image destination (this can be another registry or a
// file on disk) and the original manifest from where we can extract the
// layers that are already present.
type destwrap struct {
	types.ImageDestination
	manifests map[digest.Digest]bool
}

// TryReusingBlob is called by the image copy code to check if a layer is
// already present in the destination. If it is, we return true and the
// layer info. If it is not, we return false and the layer info. We use the
// manifest to check if the layer is already present.
func (d *destwrap) TryReusingBlob(ctx context.Context, info types.BlobInfo, cache types.BlobInfoCache, substitute bool) (bool, types.BlobInfo, error) {
	if _, ok := d.manifests[info.Digest]; ok {
		return true, info, nil
	}
	return false, info, nil
}

// NewWriterFromScratch uses the "scratch" image as base and stores the
// result in 'to'. This is useful to create a new image from scratch.
func NewWriterFromScratch(ctx context.Context, to types.ImageReference, sysctx *types.SystemContext) (*Writer, error) {
	toref, err := to.NewImageDestination(ctx, sysctx)
	if err != nil {
		return nil, fmt.Errorf("error creating destination: %w", err)
	}
	return &Writer{
		ImageReference: to,
		dest: &destwrap{
			ImageDestination: toref,
			manifests:        map[digest.Digest]bool{},
		},
	}, nil
}

// NewWriter is capable of providing an incremental copy of an image using
// 'from' as base and storing the result in 'to'.
func NewWriter(ctx context.Context, from types.ImageReference, to types.ImageReference, sysctx *types.SystemContext) (*Writer, error) {
	toref, err := to.NewImageDestination(ctx, sysctx)
	if err != nil {
		return nil, fmt.Errorf("error creating destination: %w", err)
	}
	mans, err := fetchManifests(ctx, from, sysctx)
	if err != nil {
		return nil, fmt.Errorf("error fetching manifests: %w", err)
	}
	return &Writer{
		ImageReference: to,
		dest: &destwrap{
			ImageDestination: toref,
			manifests:        buildLayersDictionary(mans...),
		},
	}, nil
}
