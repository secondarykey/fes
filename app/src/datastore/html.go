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

	html,err:= GetHTML(r,id)
	if err != nil {
		return err
	}

	if html == nil {
		key := createHTMLKey(r,id)
		html = &HTML{}
		html.SetKey(key)
	}

	var buf []byte
	w := bytes.NewBuffer(buf)
	page,err := GenerateHTML(w,r,id,false)
	if err != nil {
		return err
	}

	html.Content = w.Bytes()

	c := appengine.NewContext(r)
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

func PutHTMLs(r *http.Request,pages []Page) error {

	var err error
	//HTMLとを作成
	//HTMLキーを作成
	htmlData := make([][]byte,0)
	keys     := make([]*datastore.Key,0)
	for _,elm := range pages {
		if elm.Deleted {
			continue
		}
		var buf []byte
		w := bytes.NewBuffer(buf)
		err := createHTMLData(w,r,&elm,false)
		if err != nil {
			return err
		}
		htmlData = append(htmlData,w.Bytes())
		keys = append(keys,createHTMLKey(r,elm.Key.StringID()))
	}

	c := appengine.NewContext(r)
	htmls := make([]HTML,len(keys))

	err = ds.GetMulti(c,keys,htmls)
	if err != nil {
		if berr,ok := err.(*kerr.BaseError); ok {
			err = berr.Cause()
			_,flag := err.(appengine.MultiError)
			if !flag {
				return err
			}
		} else {
			return err
		}
	}

	for idx,_ := range htmls {
		if err != nil {
			multi := err.(appengine.MultiError)
			if m := multi[idx] ; m != nil && m != datastore.ErrNoSuchEntity {
				return err
			}
		}
		htmls[idx].SetKey(keys[idx])
		htmls[idx].Content = htmlData[idx]
	}

	return ds.PutMulti(c,keys,htmls)
}

type Public struct {
	request *http.Request
	manage  bool
}

func GenerateHTML(w io.Writer, r *http.Request, id string,mng bool) (*Page,error) {
	var err error
	page, err := SelectPage(r, id)
	if err != nil {
		return nil, err
	}
	if page == nil {
		return nil, fmt.Errorf("Page not found[%s]", id)
	}
	err = createHTMLData(w,r,page,mng)
	return page,err
}

func createHTMLData(w io.Writer, r *http.Request, page *Page,mng bool) (error) {

	p := Public {
		request:r,
		manage:mng,
	}

	dir := "/manage/page/view/"
	top := "/manage/page/view/"

	id := page.Key.StringID()

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

	//適用する
	tmpl, err := template.New(api.SITE_TEMPLATE).Funcs(p.funcMap()).Parse(siteTmpData)
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

func (p Public) list(id string,num int) []Page {
	pages, err := SelectChildPages(p.request, id,num,p.manage)
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

func (p Public) funcMap() template.FuncMap {
	return template.FuncMap{
		"html":        api.ConvertHTML,
		"plane":       api.ConvertString,
		"convertDate": api.ConvertDate,
		"list":        p.list,
		"mark":     p.mark,
		"templateContent" : p.ConvertTemplate,
	}
}

//Contentの変換時にテンプレートを実現する
func (p Public)ConvertTemplate(data string) template.HTML {

	tmpl,err := template.New("").Funcs(p.funcMap()).Parse(data)
	if err != nil {
		return template.HTML(fmt.Sprintf("Template Parse Error[%s]",err))
	}

	dto := struct {
		Dir string
		Top string
	} {"/page/","/"}
	if p.manage {
		dto.Dir = "/manage/page/view/"
		dto.Top = "/manage/page/view/"
	}

	buf := bytes.NewBuffer(make([]byte,0,len(data) + 30))
	err = tmpl.Execute(buf,dto)
	if err != nil {
		return template.HTML(fmt.Sprintf("Template Execute Error[%s]",err))
	}

	return template.HTML(buf.String())
}