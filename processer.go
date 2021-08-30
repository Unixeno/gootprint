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
		log.Infof("replace file: %s", filename)
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
	_ = filepath.WalkDir(packagePath, func(filePath string, d fs.DirEntry, err error) error {
		if pathFilter(filePath, excludePath) {
			log.Info("filter: ", filePath)
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
		files = append(files, filePath)
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

func (p *Processor) ProcessClean() {
	switch p.mode {
	case ModeFile:
		p.cleanFile(p.name)
	case ModeDir:
		p.cleanDir(p.name)
	case ModePackage:
		p.cleanPackage(p.name)
	}
}

func (p *Processor) cleanFile(filePath string) {
	log.Debugf("clean for `%s`", filePath)
	generated := filePath + ".gen.go"
	backup := filePath + ".gen_bak"
	// try to delete gen.go file
	info, err := os.Stat(generated)
	if err != nil && !os.IsNotExist(err) {
		log.WithError(err).Fatalf("can't delete file: %s", generated)
	}
	if err == nil {
		if info.IsDir() {
			log.Fatalf("can't remove `%s`, it's a directory", generated)
		}
		log.Debugf("try to remove %s", generated)
		err = os.Remove(generated)
		if err != nil && !os.IsNotExist(err) {
			log.WithError(err).Fatalf("can't delete file: %s", generated)
		}
	}
	// try to recover backup
	info, err = os.Stat(backup)
	if err != nil {
		if os.IsNotExist(err) {
			log.Errorf("backup file for `%s` does't exist", filePath)
			return
		} else {
			log.WithError(err).Fatalf("can't state backup file: %s", backup)
		}
	}
	if info.IsDir() {
		log.Fatal("can't find valid backup file")
	}
	log.Debug("try to recover ", backup)
	err = os.Rename(backup, filePath)
	if err != nil {
		log.WithError(err).Fatalf("failed to recover `%s`", filePath)
	}
	log.Infof("recoverd: %s", filePath)
}

func (p *Processor) cleanDir(dir string) {
	files, err := os.ReadDir(dir)
	if err != nil {
		log.Fatalf("failed to read directory: %v", err)
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		filename := file.Name()
		if strings.HasSuffix(filename, ".gen.go") {
			p.cleanFile(strings.TrimSuffix(path.Join(dir, filename), ".gen.go"))
		}
	}
}

func (p *Processor) cleanPackage(packagePath string) {
	files := make([]string, 0)
	excludePath := make([]string, 0)
	for _, exclude := range excludeList {
		excludePath = append(excludePath, path.Join(packagePath, exclude))
	}
	_ = filepath.WalkDir(packagePath, func(filePath string, d fs.DirEntry, err error) error {
		if pathFilter(filePath, excludePath) {
			log.Info("filter: ", filePath)
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
		if strings.HasSuffix(filename, ".gen.go") {
			files = append(files, strings.TrimSuffix(filePath, ".gen.go"))
		}
		return nil
	})
	for _, filename := range files {
		p.cleanFile(filename)
	}
}

func pathFilter(filePath string, rules []string) bool {
	for _, rule := range rules {
		if filePath == rule {
			return true
		}
	}
	return false
}
