// Package storage hold and abstraction of the filesystem

package storage

import (
	"io/fs"
)

type FSType int

// Filesystem is the current filesytem used as the underlying layer to manage the files.
// See the storage.Use method to change it.
var Filesystem Storage

const (
	// This defines the local filesystem
	OsFS FSType = iota
	// This defines the memory filesystem
	MemoryFS
)

// Use s as default storage/ This is tipically used in an init function.
func Use(s FSType) {
	switch s {
	case OsFS:
		//TODO
		Filesystem = nil
	case MemoryFS:
		//TODO
		Filesystem = nil
	default:
		panic("unexpected FSType")
	}
}

// Storage is an abstraction of the filesystem
type Storage interface {
	fs.FS
	WriteFile(name string, data []byte, perm fs.FileMode) error
}

func WriteFile(name string, data []byte, perm fs.FileMode) error {
	return Filesystem.WriteFile(name, data, perm)
}
