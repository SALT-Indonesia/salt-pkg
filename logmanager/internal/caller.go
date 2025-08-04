package internal

import (
	"runtime"
	"strings"
)

// GetCaller returns the name of the function that invoked the caller of this function.
func GetCaller() string {
	pc, _, _, _ := runtime.Caller(2)
	callerFuncName := CallerFuncName(runtime.FuncForPC(pc))
	return CallerName(callerFuncName)
}

func CallerFuncName(f *runtime.Func) string {
	if nil == f {
		return ""
	}
	return f.Name()
}

func CallerName(funcName string) string {
	if len(funcName) == 0 {
		return ""
	}
	callerNames := strings.Split(funcName, ".")
	return callerNames[len(callerNames)-1]
}
