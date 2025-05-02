package ast

import (
	"github.com/stretchr/testify/require"
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

func TestPrintVisitor_Visit(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "src.go", `package ast

import (
	"fmt"
	"go/ast"
	"reflect"
)

type PrintVisitor struct {
}

func (p PrintVisitor) Visit(node ast.Node) (w ast.Visitor) {
	if node == nil {
		fmt.Println(nil)
		return p
	}
	typ := reflect.TypeOf(node)
	val := reflect.ValueOf(node)
	for typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
		val = val.Elem()
	}
	fmt.Printf("type: %s, value: %+v", typ.Name(), val.Interface())
	return p
}
`, parser.ParseComments)
	require.NoError(t, err)
	ast.Walk(&PrintVisitor{}, f)
}
