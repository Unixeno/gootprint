package frame

import "fmt"

const SDKPackage = `"github.com/Unixeno/gootprint/sdk"` // trace sdk used in import statement
const SDKPackagePrefix = "sdk."                         // the prefix to access trace sdk package

func genLineCode(method, args string) string {
	return fmt.Sprintf("%s%s(%s);", SDKPackagePrefix, method, args)
}

func genLineCodeWithStringArg(method, arg string) string {
	return fmt.Sprintf("%s%s(\"%s\");", SDKPackagePrefix, method, arg)
}
