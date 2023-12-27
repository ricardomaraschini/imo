package main

import (
	"context"
	"io"
	"os"

	"github.com/ricardomaraschini/imo"
)

func copy() {
	// Create a new incremental puller setting its output to the standard output
	// and providing credentials for reading both images.
	inc := imo.New(
		imo.WithReporterWriter(os.Stdout),
		imo.WithBaseAuth("user", "pass"),
		imo.WithFinalAuth("user2", "pass2"),
	)
	// Make the pull, the result is an io.ReadCloser that can be used to read
	// the difference between the base and the final images. As the base image
	// is 'scratch' then this is equivalent of copying all layers from the final
	// image onto disk.
	diff, err := inc.Pull(
		context.Background(),
		"scratch",
		"myaccount/myapp:v1.0.0",
	)
	if err != nil {
		panic(err)
	}
	// We always need to close the diff reader.
	defer diff.Close()
	// Create a new place where we want to store the layers and copy it to there.
	fp, err := os.Create("difference.tar")
	if err != nil {
		panic(err)
	}
	defer fp.Close()
	if _, err := io.Copy(fp, diff); err != nil {
		panic(err)
	}
}
