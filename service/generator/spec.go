package generator

import "os"

type Filesystem interface {
	Exists(string) bool
	ReadFile(string) ([]byte, error)
	ReadDir(string) ([]os.FileInfo, error)
}
