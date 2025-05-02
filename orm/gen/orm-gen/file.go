package main

import (
	"go/ast"
)

type SingleFileEntryVisitor struct {
	file *FileVisitor
}

func (s *SingleFileEntryVisitor) Get() *File {
	return &File{
		Package: s.file.Package,
		Imports: s.file.Imports,
	}
}

func (s *SingleFileEntryVisitor) Visit(node ast.Node) (w ast.Visitor) {
	fn, ok := node.(*ast.File)
	if !ok {
		return s
	}
	s.file = &FileVisitor{
		Package: fn.Name.String(),
	}
	return s.file
}

type File struct {
	Package string
	Imports []string
}

type FileVisitor struct {
	Package string
	Imports []string
}

func (f *FileVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch n := node.(type) {
	//case *ast.GenDecl:
	//	if n.Tok == token.IMPORT {
	//		for _, spec := range n.Specs {
	//			f.Imports = append(f.Imports, spec.(*ast.ImportSpec).Path.Value)
	//		}
	//	}
	case *ast.ImportSpec:
		path := n.Path.Value
		if n.Name != nil && n.Name.String() != "" {
			path = n.Name.String() + " " + path
		}
		f.Imports = append(f.Imports, path)
	}
	return f
}
