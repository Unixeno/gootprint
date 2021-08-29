package main

import (
	"os"
	"path"

	log "github.com/sirupsen/logrus"
	"golang.org/x/mod/modfile"
)

var PackageRoot string

func try(filename string) bool {
	content, err := os.ReadFile(filename)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		log.Fatalf("cannot read file `%s`, err: %s", filename, err)
	}
	module, err := modfile.Parse(filename, content, nil)
	if err != nil {
		log.Fatalf("failed to parse go mod file: %s", err)
	}
	log.Infof("found module %s", module.Module.Mod.Path)
	log.Infof("module path: %s", path.Dir(filename))
	PackageRoot = path.Dir(filename)
	return true
}

func LocateModule(packageDir string) bool {
	var filename string
	for {
		filename = path.Join(packageDir, "go.mod")
		if try(filename) {
			return true
		}
		if packageDir == "/" || packageDir == "." { // todo: may not support Windows
			return false
		}
		packageDir = path.Dir(packageDir) // try parent directory
	}
}
