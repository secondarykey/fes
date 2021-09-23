package internal

import (
	"app/api"
	"app/datastore"
	"fmt"
	"strconv"

	"embed"
	"html/template"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
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

func ViewManage(w http.ResponseWriter, dto interface{}, name string) error {

	funcMap := template.FuncMap{
		"plane":               api.ConvertString,
		"html":                api.ConvertHTML,
		"convertDate":         api.ConvertDate,
		"convertSize":         api.ConvertSize,
		"convertTemplateType": api.ConvertTemplateType,
	}

	tmpl := template.New(api.SiteTemplateName).Funcs(funcMap)
	err := writeTemplate(w, tmpl, dto, "manage/layout.tmpl", "manage/"+name)
	if err != nil {
		return xerrors.Errorf("writeTemplate() error: %w", err)
	}
	return nil
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

func CreateFormTemplate(r *http.Request) (*datastore.TemplateSet, error) {

	vars := mux.Vars(r)
	id := vars["key"]

	tmpKey := datastore.SetTemplateKey(id)
	tmpDataKey := datastore.CreateTemplateDataKey(id)

	template := datastore.Template{}
	templateData := datastore.TemplateData{}

	ver := r.FormValue("version")
	version, err := strconv.Atoi(ver)
	if err != nil {
		return nil, xerrors.Errorf("version strconv.Atoi() error: %w", err)
	}

	if version > 0 {
		template.TargetVersion = fmt.Sprintf("%d", version)
	}

	template.LoadKey(tmpKey)
	templateData.LoadKey(tmpDataKey)

	template.Name = r.FormValue("name")
	template.Type, err = strconv.Atoi(r.FormValue("templateType"))
	if err != nil {
		return nil, xerrors.Errorf("TemplateType Atoi() error: %w", err)
	}

	templateData.Content = []byte(r.FormValue("template"))

	var ts datastore.TemplateSet

	ts.Template = &template
	ts.TemplateData = &templateData

	return &ts, nil
}
