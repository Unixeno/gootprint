package main

import (
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/Unixeno/gootprint/frame"
	log "github.com/sirupsen/logrus"
)

const (
	ModeFile = iota
	ModeDir
	ModePackage
)

var FileCounter int

type Processor struct {
	name  string
	mode  int
	stats frame.Stats
}

func NewProcessor(name string, mode int) *Processor {
	return &Processor{
		name: name,
		mode: mode,
	}
}

func (p *Processor) processFile(filename string) {
	FileCounter++
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		log.Errorf("failed to get absolute filename for: %s", filename)
		absFilename = filename
	}
	log.Infof("parsering file: %s", absFilename)
	parser := NewParser(absFilename)
	parser.Parse()
	parser.FrameContext().PostOrderDump()
	p.stats.Add(parser.FrameContext().Stats())
	if *dryRun {
		return
	}
	log.Info("start generating...")
	generator := NewGenerator(absFilename, parser.FrameContext())
	_ = generator
	generator.Generate()
	if !*noRename {
		log.Infof("replace file: %s", *file)
		generator.RenameSource()
	}
}

func (p *Processor) processDir(dirname string) {
	files, err := os.ReadDir(dirname)
	if err != nil {
		log.Fatalf("failed to read directory: %v", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		if !strings.HasSuffix(filename, ".go") ||
			strings.HasSuffix(filename, "_test.go") ||
			strings.HasSuffix(filename, ".gen.go") {
			continue
		}
		p.processFile(path.Join(dirname, file.Name()))
	}
}

func (p *Processor) processPackage(packagePath string) {
	files := make([]string, 0)
	excludePath := make([]string, 0)
	for _, exclude := range excludeList {
		excludePath = append(excludePath, path.Join(packagePath, exclude))
	}
	_ = filepath.WalkDir(packagePath, func(path string, d fs.DirEntry, err error) error {
		if pathFilter(path, excludePath) {
			log.Info("filter: ", path)
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		filename := d.Name()
		if !strings.HasSuffix(filename, ".go") ||
			strings.HasSuffix(filename, "_test.go") ||
			strings.HasSuffix(filename, ".gen.go") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	for _, filename := range files {
		p.processFile(filename)
	}
}

func (p *Processor) Stats() frame.Stats {
	return p.stats
}

func (p *Processor) Process() {
	switch p.mode {
	case ModeFile:
		p.processFile(p.name)
	case ModeDir:
		p.processDir(p.name)
	case ModePackage:
		p.processPackage(p.name)
	}
}

func pathFilter(path string, rules []string) bool {
	for _, rule := range rules {
		if path == rule {
			return true
		}
	}
	return false
}
