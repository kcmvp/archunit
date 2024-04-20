// nolint
package internal

import (
	"fmt"
	"reflect"
	"strings"
)

type Method struct {
	name        string
	ofType      string
	returnTypes []reflect.Type
}

func (m Method) Public() bool {
	first := []rune(m.name)[0]
	return strings.ToUpper(string(first)) == string(first)
}

func (m Method) Name() string {
	return m.name
}

func (m Method) String() string {
	return fmt.Sprintf("%s.%s", m.ofType, m.name)
}
