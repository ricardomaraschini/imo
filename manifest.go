package imo

import (
	"context"
	"fmt"
	"sync"

	"github.com/containers/image/v5/manifest"
	"github.com/containers/image/v5/types"
	"github.com/opencontainers/go-digest"
)

// ManifestsIndex is an entity that indexes multiple manifests that are part
// of the same image. Provide tooling around the manifests.
type ManifestsIndex struct {
	mtx       sync.RWMutex
	index     map[digest.Digest]bool
	sysctx    *types.SystemContext
	manifests []manifest.Manifest
}

// HasLayer returns true if the layer is referred by any of the indexed
// manifests.
func (m *ManifestsIndex) HasLayer(dgst digest.Digest) bool {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	_, ok := m.index[dgst]
	return ok
}

// Manifests returns the manifests that were fetched for the image.
func (m *ManifestsIndex) Manifests() []manifest.Manifest {
	m.mtx.RLock()
	defer m.mtx.RUnlock()
	result := make([]manifest.Manifest, len(m.manifests))
	copy(result, m.manifests)
	return result
}

// FetchManifests gets the manifests from the source image and indexes all the
// layers that are present in the manifests in the internal 'index' map. Users
// can then call 'HasLayer' to check if a layer is present in the source image.
func (m *ManifestsIndex) FetchManifests(ctx context.Context, from types.ImageReference) error {
	m.mtx.Lock()
	defer m.mtx.Unlock()
	fromref, err := from.NewImageSource(ctx, m.sysctx)
	if err != nil {
		return fmt.Errorf("error creating image source: %w", err)
	}
	defer fromref.Close()
	raw, mime, err := fromref.GetManifest(ctx, nil)
	if err != nil {
		return fmt.Errorf("error getting manifest: %w", err)
	}
	if manifest.MIMETypeIsMultiImage(mime) {
		return m.fetchFromList(ctx, fromref, raw, mime)
	}
	man, err := manifest.FromBlob(raw, mime)
	if err != nil {
		return fmt.Errorf("error parsing manifest: %w", err)
	}
	m.manifests = []manifest.Manifest{man}
	m.buildIndex()
	return nil
}

// NewManifestsIndex creates a new ManifestsIndex. A ManifestsIndex is an entity capable
// of indexing all layers that are present in the manifests of an image. Uses the provided
// SystemContext to access the remote manifests.
func NewManifestsIndex(sysctx *types.SystemContext) *ManifestsIndex {
	return &ManifestsIndex{
		index:  map[digest.Digest]bool{},
		sysctx: sysctx,
	}
}

// buildIndex builds an index of all the layers that are present in the
// provided manifests.
func (m *ManifestsIndex) buildIndex() {
	m.index = map[digest.Digest]bool{}
	for _, man := range m.manifests {
		for _, layer := range man.LayerInfos() {
			m.index[layer.Digest] = true
		}
	}
}

// fetchFromList is used to parse children manifests of a manifest list.
func (m *ManifestsIndex) fetchFromList(ctx context.Context, fromref types.ImageSource, raw []byte, mime string) error {
	list, err := manifest.ListFromBlob(raw, mime)
	if err != nil {
		return fmt.Errorf("error parsing manifests: %w", err)
	}
	children := []manifest.Manifest{}
	for _, digest := range list.Instances() {
		raw, mime, err := fromref.GetManifest(ctx, &digest)
		if err != nil {
			return fmt.Errorf("error getting child manifest: %w", err)
		}
		man, err := manifest.FromBlob(raw, mime)
		if err != nil {
			return fmt.Errorf("error parsing manifest: %w", err)
		}
		children = append(children, man)
	}
	m.manifests = children
	m.buildIndex()
	return nil
}
