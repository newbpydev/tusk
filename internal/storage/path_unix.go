//go:build !windows

package storage

import "os"

func isReparse(os.FileInfo) bool { return false }
