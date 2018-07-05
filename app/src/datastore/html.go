package datastore

import (
	"api"

	"fmt"
	"time"
	"io"
	"html/template"
	"net/http"
	"bytes"

	"golang.org/x/net/context"
	"github.com/knightso/base/gae/ds"
	kerr "github.com/knightso/base/errors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"

)

type HTML struct {
	Content []byte
	ds.Meta
}

const KindHTMLName = "HTML"

func createHTMLKey(r *http.Request,id string) *datastore.Key {
	c := appengine.NewContext(r)
	return datastore.NewKey(c,KindHTMLName,id,0,nil)
}

func GetHTML(r *http.Request,id string) (*HTML,error) {

	var err error
	key := createHTMLKey(r,id)
	html := HTML{}

	c := appengine.NewContext(r)
	err = ds.Get(c,key,&html)
	if err != nil {
		if kerr.Root(err) != datastore.ErrNoSuchEntity {
			return nil, err
		} else {
			return nil,nil
		}
	}
	return &html,nil
}

func PutHTML(r *http.Request,id string) error {

	c := appengine.NewContext(r)
	page,err := SelectPage(r,id)
	if err != nil {
		return err
	}
	if page == nil {
		return fmt.Errorf("page not found[%s]",id)
	}

	html,err:= GetHTML(r,id)
	if err != nil {
		return err
	}

	key := createHTMLKey(r,id)
	if html == nil {
		html = &HTML{}
		html.SetKey(key)
	}

	var buf []byte
	w := bytes.NewBuffer(buf)
	err = GenerateHTML(w,r,id,true)
	if err != nil {
		return err
	}

	html.Content = w.Bytes()

	option := &datastore.TransactionOptions{XG: true}
	return datastore.RunInTransaction(c, func(ctx context.Context) error {
		err = ds.Put(c, html)
		if err != nil {
			return err
		}

		page.Publish = time.Now()
		err = ds.Put(c,page)
		if err != nil {
			return err
		}
		return nil
	},option)
}

func RemoveHTML(r *http.Request,id string) error {

	c := appengine.NewContext(r)
	page,err := SelectPage(r,id)
	if err != nil {
		return err
	}
	if page == nil {
		return fmt.Errorf("page not found[%s]",id)
	}

	option := &datastore.TransactionOptions{XG: true}
	return datastore.RunInTransaction(c, func(ctx context.Context) error {
		key := createHTMLKey(r, id)
		err = ds.Delete(c, key)
		if err != nil {
			return err
		}

		page.Publish = time.Time{}
		err = ds.Put(c,page)
		if err != nil {
			return err
		}
		return nil
	}, option)
}

type Public struct {
	request *http.Request
	manage  bool
}

func GenerateHTML(w io.Writer, r *http.Request, id string,mng bool) error {

	var err error

	page, err := SelectPage(r, id)
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

	site := GetSite(r)
	//テンプレートを取得
	siteTmp, err := SelectTemplateData(r, page.SiteTemplate)
	if err != nil {
		return err
	}
	pageTmp, err := SelectTemplateData(r, page.PageTemplate)
	if err != nil {
		return err
	}

	pData, err := SelectPageData(r, id)
	if err != nil {
		return err
	}
	children, err := SelectChildPages(r,id,0,mng)
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
		Site     *Site
		Page     *Page
		PageData *PageData
		Content  string
		Children []Page
		Top      string
		Dir      string
	}{site, page, pData,string(pData.Content), children,top,dir}

	err = tmpl.Execute(w, dto)
	if err != nil {
		return err
	}

	return nil
}

func (p Public) list(id string) []Page {
	pages, err := SelectChildPages(p.request, id,10,p.manage)
	if err != nil {
		return make([]Page, 0)
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