package frame

import (
	"fmt"
	"github.com/itchyny/base58-go"
	log "github.com/sirupsen/logrus"
	"hash/fnv"
)

type baseEnv struct {
	filename      string
	filenameConst string
	prefix        string
	pointVarIndex int
	funcIndex     int

	funcEnvStack    [128]funcEnv
	funcEnvStackTop int
}

type funcEnv struct {
	GoroutineIDVarName string // variable name for current goroutine id
}

func NewBaseEnv(filename string) *baseEnv {
	var encoder = fnv.New32a()
	_, _ = encoder.Write([]byte(filename))
	sum := encoder.Sum32()
	prefix := fmt.Sprintf("_%s", string(base58.FlickrEncoding.EncodeUint64(uint64(sum))))
	return &baseEnv{
		filename:      filename,
		filenameConst: prefix + "_fName",
		prefix:        prefix,
	}
}

func (e *baseEnv) genPointVarName() string {
	e.pointVarIndex++
	return fmt.Sprintf("%s_e%d", e.prefix, e.pointVarIndex)
}

func (e *baseEnv) genGoIDVarName() string {
	e.funcIndex++
	return fmt.Sprintf("%s_g%d", e.prefix, e.funcIndex)
}

func (e *baseEnv) NewFuncEnv() {
	if e.funcEnvStackTop == 128 {
		log.Fatal("too many levels")
	}
	e.funcEnvStack[e.funcEnvStackTop] = funcEnv{e.genGoIDVarName()}
	e.funcEnvStackTop++
}

func (e *baseEnv) PopFuncEnv() {
	if e.funcEnvStackTop == 0 {
		return
	}
	e.funcEnvStackTop--
}

func (e *baseEnv) GetCurrentGoIDVarName() string {
	if e.funcEnvStackTop == 0 {
		log.Fatal("func env was corrupted")
		return ""
	}
	return e.funcEnvStack[e.funcEnvStackTop-1].GoroutineIDVarName
}

func (e *baseEnv) GetLastGoIDVarName() string {
	if e.funcEnvStackTop < 1 {
		log.Fatal("func env was corrupted")
		return ""
	}
	fmt.Println(e.funcEnvStackTop)
	return e.funcEnvStack[e.funcEnvStackTop-2].GoroutineIDVarName
}

func (e *baseEnv) genPoint(varName string, path string) string {
	return fmt.Sprintf("var %s = %s\n", varName, genSDKFunCallWithArgs("NewE", e.filenameConst, wrapString(path)))
}

func (e *baseEnv) genCall(resultVarName string, varName string) string {
	return fmt.Sprintf("var %s = %s", resultVarName, genSDKFunCallWithArgs("Call", varName))
}

func (e *baseEnv) genCollect(varName string) string {
	return genSDKFunCallWithArgs("C", e.GetCurrentGoIDVarName(), varName)
}

func (e *baseEnv) genBind() string {
	return genSDKFunCallWithArgs("Bind", e.GetLastGoIDVarName())
}
