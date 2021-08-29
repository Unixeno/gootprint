package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"reflect"

	"github.com/Unixeno/gootprint/frame"
	log "github.com/sirupsen/logrus"
)

type Parser struct {
	filename    string
	sourceFile  io.Reader
	packageName string
	fSet        *token.FileSet
	fileNode    *ast.File
	level       int
	frameCtx    *frame.Context
}

func NewParser(filename string) *Parser {
	fSet := token.NewFileSet()
	node, err := parser.ParseFile(fSet, filename, nil, 0)
	if err != nil {
		log.WithError(err).WithField("filename", filename).Fatal("failed to parse source file")
	}
	return &Parser{
		filename: filename,
		fSet:     fSet,
		fileNode: node,
	}
}

func (p *Parser) Parse() {
	p.packageName = p.fileNode.Name.Name
	log.Debugf("found package %v", p.packageName)
	p.frameCtx = frame.NewFrameContext(p.filename, p.packageName, p.getLine(p.fileNode.Pos()), p.getLine(p.fileNode.End()))
	f := p.fileNode
	for _, decl := range f.Decls {
		if funcDecl, ok := decl.(*ast.FuncDecl); ok {
			p.parseFunc(funcDecl)
		} else if genDecl, ok := decl.(*ast.GenDecl); ok && genDecl.Tok == token.IMPORT {
			log.Debugf("found import")
			p.parseImport(genDecl.Specs)
		}
	}
	return
}

func (p *Parser) parseImport(specs []ast.Spec) {
	for _, spec := range specs {
		if importSpec, ok := spec.(*ast.ImportSpec); ok {
			log.Debugf("  >> %s", importSpec.Path.Value)
			p.frameCtx.Import(importSpec.Path.Value)
		}
	}
}

func (p *Parser) parseFunc(funcDecl *ast.FuncDecl) {
	funcName := funcDecl.Name.Name
	fullFuncName := funcName // fullFuncName will contain receiver type when it's a method
	isReceiver := funcDecl.Recv != nil
	receiverName := ""
	if isReceiver {
		switch recvType := funcDecl.Recv.List[0].Type.(type) {
		case *ast.StarExpr:
			if realRecvType, ok := recvType.X.(*ast.Ident); !ok {
				log.Fatalf("unsupported reciver type, too many `*`, at %v", p.fSet.Position(recvType.Pos()))
			} else {
				receiverName = "*" + realRecvType.Name
			}
		case *ast.Ident:
			receiverName = recvType.Name
		}
	}
	if isReceiver {
		log.Debugf("found method (%v)`%v`  from %v to %v", receiverName, funcName,
			p.fSet.Position(funcDecl.Pos()),
			p.fSet.Position(funcDecl.End()),
		)
		fullFuncName = receiverName + "_" + funcName
	} else {
		log.Debugf("found func `%v`  from %v to %v", funcName,
			p.fSet.Position(funcDecl.Pos()),
			p.fSet.Position(funcDecl.End()),
		)
	}
	funcFrame := frame.NewFuncFrame(p.frameCtx.GetInnerName(fullFuncName))
	if funcDecl.Type.Results != nil {
		funcFrame.MarkResult()
	}
	p.parseBlock(funcDecl.Body, fullFuncName, p.getLine(funcDecl.Pos()), funcFrame)
}

func (p *Parser) getLine(pos token.Pos) int {
	return p.fSet.Position(pos).Line
}

func (p *Parser) parseIf(stmt *ast.IfStmt) {
	log.Debugf("%s>>>> found if from pos: %v to %v", p.genPrintPrefix(), p.fSet.Position(stmt.Pos()), p.fSet.Position(stmt.Body.End()))
	p.parseBlock(stmt.Body, "if", p.getLine(stmt.Pos()), frame.NewIfElseFrame(p.frameCtx.GetInnerName("if")))

	if stmt.Else != nil { // else
		switch typedElse := stmt.Else.(type) {
		case *ast.IfStmt: // else-if
			p.parseIf(typedElse)
		case *ast.BlockStmt: // else
			log.Debugf("%s>>>> found if-else from pos: %v to %v", p.genPrintPrefix(), p.fSet.Position(typedElse.Pos()), p.fSet.Position(typedElse.End()))
			p.parseBlock(typedElse, "else", p.getLine(typedElse.Pos()), frame.NewIfElseFrame(p.frameCtx.GetInnerName("else")))
		}
	}
}

// as switch and select has almost the same structure, we can parse them in the same way
func (p *Parser) parseSwitchSelect(unionStmt ast.Stmt) {
	var realType string
	var body *ast.BlockStmt
	switch typed := unionStmt.(type) {
	case *ast.SwitchStmt:
		realType = "switch"
		body = typed.Body
	case *ast.TypeSwitchStmt:
		realType = "typed-switch"
		body = typed.Body
	case *ast.SelectStmt:
		realType = "select"
		body = typed.Body
	default:
		log.Fatalf("not switch or select")
	}

	log.Debugf("%s>>>> found %s from pos: %v to %v", p.genPrintPrefix(), realType, p.fSet.Position(unionStmt.Pos()), p.fSet.Position(body.End()))
	// only `case` and `default` statement can exist in the body of switch and select
	// since we only need to track the case, there is no need to create new frame
	p.level++
	defer func() { p.level-- }()
	for _, stmt := range body.List {
		var caseBody []ast.Stmt
		var headBegin, bodyBegin, bodyEnd token.Pos
		switch typed := stmt.(type) {
		case *ast.CaseClause:
			if typed.List == nil {
				log.Debugf("%s>>>> found default at pos: %v", p.genPrintPrefix(), p.fSet.Position(typed.Pos()))
			} else {
				log.Debugf("%s>>>> found case at pos: %v", p.genPrintPrefix(), p.fSet.Position(typed.Pos()))
			}
			caseBody = typed.Body
			headBegin = typed.Case
			bodyBegin = typed.Case
			bodyEnd = typed.End()
		case *ast.CommClause:
			if typed.Comm == nil {
				log.Debugf("%s>>>> found default at pos: %v", p.genPrintPrefix(), p.fSet.Position(typed.Pos()))
			} else {
				log.Debugf("%s>>>> found case at pos: %v", p.genPrintPrefix(), p.fSet.Position(typed.Pos()))
			}
			caseBody = typed.Body
			headBegin = typed.Case
			bodyBegin = typed.Case
			bodyEnd = typed.End()
		default:
			log.Fatalf("unexpected element in %s body, %v, %v", realType, p.fSet.Position(stmt.Pos()), reflect.TypeOf(stmt))
		}

		newFrame := frame.NewCaseFrame(p.frameCtx.GetInnerName(realType))
		newFrame.SetPosLine(p.getLine(headBegin), p.getLine(bodyBegin), p.getLine(bodyEnd))
		p.frameCtx.Push(newFrame)
		p.parseBlockBody(caseBody, newFrame)
		p.frameCtx.Pop()
	}
}

func (p *Parser) parseStmt(stmt ast.Stmt, currentFrame frame.Frame) {
	switch typed := stmt.(type) {
	case *ast.ReturnStmt:
		// todo: parse function lit in return statement
		currentFrame.SetReturn(p.getLine(typed.Pos()))
		log.Debugf("%s>>>> found return at pos: %v", p.genPrintPrefix(), p.getLine(typed.End()))
	case *ast.IfStmt:
		p.parseIf(typed)
	case *ast.SwitchStmt:
		p.parseSwitchSelect(typed)
	case *ast.SelectStmt:
		p.parseSwitchSelect(typed)
	case *ast.TypeSwitchStmt:
		p.parseSwitchSelect(typed)
	case *ast.LabeledStmt: // todo: check label jump
		p.parseStmt(typed.Stmt, currentFrame)
	case *ast.RangeStmt:
		log.Debugf("%s>>>> found for-range at pos: %v", p.genPrintPrefix(), p.fSet.Position(typed.Pos()))
		p.parseBlock(typed.Body, "for-range", p.getLine(typed.Pos()), frame.NewForFrame(p.frameCtx.GetInnerName("for-range")))
	case *ast.ForStmt:
		log.Debugf("%s>>>> found for at pos: %v", p.genPrintPrefix(), p.fSet.Position(typed.Pos()))
		p.parseBlock(typed.Body, "for", p.getLine(typed.Pos()), frame.NewForFrame(p.frameCtx.GetInnerName("for")))
	case *ast.DeclStmt:
		if genDecl, ok := typed.Decl.(*ast.GenDecl); ok {
			for _, spec := range genDecl.Specs {
				if valueSpec, ok := spec.(*ast.ValueSpec); ok { // var x = func() {}
					for _, value := range valueSpec.Values {
						if funcLit, ok := value.(*ast.FuncLit); ok {
							funcFrame := frame.NewFuncFrame(p.frameCtx.GetInnerName("anonymous-decl-assign"))
							p.parseBlock(funcLit.Body, "anonymous-decl-assign", p.getLine(funcLit.Pos()), funcFrame)
						} else if funcCall, ok := value.(*ast.CallExpr); ok {
							p.parseFuncCallExpr(funcCall, "decl-assign-call")
						}
					}
				}
			}
		}
	case *ast.ExprStmt: // function call
		if funcCall, ok := typed.X.(*ast.CallExpr); ok { // func(){}()
			p.parseFuncCallExpr(funcCall, "call")
		}
	case *ast.AssignStmt: // function call may exist in the right of assignment
		for _, expr := range typed.Rhs {
			if funcCall, ok := expr.(*ast.CallExpr); ok { // x = func()T{return T}()
				p.parseFuncCallExpr(funcCall, "assign-call")
			} else if funcLit, ok := expr.(*ast.FuncLit); ok { // x = func(){}
				funcFrame := frame.NewFuncFrame(p.frameCtx.GetInnerName("anonymous-assign"))
				p.parseBlock(funcLit.Body, "anonymous-assign", p.getLine(funcLit.Pos()), funcFrame)
			}
		}
	case *ast.GoStmt: // go func(){}
		if funcLit, ok := typed.Call.Fun.(*ast.FuncLit); ok {
			log.Debugf("%sfound go func-lit call, at pos: %v", p.genPrintPrefix(), p.fSet.Position(funcLit.Pos()))
			newGoFrame := frame.NewGoFuncFrame(p.frameCtx.GetInnerName("go-anonymous"))
			p.parseBlock(funcLit.Body, "anonymous-go", p.getLine(funcLit.Pos()), newGoFrame)
		} else if ident, ok := typed.Call.Fun.(*ast.Ident); ok {
			// todo: go func()，need some tricks to inject our code
			log.Debugf("%sfound go func call, target `%v` at pos: %v", p.genPrintPrefix(), ident.Name, p.fSet.Position(ident.Pos()))
			newGoFrame := frame.NewGoFuncFrame(p.frameCtx.GetInnerName("go-" + ident.Name))
			newGoFrame.SetPosLine(p.getLine(typed.Pos()), p.getLine(typed.Pos()), p.getLine(typed.Call.End()))
			p.frameCtx.Push(newGoFrame)
			p.frameCtx.Pop() // inject a frame to track goroutine
		}
		// anonymous function may exist as an arguments in a function call: go func(int){}(func()int{}())
		for _, x := range typed.Call.Args {
			if call, ok := x.(*ast.CallExpr); ok {
				p.parseFuncCallExpr(call, "go-args")
			}
		}
	}
}

func (p *Parser) parseBlockBody(stmts []ast.Stmt, currentFrame frame.Frame) {
	for _, stmt := range stmts {
		p.parseStmt(stmt, currentFrame)
	}
}

func (p *Parser) parseFuncCallExpr(callExpr *ast.CallExpr, suffix string) {
	if funcLit, ok := callExpr.Fun.(*ast.FuncLit); ok {
		p.parseBlock(funcLit.Body, "anonymous-"+suffix, p.getLine(funcLit.Pos()),
			frame.NewFuncFrame(p.frameCtx.GetInnerName("anonymous-"+suffix)))
	}
	// the parameter of the function call may also come from an anonymous function,
	// but no one would write it like that :)
	for _, arg := range callExpr.Args {
		if x, ok := arg.(*ast.CallExpr); ok {
			p.parseFuncCallExpr(x, suffix+"-args")
		}
	}
}

func (p *Parser) parseBlock(body *ast.BlockStmt, blockName string, startLine int, blockFrame frame.Frame) {
	p.level++
	defer func() { p.level-- }()

	stmts := body.List

	blockFrame.SetPosLine(startLine, p.getLine(body.Lbrace), p.getLine(body.Rbrace))
	p.frameCtx.Push(blockFrame)
	defer p.frameCtx.Pop()

	p.parseBlockBody(stmts, blockFrame)
}

func (p *Parser) genPrintPrefix() string {
	result := ""
	for i := p.level; i > 0; i-- {
		result += "  "
	}
	return result
}

func (p *Parser) FrameContext() *frame.Context {
	return p.frameCtx
}

// This example illustrates how to remove a variable declaration
// in a Go program while maintaining correct comment association
// using an ast.CommentMap.
func ExampleCommentMap() {
	// src is the input for which we create the AST that we
	// are going to manipulate.
	src := `
// This is the package comment.
package main

// This comment is associated with the hello constant.
const hello = "Hello, World!" // line comment 1

// This comment is associated with the foo variable.
var foo = hello // line comment 2

// This comment is associated with the main function.
func main() {
	fmt.Println(hello) // line comment 3
}
`

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "src.go", src, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	// Create an ast.CommentMap from the ast.File's comments.
	// This helps keeping the association between comments
	// and AST nodes.
	cmap := ast.NewCommentMap(fset, f, f.Comments)

	// Remove the first variable declaration from the list of declarations.
	for i, decl := range f.Decls {
		if gen, ok := decl.(*ast.GenDecl); ok && gen.Tok == token.VAR {
			copy(f.Decls[i:], f.Decls[i+1:])
			f.Decls = f.Decls[:len(f.Decls)-1]
			break
		}
	}

	// Use the comment map to filter comments that don't belong anymore
	// (the comments associated with the variable declaration), and create
	// the new comments list.
	f.Comments = cmap.Filter(f).Comments()

	// Print the modified AST.
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, f); err != nil {
		panic(err)
	}
	fmt.Printf("%s", buf.Bytes())

	// Output:
	// // This is the package comment.
	// package main
	//
	// // This comment is associated with the hello constant.
	// const hello = "Hello, World!" // line comment 1
	//
	// // This comment is associated with the main function.
	// func main() {
	// 	fmt.Println(hello) // line comment 3
	// }
}
