package internal

import (
	"embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"

	"golang.org/x/xerrors"
)

//go:embed _assets/templates
var embTmpl embed.FS
var tmplFs fs.FS

func init() {
	var err error
	tmplFs, err = fs.Sub(embTmpl, "_assets/templates")
	if err != nil {
		log.Printf("%+v", err)
	}
}

func View(w http.ResponseWriter, dto interface{}, names ...string) error {
	funcMap := map[string]interface{}{
		"env": os.Getenv,
	}
	tmpl := template.New(names[0]).Funcs(funcMap)
	return writeTemplate(w, tmpl, dto, names...)
}

func writeTemplate(w io.Writer, root *template.Template, dto interface{}, names ...string) error {

	tmpl, err := root.ParseFS(tmplFs, names...)
	if err != nil {
		return xerrors.Errorf("ParseFS() error: %w", err)
	}

	err = tmpl.Execute(w, dto)
	if err != nil {
		return xerrors.Errorf("template Execute() error: %w", err)
	}
	return nil
}

func WriteTemplate(w io.Writer, dto interface{}, names ...string) error {
	root := template.New(names[0])
	return writeTemplate(w, root, dto, names...)
}
