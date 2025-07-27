package template_demo

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"html/template"
	"testing"
)

func TestHelloWorld(t *testing.T) {
	tpl := template.New("hello-world")
	// . means this
	// tpl, err := tpl.Parse(`Hello, {{ .Name}}`)

	// tpl, err := tpl.Parse(`Hello, {{index.0}}`) error
	// tpl, err := tpl.Parse(`Hello, {{.}}`)
	tpl, err := tpl.Parse(`Hello, {{index . 0}}`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	// type user struct {
	// 	 Name string
	// }
	// err = tpl.Execute(buffer, user{Name: "John"})
	// err = tpl.Execute(buffer, map[string]any{"Name": "John"})
	// err = tpl.Execute(buffer, 123)

	err = tpl.Execute(buffer, []string{"John"})
	require.NoError(t, err)
	assert.Equal(t, `Hello, John`, buffer.String())
}

func TestFuncCall(t *testing.T) {
	tpl := template.New("hello-world")
	// . means this

	tpl, err := tpl.Parse(`
len: {{len .Slice}}
{{printf "%.2f" 1.234}}
Hello, {{.Hello "Tom" "John"}}
`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	// type user struct {
	// 	 Name string
	// }
	// err = tpl.Execute(buffer, user{Name: "John"})
	// err = tpl.Execute(buffer, map[string]any{"Name": "John"})
	// err = tpl.Execute(buffer, 123)

	err = tpl.Execute(buffer, funcCall{Slice: []string{"1", "2", "3"}})
	require.NoError(t, err)
	assert.Equal(t, `
len: 3
1.23
Hello, Tom, John
`, buffer.String())
}

type funcCall struct {
	Slice []string
}

func (f funcCall) Hello(first string, second string) string {
	return fmt.Sprintf("%s, %s", first, second)
}

func TestForLoop(t *testing.T) {
	tpl := template.New("hello-world")
	// . means this

	tpl, err := tpl.Parse(`
{{- range $idx, $elem := .Slice}}
{{- .}}
{{$idx}}-{{$elem}}
{{end}}
`)
	require.NoError(t, err)
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, funcCall{Slice: []string{"a", "b"}})
	require.NoError(t, err)
	assert.Equal(t, `a
0-a
b
1-b

`, buffer.String())
}
