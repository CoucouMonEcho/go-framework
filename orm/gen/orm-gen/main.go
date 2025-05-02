package main

import (
	_ "embed"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	//"html/template"
	"text/template"
)

//func main() {
//	f, err := os.OpenFile("testdata/user.gen.go")
//	if err != nil {
//		fmt.
//	}
//}

//go:embed tpl.gohtml
var genOrm string

func gen(w io.Writer, srcFile string) error {
	// ast
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, srcFile, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	s := &SingleFileEntryVisitor{}
	ast.Walk(s, f)
	file := s.Get()
	// template
	tpl := template.New("gen-orm")
	tpl, err = tpl.Parse(genOrm)
	if err != nil {
		return err
	}
	return tpl.Execute(w, file)
}
