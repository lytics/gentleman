package utils

import (
	"reflect"
	"runtime"
	"strings"
)

const (
	equal   = "\n%s:%d: should be == \n%s \thave: (%T) %+v\n\twant: (%T) %+v"
	unequal = "\n%s:%d: should be != \n%s \thave: (%T) %+v\n\tand : (%T) %+v"
)

// Errorf is satisfied by testing.T and testing.B.
type Errorf interface {
	Errorf(format string, args ...interface{})
}

// Fatalf is satisfied by testing.T and testing.B.
type Fatalf interface {
	Fatalf(format string, args ...interface{})
}

// Equal calls t.Fatal to abort the test when have != want.
func Equal(t Fatalf, have, want interface{}) {
	if !reflect.DeepEqual(have, want) {
		file, line := caller()
		t.Fatalf(equal, file, line, "", have, have, want, want)
	}
}

// NotEqual calls t.Fatal to abort the test when have == want.
func NotEqual(t Fatalf, have, want interface{}) {
	if reflect.DeepEqual(have, want) {
		file, line := caller()
		t.Fatalf(unequal, file, line, "", have, have, want, want)
	}
}

// returns file and line two stack frames above its invocation
func caller() (file string, line int) {
	var ok bool
	_, file, line, ok = runtime.Caller(2)
	if !ok {
		file = "???"
		line = 1
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}
	return
}
