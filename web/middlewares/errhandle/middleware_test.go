package errhandle

import (
	"bytes"
	"github.com/CoucouMonEcho/go-framework/web"
	"github.com/stretchr/testify/require"
	"html/template"
	"net/http"
	"testing"
)

func TestNewMiddlewareBuilder(t *testing.T) {
	page := `
<html>
	<body>
		<h1>404 Not Found</h1>
	</body>
</html>
`
	tpl, err := template.New("404").Parse(page)
	if err != nil {
		t.Fatal(err)
	}
	buffer := &bytes.Buffer{}
	err = tpl.Execute(buffer, nil)
	if err != nil {
		t.Fatal(err)
	}
	server := web.NewHTTPServer(web.ServerWithMiddlewares(NewMiddlewareBuilder().
		RegisterError(http.StatusNotFound, buffer.Bytes()).Build()))
	err = server.Start(":8081")
	require.NoError(t, err)
}
