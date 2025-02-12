[![Go Reference](https://pkg.go.dev/badge/github.com/ricardomaraschini/imo.svg)](https://pkg.go.dev/github.com/ricardomaraschini/imo)
[![Unit Tests](https://github.com/ricardomaraschini/imo/actions/workflows/tests-on-merge.yaml/badge.svg)](https://github.com/ricardomaraschini/imo/actions/workflows/tests-on-merge.yaml)

# Image Overlay Kit

The `imo` module manages incremental updates for container images in
registries. It enables users to pull differences between image versions and
push these "incremental updates" to a registry, which can significantly reduce
data transfer size, particularly beneficial in air-gapped environments.

## Key Types and Functions

### Incremental

The `Incremental` type provides methods to retrieve (using `Pull`) or transmit
(using `Push`) the differences between two images. These differences are
determined based on the `base` and `final` images.

When comparing an image tag `v1` with an image tag `v2`, the `base` image is
`v1`, and the `final` image is `v2`. The `Incremental` object calculates the
difference between these two images.

#### Methods

- **New**
  - Creates a new Incremental object, enabling callers to calculate or send the
    incremental difference between two images.
- **Pull**
  - Pulls the incremental difference between two images as an `io.ReadCloser`,
    from which a tarball can be read. The caller is responsible for closing the
    reader.
- **PushVet**
  - Verifies whether all necessary layers exist in the destination registry.
    Returns an error if any layer is missing. Particularly useful before
    pushing an incremental update.
- **Push**
  - Pushes the incremental difference stored in a tarball to the destination
    registry. Fails if the remote registry lacks any required layers not
    included in the incremental update.

## Usage

### Pulling Image Differences

The `imo` library allows you to pull differences between container images as a
tarball:

```go
package main

import (
	"context"
	"io"
	"os"

	"github.com/ricardomaraschini/imo"
)

func pull() {
	// create a new incremental puller
	inc := imo.New(
		imo.WithReporterWriter(os.Stdout),
		imo.WithBaseAuth("user", "pass"),
		imo.WithFinalAuth("user2", "pass2"),
	)

	// pull the differential update
	diff, err := inc.Pull(
		context.Background(),
		"docker.io/myaccount/myapp:v1.0.0",
		"docker.io/myaccount/myapp:v2.0.0",
	)
	if err != nil {
		panic(err)
	}
	defer diff.Close()

	// save the differential update to a file
	fp, err := os.Create("difference.tar")
	if err != nil {
		panic(err)
	}
	defer fp.Close()

	if _, err := io.Copy(fp, diff); err != nil {
		panic(err)
	}
}
```

The method `Pull` requires two image tags: the `base` image and the `final` image.
If you want to pull an image as a whole, without calculating the difference, you
can use `scratch` as `base`:

```go
// this will pull docker.io/myaccount/myapp:v1.0.0.
diff, err := inc.Pull(
	context.Background(),
	"scratch",
	"docker.io/myaccount/myapp:v1.0.0",
)
```

### Pushing Image Differences

You can also use `imo` to push a differential update to a container registry.
Before pushing it is important to validate that the remote registry has all the
layers not included in the incremental difference. You can use `PushVet` for
this:

```go
package main

import (
	"context"
	"os"

	"github.com/ricardomaraschini/imo"
)

func push() {
	// create a new incremental pusher
	inc := imo.New(
		imo.WithReporterWriter(os.Stdout),
		imo.WithPushAuth("user", "pass"),
	)

	// check if the remote registry has all needed layers.
	if err := imo.PushVet(
		context.Background(),
		"difference.tar",
		"myregistry.io/myaccount/app:v1.0.0",
	); err != nil {
		panic(err)
	}

	// push the differential update to the registry
	if err := inc.Push(
		context.Background(),
		"difference.tar",
		"myregistry.io/myaccount/app:v2.0.0",,
	); err != nil {
		panic(err)
	}
}
```
