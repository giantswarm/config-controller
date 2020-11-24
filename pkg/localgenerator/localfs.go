package localgenerator

import (
	"io/ioutil"
	"os"
)

type LocalGenerator struct{}

func (l LocalGenerator) ReadDir(dir string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(dir)
}

func (l LocalGenerator) ReadFile(filename string) ([]byte, error) {
	return ioutil.ReadFile(filename)
}
