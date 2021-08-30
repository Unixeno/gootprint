package frame

import (
	"fmt"
	"strconv"

	log "github.com/sirupsen/logrus"
)

type Context struct {
	rootFrame Frame // root frame is the package frame
	indexes   []int // record levels from root frame to current frame，works as a stack
	hooks     map[int][]func([]byte) []byte
}

func NewFrameContext(filename, packageName string, bodyBegin, bodyEnd int) *Context {
	return &Context{
		rootFrame: NewPackageFrame(filename, packageName, bodyBegin, bodyEnd),
		indexes:   make([]int, 1, 16),
		hooks:     map[int][]func([]byte) []byte{},
	}
}

func (root *Context) Import(path string) {
	root.rootFrame.(*PackageFrame).Import(path)
}

func (root *Context) GetIndexName() string {
	return strconv.Itoa(root.indexes[len(root.indexes)-1])
}

func (root *Context) GetCurrent() Frame {
	var top = root.rootFrame
	var level = 1
	for {
		if level >= len(root.indexes) || top.Len() == 0 {
			return top
		}
		level++
		top = top.GetInner(top.Len() - 1)
	}
}

func (root *Context) Push(frame Frame) {
	root.GetCurrent().Append(frame)
	root.indexes = append(root.indexes, 0)
}

func (root *Context) Pop() {
	root.indexes = root.indexes[:len(root.indexes)-1]
	root.indexes[len(root.indexes)-1]++
}

func (root *Context) PreOrderDump() {
	Visit(root.rootFrame, VisitPreOrder, func(frame Frame) {
		fmt.Println(frame)
	})
}

func (root *Context) PostOrderDump() {
	log.Debug("================Dump(post)==================")
	Visit(root.rootFrame, VisitPostOrder, func(frame Frame) {
		log.Debug(frame)
	})
	log.Debug("============================================")
}

func (root *Context) PrepareGenerate() {
	Visit(root.rootFrame, VisitPreOrder, func(frame Frame) {
		beginLine := frame.BodyBeginning()
		endingLine := frame.BodyEnding()
		if beginLine == endingLine {
			if _, ok := frame.(*GoFuncFrame); !ok {
				log.Warnf("block exist in same line, ignore generate, %s", frame)
				return
			}
		}

		root.hooks[frame.BodyBeginning()] = append(root.hooks[frame.BodyBeginning()], frame.GenBeginning)
		if !frame.Unreachable() {
			root.hooks[frame.BodyEnding()] = append(root.hooks[frame.BodyEnding()], frame.GenEnding)
		} else {
			log.Infof("skip `%s` because it's unreachable", frame.Path())
		}
	})
	log.Info("prepared")
}

func (root *Context) GenerateLine(line int, content []byte) []byte {
	if genFuncs, exist := root.hooks[line]; exist {
		for _, genFunc := range genFuncs {
			content = genFunc(content)
		}
	}
	return content
}

func (root *Context) GetInnerName(suffix string) string {
	current := root.GetCurrent()
	name := current.Path() + "." + suffix + "_" + strconv.Itoa(current.Len()+1)
	return name
}

func (root *Context) Stats() Stats {
	s := Stats{}
	Visit(root.rootFrame, VisitPreOrder, func(frame Frame) {
		switch frame.(type) {
		case *IfElseFrame:
			s.IfAmount++
		case *FuncFrame:
			s.FuncAmount++
		case *CaseFrame:
			s.CaseAmount++
		case *ForFrame:
			s.ForAmount++
		case *GoFuncFrame:
			s.GoFuncAmount++
		case *PackageFrame:
			s.Lines = frame.BodyEnding()
		}
		s.InjectionPoint = s.IfAmount + s.CaseAmount + s.FuncAmount + s.ForAmount + s.GoFuncAmount
	})
	return s
}
