package imo

import (
	"context"
	"fmt"

	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
)

// ProcessList expects raw to point to a manifest list and will iterate over
// the children and return them.
func ProcessList(ctx context.Context, fromref types.ImageSource, raw []byte, mime string) ([]manifest.Manifest, error) {
	list, err := manifest.ListFromBlob(raw, mime)
	if err != nil {
		return nil, fmt.Errorf("error parsing manifests: %w", err)
	}
	children := []manifest.Manifest{}
	for _, digest := range list.Instances() {
		raw, mime, err := fromref.GetManifest(ctx, &digest)
		if err != nil {
			return nil, fmt.Errorf("error getting child manifest: %w", err)
		}
		man, err := manifest.FromBlob(raw, mime)
		if err != nil {
			return nil, fmt.Errorf("error parsing manifest: %w", err)
		}
		children = append(children, man)
	}
	return children, nil
}

// fetchManifests returns the list of manifests that are present in the
// source image. In case of a manifest list it will iterate over the
// children and return them.
func FetchManifests(ctx context.Context, from types.ImageReference, sysctx *types.SystemContext) ([]manifest.Manifest, error) {
	fromref, err := from.NewImageSource(ctx, sysctx)
	if err != nil {
		return nil, fmt.Errorf("error creating image source: %w", err)
	}
	defer fromref.Close()
	raw, mime, err := fromref.GetManifest(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error getting manifest: %w", err)
	}
	if manifest.MIMETypeIsMultiImage(mime) {
		return ProcessList(ctx, fromref, raw, mime)
	}
	man, err := manifest.FromBlob(raw, mime)
	if err != nil {
		return nil, fmt.Errorf("error parsing manifest: %w", err)
	}
	return []manifest.Manifest{man}, nil
}

// BuildLayersDictionary goes through all the provided manifests and indexes all
// the layers by their digest in a map. The value of each entry in the map is a
// boolean that is always true.
func BuildLayersDictionary(manifests ...manifest.Manifest) map[digest.Digest]bool {
	mandict := map[digest.Digest]bool{}
	for _, man := range manifests {
		for _, layer := range man.LayerInfos() {
			mandict[layer.Digest] = true
		}
	}
	return mandict
}
