<p align="center">
    <img src="banner.png" alt="Banner">
</p>

# Image Overlay Kit

The `imo` library provides a way to handle incremental image
updates in container registries. It allows users to pull the
difference between two image versions and push this differential
update to a registry.

This can significantly reduce the amount of data transferred,
especially useful in air-gapped (disconnected) environments.

## Key Types and Functions

### Incremental

The `Incremental` type offers methods to retrieve (via `Pull`) or transmit
(via `Push`) the differences between two images. These differences are determined
based on the base and final images. It is crucial to ensure that the destination
registry, when pushing these differences, possesses all the layers not encompassed
in the incremental difference." You can use _PushVet()_ for checking.

#### Methods

- **PushVet**: Verifies if all layers not included in the incremental difference
exist in the destination registry. It returns an error if any layer is missing.
- **Push**: Pushes the incremental difference stored in an OCI-archive tarball to
the destination registry. It fails if the remote registry lacks any layers not
included in the incremental difference.
- **Pull**: Pulls the incremental difference between two images. It returns an
`io.ReadCloser` from which an OCI-archive tarball can be read. The caller is
responsible for closing the reader.
- **New**: Returns a new Incremental object. With this object, callers can calculate
or send the incremental difference between two images.

## Usage

### Pulling Image Differences

The `imo` library allows you to pull the difference between two container images as
a tarball. Here's how you can do it:

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

### Pushing Image Differences

You can also use `imo` to push a differential update to a container registry:

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
