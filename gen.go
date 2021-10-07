package main

import (
	"bytes"
	"fmt"
	"os"

	"github.com/Unixeno/gootprint/frame"
	log "github.com/sirupsen/logrus"
)

type Generator struct {
	sourceFilename string
	sourceContent  []byte
	outputFilename string
	outputFile     *os.File
	contextFrame   *frame.Context
}

func NewGenerator(source string, contextFrame *frame.Context) *Generator {
	content, err := os.ReadFile(source)
	if err != nil {
		log.WithError(err).Fatalf("failed to read source file")
	}
	sourceFileInfo, err := os.Stat(source)
	if err != nil {
		log.WithError(err).Fatalf("failed to stat source file")
	}

	// detect `\r\n` line break
	if bytes.Contains(content, []byte("\r\n")) {
		log.Fatal("found `\\r\\n` in source file, this will break code generation")
	}

	outputFilename := fmt.Sprintf("%s.gen.go", source)
	log.Infof("output file is %s", outputFilename)
	fd, err := os.OpenFile(outputFilename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, sourceFileInfo.Mode())
	if err != nil {
		log.WithError(err).Fatal("failed to open file for writing")
	}
	generator := Generator{
		sourceFilename: source,
		sourceContent:  content,
		outputFilename: outputFilename,
		outputFile:     fd,
		contextFrame:   contextFrame,
	}
	return &generator
}

var newLine = []byte("\n")

func (g *Generator) Generate() {
	g.contextFrame.PrepareGenerate()
	lines := bytes.Split(g.sourceContent, []byte("\n"))
	for lineNumber, line := range lines {
		g.outputFile.Write(g.contextFrame.GenerateLine(lineNumber+1, line))
		if lineNumber < len(lines)-1 {
			g.outputFile.Write(newLine)
		} else {
			// last line, if source file doesn't end with a new line, we need to add one
			if len(line) != 0 {
				_, _ = g.outputFile.Write(newLine)
			}
		}
	}
	g.outputFile.Write(g.contextFrame.GenerateEnv())
}

func (g *Generator) RenameSource() {
	err := os.Rename(g.sourceFilename, g.sourceFilename+".gen_bak")
	if err != nil {
		log.WithError(err).Fatalf("failed to rename source file %v", g.sourceFilename)
	}
}
