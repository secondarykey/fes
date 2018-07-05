package logic

import (
	"datastore"
	"api"

	"net/http"
	"html/template"
	"fmt"
)

type Public struct {
	request *http.Request
	manage  bool
}

func GenerateHTML(w http.ResponseWriter, r *http.Request, id string,mng bool) error {

	var err error

	page, err := datastore.SelectPage(r, id)
	if err != nil {
		return err
	}
	if page == nil {
		return fmt.Errorf("Page not found[%s]",id)
	}

	p := Public {
		request:r,
		manage:mng,
	}

	dir := "/manage/page/view/"
	top := "/manage/page/view/"

	if !mng {
		if page.Deleted {
			return fmt.Errorf("Page is private[%s]",id)
		}
		dir = "/page/"
		top = "/"
	}

	site := datastore.GetSite(r)
	//テンプレートを取得
	siteTmp, err := datastore.SelectTemplateData(r, page.SiteTemplate)
	if err != nil {
		return err
	}
	pageTmp, err := datastore.SelectTemplateData(r, page.PageTemplate)
	if err != nil {
		return err
	}

	pData, err := datastore.SelectPageData(r, id)
	if err != nil {
		return err
	}
	children, err := datastore.SelectChildPages(r,id,0,mng)
	if err != nil {
		return err
	}

	siteTmpData := string(siteTmp.Content)
	pageTmpData := string(pageTmp.Content)
	siteTmpData = "{{define \"" + api.SITE_TEMPLATE + "\"}}" + "\n" + siteTmpData + "\n" + "{{end}}"
	pageTmpData = "{{define \"" + api.PAGE_TEMPLATE + "\"}}" + "\n" + pageTmpData + "\n" + "{{end}}"

	funcMap := template.FuncMap{
		"html":        api.ConvertHTML,
		"plane":       api.ConvertString,
		"convertDate": api.ConvertDate,
		"list":        p.list,
		"mark":     p.mark,
	}

	//適用する
	tmpl, err := template.New(api.SITE_TEMPLATE).Funcs(funcMap).Parse(siteTmpData)
	if err != nil {
		return err
	}
	tmpl, err = tmpl.Parse(pageTmpData)
	if err != nil {
		return err
	}

	dto := struct {
		Site     *datastore.Site
		Page     *datastore.Page
		PageData *datastore.PageData
		Content  string
		Children []datastore.Page
		Top      string
		Dir      string
	}{site, page, pData,string(pData.Content), children,top,dir}

	err = tmpl.Execute(w, dto)
	if err != nil {
		return err
	}

	return nil
}

func (p Public) list(id string) []datastore.Page {
	pages, err := datastore.SelectChildPages(p.request, id,10,p.manage)
	if err != nil {
		return make([]datastore.Page, 0)
	}
	return pages
}

func (p Public) mark() template.HTML {
	src := ""
	if p.manage {
		src = `<img src="/images/private.png" class="private-mark" />`
	}
	return template.HTML(src)
}