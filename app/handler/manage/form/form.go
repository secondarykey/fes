package form

import (
	"app/datastore"
	"app/logic"

	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"golang.org/x/xerrors"
)

func SetFile(r *http.Request, fs *datastore.FileSet, ft int) error {

	upload, header, err := r.FormFile("file")
	if err != nil {
		if !errors.Is(err, http.ErrMissingFile) {
			return xerrors.Errorf("FromFile() error: %w", err)
		} else {
			return nil
		}
	}
	//ファイルデータの作成
	defer upload.Close()

	b, flg, err := logic.ConvertImage(upload)
	if err != nil {
		return xerrors.Errorf("convertImage() error: %w", err)
	}

	var f datastore.File
	var fd datastore.FileData

	fs.Name = header.Filename
	f.Size = int64(len(b))
	f.Type = ft

	mime := header.Header["Content-Type"][0]
	if flg {
		mime = "image/jpeg"
	}

	fd.Content = b
	fd.Mime = mime

	fs.File = &f
	fs.FileData = &fd

	return nil
}

func SetPage(r *http.Request, ps *datastore.PageSet, id string) error {

	p := ps.Page
	pd := ps.PageData

	ver := r.FormValue("version")
	p.SetTargetVersion(ver)

	p.Name = r.FormValue("pageName")
	p.Parent = r.FormValue("parentID")
	seqBuf := r.FormValue("seq")
	v, err := strconv.Atoi(seqBuf)
	if err != nil {
		return xerrors.Errorf("Seq parse error: %w", err)
	}

	p.Seq = v
	p.Description = r.FormValue("pageDescription")
	p.SiteTemplate = r.FormValue("siteTemplateID")
	p.PageTemplate = r.FormValue("pageTemplateID")

	paging, err := strconv.Atoi(r.FormValue("paging"))
	if err == nil {
		p.Paging = paging
	} else {
		//TODO error
	}

	flag := r.FormValue("publish")
	if flag == "on" {
		p.Deleted = false
	} else {
		p.Deleted = true
	}

	if p.SiteTemplate == "" || p.PageTemplate == "" {
		//ページは選択しなくても表示はできるのでOK
		return errors.New("Error:Select Site Template")
	}

	pd.Content = []byte(r.FormValue("pageContent"))

	return nil
}

func SetSite(r *http.Request, site *datastore.Site) error {

	ver := r.FormValue("version")
	site.SetTargetVersion(ver)

	site.Name = r.FormValue("name")
	site.Description = r.FormValue("description")
	site.Root = r.FormValue("rootPage")
	site.ManageURL = r.FormValue("manageURL")
	site.Managers = strings.Split(r.FormValue("manager"), ",")

	return nil
}

func SetTemplate(r *http.Request, ts *datastore.TemplateSet) error {

	var err error

	vars := mux.Vars(r)
	id := vars["key"]

	template := ts.Template
	templateData := ts.TemplateData
	tmpKey := datastore.GetTemplateKey(id)
	tmpDataKey := datastore.GetTemplateDataKey(id)

	ver := r.FormValue("version")
	template.SetTargetVersion(ver)

	template.LoadKey(tmpKey)
	templateData.LoadKey(tmpDataKey)

	template.Name = r.FormValue("name")
	template.Type, err = strconv.Atoi(r.FormValue("templateType"))
	if err != nil {
		return xerrors.Errorf("TemplateType Atoi() error: %w", err)
	}

	templateData.Content = []byte(r.FormValue("template"))

	return nil
}

func SetVariable(r *http.Request, vs *datastore.VariableSet) error {

	vari := vs.Variable
	variData := vs.VariableData

	val := r.FormValue("variableData")
	ver := r.FormValue("version")

	vari.SetTargetVersion(ver)
	variData.Content = []byte(val)

	return nil
}

func SetDraftSet(r *http.Request, set *datastore.DraftSet) error {

	draft := set.Draft

	vars := mux.Vars(r)
	id := vars["key"]

	ver := r.FormValue("version")
	draft.SetTargetVersion(ver)

	draft.LoadKey(datastore.GetDraftKey(id))
	draft.Name = r.FormValue("name")

	ids := r.FormValue("ids")
	versions := r.FormValue("versions")
	updateds := r.FormValue("updateds")

	if ids == "" {
		return nil
	}

	idSlice := strings.Split(ids, ",")
	verSlice := strings.Split(versions, ",")
	upSlice := strings.Split(updateds, ",")

	pages := make([]*datastore.DraftPage, len(idSlice))
	for idx, id := range idSlice {
		var page datastore.DraftPage
		page.LoadKey(datastore.GetDraftPageKey(id))
		page.Seq = idx + 1
		page.SetTargetVersion(verSlice[idx])
		page.PublishUpdate = parseBool(upSlice[idx])

		pages[idx] = &page
	}

	set.Pages = pages

	return nil
}

func parseBool(v string) bool {
	if v == "true" {
		return true
	}
	return false
}
