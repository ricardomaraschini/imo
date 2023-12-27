package main

import (
	"context"
	"os"

	"github.com/ricardomaraschini/imo"
)

func push() {
	// Create a new incremental pusher setting its output to the standard output
	// and providing credentials for writing (pushing) images.
	inc := imo.New(
		imo.WithReporterWriter(os.Stdout),
		imo.WithPushAuth("user", "pass"),
	)
	// PushVet verifies if we can push the difference stored in the incremental
	// file (difference.tar) on top of the the remote image myaccount/app:v1.0.0.
	// This ensures that the remote registry has all the layers we do not have
	// locally (blobs we haven't pulled).
	if err := inc.PushVet(
		context.Background(),
		"difference.tar",
		"myaccount/app:v1.0.0",
	); err != nil {
		panic(err)
	}
	// Push the difference.tar file to the registry. The difference.tar file was
	// created by the puller in the other example. If the remote registry misses
	// any of the layers this will fail. In other words, if we generate a diff
	// between v1 and v2 on registry A when we try to push to registry B it will
	// fail if registry B does not have the layers from v1.
	if err := inc.Push(
		context.Background(),
		"difference.tar",
		"myaccount/app:v2.0.0",
	); err != nil {
		panic(err)
	}
}
