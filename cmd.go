package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

var file = flag.String("file", "", "specify source `filename`")
var dir = flag.String("dir", "", "specify `directory`")
var packageDir = flag.String("package", "", "specify package path")
var dryRun = flag.Bool("dry", false, "dry run, only parse file, not generate code")
var outputStdout = flag.Bool("print", false, "output to stdout instead of a file, will set -no-rename by default")
var noRename = flag.Bool("no-rename", false, "do not replace source file after generate")
var stats = flag.Bool("stat", false, "show source code statistics")
var clean = flag.Bool("clean", false, "delete generated files and rename source file back")
var verbose = flag.Bool("v", false, "verbose mode, show debug log")
var silence = flag.Bool("s", false, "silence mode, hide info log")
var excludeList []string
var path string

func validateDir(dir string) bool {
	fileInfo, err := os.Stat(dir)
	if err != nil {
		log.Fatalf("can't stat directory %s", dir)
		return false
	}
	return fileInfo.IsDir()
}

func validate() {
	if *silence {
		log.SetLevel(log.WarnLevel)
	}
	if *verbose {
		log.SetLevel(log.DebugLevel)
	}
	if *file == "" && *packageDir == "" && *dir == "" {
		log.Fatal("need a valid filename, directory or package to process")
	}
	if *dir != "" && !validateDir(*dir) {
		log.Fatalf("path `%s` is not a directory", *dir)
	}

	if *packageDir != "" {
		if validateDir(*packageDir) { // package in directory form
			if !LocateModule(*packageDir) {
				log.Fatalf("cannot find go mod file in `%s` and it's upper directory", *packageDir)
			}
		} else {
			log.Fatalf("package directory `%s` is invalid", *packageDir)
		}
	}
}

func main() {
	flag.Func("e", "short version of '-exclude'", func(s string) error {
		excludeList = append(excludeList, s)
		return nil
	})
	flag.Func("exclude", "specify exclude `directory`, only works in package mode", func(s string) error {
		excludeList = append(excludeList, s)
		return nil
	})
	flag.Parse()
	validate()

	var processor *Processor
	switch {
	case *file != "":
		processor = NewProcessor(*file, ModeFile)
		path = *file
	case *dir != "":
		processor = NewProcessor(*dir, ModeDir)
		path = *dir
	case *packageDir != "":
		processor = NewProcessor(*packageDir, ModePackage)
		path = *packageDir
	}

	if *clean {
		processor.ProcessClean()
		return
	}

	startTime := time.Now()
	processor.Process()

	if *stats {
		codeStats := processor.Stats()
		fmt.Println("********************************************************")
		fmt.Printf("code structure statistics for: %v\n", path)
		fmt.Printf(" >> function amount: \t%d\n", codeStats.FuncAmount)
		fmt.Printf(" >> go func amount: \t%d\n", codeStats.GoFuncAmount)
		fmt.Printf(" >> if amount: \t\t%d\n", codeStats.IfAmount)
		fmt.Printf(" >> for amount: \t%d\n", codeStats.ForAmount)
		fmt.Printf(" >> case amount: \t%d\n", codeStats.CaseAmount)
		fmt.Printf("source file has %d lines, will produce %d tracing point(%.2f%%)\n",
			codeStats.Lines,
			codeStats.InjectionPoint,
			100*float32(codeStats.InjectionPoint)/float32(codeStats.Lines),
		)
		fmt.Printf("process %d files in %v\n", FileCounter, time.Since(startTime))
		fmt.Println("********************************************************")
	}
	log.Info("done all success")
}
