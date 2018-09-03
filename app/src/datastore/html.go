package datastore

import (
	"api"
	"fmt"
	"time"
	"net/http"
	"bytes"

	"golang.org/x/net/context"
	"github.com/knightso/base/gae/ds"
	kerr "github.com/knightso/base/errors"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
	"html/template"
	"strconv"
)

type HTML struct {
	Content  []byte
	Children int
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

	var err error

	page, err := SelectPage(r, id, -1)
	if err != nil {
		return err
	}

	dtos,err := NewDtos(r,page,1,false)
	if err != nil {
		return err
	}

	tmpl,err := createTemplate(r,page,false)
	if err != nil {
		return err
	}

	htmls := make([]*HTML,len(dtos))
	keys := make([]*datastore.Key,len(dtos))
	for idx,dto := range dtos {
		var buf []byte
		w := bytes.NewBuffer(buf)
		//TODO CHILDREN 一回だけ検索
		err = tmpl.Execute(w,dto)
		if err != nil {
			return err
		}

		realId := id
		if idx > 0 {
			realId = fmt.Sprintf("%s?page=%d",id,idx + 1)
		}

		html,err:= GetHTML(r,realId)
		if err != nil {
			return err
		}
		if html == nil {
			key := createHTMLKey(r,realId)
			html = &HTML{}
			html.SetKey(key)
		}
		html.Content = w.Bytes()
		html.Children = len(dtos)
		htmls[idx] = html
		keys[idx] = html.GetKey()
	}

	c := appengine.NewContext(r)
	option := &datastore.TransactionOptions{XG: true}
	return datastore.RunInTransaction(c, func(ctx context.Context) error {

		err = ds.PutMulti(c, keys, htmls)
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
	page,err := SelectPage(r,id,-1)
	if err != nil {
		return err
	}
	if page == nil {
		return fmt.Errorf("page not found[%s]",id)
	}

	//TODO ページ数個削除

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

		//TODO 固定はまずいけど、Referenceの場所なので
		//TODO ページングがいらない可能性が高い
		dtos,err := NewDtos(r,&elm,0,false)
		if err != nil {
			return err
		}

		tmpl,err := createTemplate(r,&elm,false)
		if err != nil {
			return err
		}
		err = tmpl.Execute(w,dtos[0])
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


//TOOL

type HTMLDto struct {
	Site     *Site
	Page     *Page
	PageData *PageData
	Content  string
	Children []Page
	Top      string
	Dir      string
	Prev      string
	Next      string
}

func WriteManageHTML(w http.ResponseWriter, r *http.Request,id string,p int) (error) {

	var err error
	page, err := SelectPage(r, id,-1)
	if err != nil {
		return err
	}
	if page == nil {
		return fmt.Errorf("Page not found[%s]", id)
	}
	//テンプレートの作成
	tmpl,err := createTemplate(r,page,true)
	if err != nil {
		return err
	}
	//DTOの作成
	dtos,err := NewDtos(r,page,p,true)
	if err != nil {
		return err
	}

	//書き込み
	err = tmpl.Execute(w, dtos[0])
	if err != nil {
		return err
	}
	return err
}

func NewDtos(r *http.Request, page *Page,pageNum int,mng bool) ([]*HTMLDto,error) {

	id := page.Key.StringID()
	site,err := SelectSite(r,-1)
	if err != nil {
		return nil,err
	}

	pData, err := SelectPageData(r, id)
	if err != nil {
		return nil,err
	}

    content := string(pData.Content)
	dto := HTMLDto {
		Site     : site,
		Page     : page,
		PageData : pData,
		Content  : content,
	}

	dir := "/manage/page/view/"
	top := "/manage/page/view/"
	if !mng {
		if page.Deleted {
			return nil,fmt.Errorf("Page is private[%s]",id)
		}
		dir = "/page/"
		top = "/"
	}

	childNum := 0
	//本番時には複数件取得して展開
	if mng && page.Paging > 0 {
		childNum = page.Paging
	}

	children, err := SelectChildPages(r,id,pageNum,childNum,mng)
	if err != nil {
		return nil,err
	}

	leng := len(children)
	last := 1
	dtoNum := 1

	if page.Paging > 0 {
		last = leng / page.Paging + 1
		mod := leng % page.Paging

		if mod == 0 {
			last -= 1
		}

		if !mng {
			dtoNum = last
		}
	}

	dtos := make([]*HTMLDto,dtoNum)

	for idx := 0 ; idx < dtoNum; idx++ {

		cp := dto
		prevId := ""
		nextId := ""
		pNum := idx + 1

		if mng {
			pNum = pageNum
		}

		if last != pNum || mng {
			nextId = id + "?page=" + strconv.Itoa(pNum+1)
		}

		if pNum > 1 {
			prevId = id
			if pNum > 2 {
				prevId += "?page=" + strconv.Itoa(pNum-1)
			}
		}

		start := 0
		end := len(children)
		if !mng && page.Paging > 0 {
			start = idx * page.Paging

			end = start + page.Paging
			if end > leng {
				end = leng
			}
		}

		cpChildren := children[start:end]

		cp.Top = top
		cp.Dir = dir
		cp.Children = cpChildren
		cp.Prev = prevId
		cp.Next = nextId

		dtos[idx] = &cp
	}

	return dtos,nil
}

func createTemplate(r *http.Request,page *Page,mng bool) (*template.Template,error){
	//テンプレートを取得
	siteTmp, err := SelectTemplateData(r, page.SiteTemplate)
	if err != nil {
		return nil,err
	}
	pageTmp, err := SelectTemplateData(r, page.PageTemplate)
	if err != nil {
		return nil,err
	}
	siteTmpData := string(siteTmp.Content)
	pageTmpData := string(pageTmp.Content)
	siteTmpData = "{{define \"" + api.SiteTemplateName + "\"}}" + "\n" + siteTmpData + "\n" + "{{end}}"
	pageTmpData = "{{define \"" + api.PageTemplateName + "\"}}" + "\n" + pageTmpData + "\n" + "{{end}}"

	pub := Public {
		request:r,
		manage:mng,
	}
	//適用する
	tmpl, err := template.New(api.SiteTemplateName).Funcs(pub.funcMap()).Parse(siteTmpData)
	if err != nil {
		return nil,err
	}
	tmpl, err = tmpl.Parse(pageTmpData)
	if err != nil {
		return nil,err
	}
	return tmpl,nil
}


//Public template object
type Public struct {
	request *http.Request
	manage  bool
}

func (p Public) list(id string,num int) []Page {
	//1ページ目固定
	pages, err := SelectChildPages(p.request,id,0,num,p.manage)
	if err != nil {
		return make([]Page, 0)
	}
	return pages
}

var privateMark = `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAENklEQVR4XuVb240cMQzLtZIKkv6LSCpIKwkmgA9arSiSsuceuPzkHrYs0RTFGdy+fPvi/16m9f/58fPv99+/Xq7/pzGcfeusfGb8/vraiXmttTe8VcG5EBVsFwQZgPcq3L3RtV4FggLACo8HrbWZlijG2lvtmxZ+7avaBMWDADiF7yR7Ym8E8IpX5Y4YUQLAbuxE0nfGQAJdgSADoPbUnYU5sdVLfABA3eQk8t5rWTtQAD7bzWfAtwD47MUvMDoQXhmQF02Kn8TIghUd36n2eRMA0ChSgOxGbrV/geYAhEA4wgB08xmULuHODFVmK8aagnztKwFQAiLD0RXJ4sabZaBW5yjxM3D/AXB7F9GJucdlUxUmVA9DHeiKdlR12gCwItktsL5lerDjVcYAdIeyOcsKrn7fxdw5bwwAK6IaZWyP2gaZUbcAkHvLofFOQgwEVrz7NggyIDsmFYCdfqzOZGJc/X4yatfZD2MwskAFgDFHSY4VHVlSGSDHFN2iAahItzBmbE602nEAYvExOHrVhWwtKx7Nf4epyO/YPmAFyu/dKuPCbk3VEOYN1IlzGwOQa1MSy8xBfZ5ZougLOj/uHTEgK7hCYZZMF4MxSQG6yvl1CjiClVW5OtzpTbU4dR3zFnEEXl9vMyA+hKhJRpp3e5DITluuEuoxAJ2A5f6sCo6TAjm6Ks4pxq3zjwNwJVhNiGo0onbKa1dMZIocQcwsGAHQjSUEQEfbSoOYr2DgVUxpx6DzYDEFoDqDxXKnQ8UWNLUepoDyRqULlBF3BHHCmmvP0hZ1ikEGrGB5RDAaMdYoDk5hQDdWHRAkANgMz73p3jQCmcVFeXXTSLnAp7fCDgBM8VHLxD7tblBVd/USItPHYxAlXIlVtxY9MbomZ1J8ZKE1BhFNOy2It8ho7hbviHLWucXCIwDkVqiMTFwT20wRyqqXq6kzad8RADsF5iRdEVPAQGsqc2UDoKgxKlJ5I9QZmZ3ic7vYIrgELSYRf6Yqdt5fFcWofAKIFcNiQNV3OzfmmBg00x2wqks6BkAsRknqhPhVrOzYccQJIgorRXf0Z7b6JPsiE47/fQADYtf753nOzqtAjw9+2wDklx8sodyHCiDRue2MTakFHFGrpgDaX83gyp11PeyMVySa8eev7wOQTWTjpkK0MyGqD0AxKrY54xcywHmgcOc4U2q1BXLxuS3U1ruVAU77ZGfG2LZin7qsByd4ogXYDSgFdpqCNESNm9c9ATABwdEAJdE71yAAv8wfSyOj9fR5gWmP3Xl709hKLfADE67BmSZ5175MeTQu7c8MnRC7u4pmxqrK3QYgjqOPCIZrlduPzXXBVAOTjUelxtU0OXU2u6Ttzw26dEbvE904ynpW/BWDAlA5ts6Wsj5UEs+O0mWbUvjKQwYgJ37K4yNAYhEMgPh8rwK8DYB70Edd/w+BjJCMxE5c7wAAAABJRU5ErkJggg==`
func (p Public) mark() template.HTML {
	src := ""
	if p.manage {
		src = `<img src="`  + privateMark + `" style="position: fixed; display: block; right: 0; bottom: 0; margin-right: 40px; margin-bottom: 40px; z-index: 900;" />`
	}
	return template.HTML(src)
}


func (p Public) funcMap() template.FuncMap {
	return template.FuncMap{
		"html":        api.ConvertHTML,
		"plane":       api.ConvertString,
		"convertDate": api.ConvertDate,
		"list":        p.list,
		"mark":        p.mark,
		"templateContent" : p.ConvertTemplate,
	}
}
//Contentの変換時にテンプレートを実現する
func (p Public) ConvertTemplate(data string) template.HTML {

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


