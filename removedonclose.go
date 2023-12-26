package imo

import "os"

// RemoveOnClose is a wrapper around a file that removes the file when closed.
type RemoveOnClose struct {
	*os.File
	path string
}

// Close removes the file and closes the underlying file.
func (r RemoveOnClose) Close() error {
	if err := r.File.Close(); err != nil {
		return err
	}
	return os.Remove(r.path)
}
