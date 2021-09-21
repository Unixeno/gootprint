package frame

import (
	"fmt"
	"strconv"
	"strings"
)

const SDKPackage = `_g_sdk "github.com/Unixeno/gootprint/sdk"` // trace sdk used in import statement
const SDKPackagePrefix = "_g_sdk."                             // the prefix to access trace sdk package

func wrapString(src string) string {
	return strconv.Quote(src)
}

func genSDKFunCallWithArgs(method string, args ...string) string {
	formattedArgs := ""
	if len(args) > 0 {
		formattedArgs = strings.Join(args, ", ")
	}
	return fmt.Sprintf("%s%s(%s);", SDKPackagePrefix, method, formattedArgs)
}
