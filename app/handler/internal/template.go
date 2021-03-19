package internal

import (
	"app/api"
	"io"

	"embed"
	"html/template"
	"io/fs"
	"log"
	"net/http"

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
	tmpl := template.New(names[0])
	return writeTemplate(w, tmpl, dto, names...)
}

func ViewManage(w http.ResponseWriter, dto interface{}, name string) error {

	funcMap := template.FuncMap{
		"plane":               api.ConvertString,
		"html":                api.ConvertHTML,
		"convertDate":         api.ConvertDate,
		"convertSize":         api.ConvertSize,
		"convertTemplateType": api.ConvertTemplateType,
	}

	tmpl := template.New(api.SiteTemplateName).Funcs(funcMap)
	return writeTemplate(w, tmpl, dto, "manage/layout.tmpl", "manage/"+name)
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
