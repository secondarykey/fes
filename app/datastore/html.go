package datastore

import (
	"app/api"
	"context"
	"errors"

	"bytes"
	"fmt"
	"net/http"
	"time"

	"html/template"
	"strconv"

	"golang.org/x/xerrors"

	"cloud.google.com/go/datastore"
)

const KindHTMLName = "HTML"

type HTML struct {
	Content       []byte `datastore:",noindex"`
	Children      int
	TargetVersion string `datastore:"-"`
	Meta
}

func (h *HTML) Load(props []datastore.Property) error {
	return datastore.LoadStruct(h, props)
}

func (h *HTML) Save() ([]datastore.Property, error) {
	h.update(h.TargetVersion)
	return datastore.SaveStruct(h)
}

func createHTMLKey(id string) *datastore.Key {
	return datastore.NameKey(KindHTMLName, id, nil)
}

func GetHTML(ctx context.Context, id string) (*HTML, error) {

	var err error
	key := createHTMLKey(id)
	html := HTML{}

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.Get(ctx, key, &html)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, err
		} else {
			return nil, nil
		}
	}
	return &html, nil
}

func PutHTML(r *http.Request, id string) error {

	var err error

	page, err := SelectPage(r, id, -1)
	if err != nil {
		return xerrors.Errorf("SelectPage() error: %w", err)
	}

	dtos, _, err := NewDtos(r, page, "", false)
	if err != nil {
		return xerrors.Errorf("NewDtos() error: %w", err)
	}

	tmpl, err := createTemplate(r, page, false)
	if err != nil {
		return xerrors.Errorf("createTemplate() error: %w", err)
	}

	ctx := r.Context()

	htmls := make([]*HTML, len(dtos))
	keys := make([]*datastore.Key, len(dtos))

	for idx, dto := range dtos {
		var buf []byte
		w := bytes.NewBuffer(buf)

		err = tmpl.Execute(w, dto)
		if err != nil {
			return xerrors.Errorf("PageTemplate Execute() error: %w", err)
		}

		realId := id
		if idx > 0 {
			realId = fmt.Sprintf("%s?page=%d", id, idx+1)
		}

		html, err := GetHTML(ctx, realId)
		if err != nil {
			return xerrors.Errorf("GetHTML() error: %w", err)
		}
		if html == nil {
			key := createHTMLKey(realId)
			html = &HTML{}
			html.LoadKey(key)
		}
		html.Content = w.Bytes()
		html.Children = len(dtos)
		htmls[idx] = html
		keys[idx] = html.GetKey()
	}

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		_, err = tx.PutMulti(keys, htmls)
		if err != nil {
			return xerrors.Errorf("HTML PutMulti() error: %w", err)
		}
		page.Publish = time.Now()
		_, err = tx.Put(page.GetKey(), page)
		if err != nil {
			return xerrors.Errorf("Page(Publish) Put() error: %w", err)
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

func RemoveHTML(r *http.Request, id string) error {

	ctx := r.Context()
	page, err := SelectPage(r, id, -1)
	if err != nil {
		return err
	}
	if page == nil {
		return fmt.Errorf("page not found[%s]", id)
	}

	//TODO ページ数個削除
	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		key := createHTMLKey(id)
		err = tx.Delete(key)
		if err != nil {
			return err
		}

		page.Publish = time.Time{}
		_, err = tx.Put(page.GetKey(), page)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}

func PutHTMLs(r *http.Request, pages []Page) error {

	var err error
	//HTMLとを作成
	//HTMLキーを作成
	htmlData := make([][]byte, 0)
	keys := make([]*datastore.Key, 0)
	for _, elm := range pages {

		if elm.Deleted {
			continue
		}

		var buf []byte
		w := bytes.NewBuffer(buf)

		//TODO Referenceなので、NextはすでにDtoに埋め込まれている

		dtos, _, err := NewDtos(r, &elm, NoLimitCursor, false)
		if err != nil {
			return xerrors.Errorf("Reference NewDto() error: %w", err)
		}

		tmpl, err := createTemplate(r, &elm, false)
		if err != nil {
			return xerrors.Errorf("Reference createTemplate() error: %w", err)
		}

		err = tmpl.Execute(w, dtos[0])
		if err != nil {
			return xerrors.Errorf("Reference createTemplate() error: %w", err)
		}

		htmlData = append(htmlData, w.Bytes())
		keys = append(keys, createHTMLKey(elm.Key.Name))
	}

	htmls := make([]HTML, len(keys))

	ctx := r.Context()
	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.GetMulti(ctx, keys, htmls)
	if err != nil {
		//TODO
		return xerrors.Errorf("GetMulti() error: %w", err)
	}

	for idx, _ := range htmls {
		htmls[idx].LoadKey(keys[idx])
		htmls[idx].Content = htmlData[idx]
	}

	_, err = cli.PutMulti(ctx, keys, htmls)
	if err != nil {
		return xerrors.Errorf("htmls PutMulti() error: %w", err)
	}

	return nil
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
	Prev     string
	Next     string
}

//
func WriteManageHTML(w http.ResponseWriter, r *http.Request, id string, p int) error {

	var err error
	page, err := SelectPage(r, id, -1)
	if err != nil {
		return err
	}
	if page == nil {
		return fmt.Errorf("Page not found[%s]", id)
	}

	//テンプレートの作成
	tmpl, err := createTemplate(r, page, true)
	if err != nil {
		return err
	}

	//TODO カーソルが空
	//DTOの作成
	dtos, _, err := NewDtos(r, page, "", true)
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

//TODO 出力時と表示時のテスト
func NewDtos(r *http.Request, page *Page, cur string, view bool) ([]*HTMLDto, string, error) {

	id := page.Key.Name
	site, err := SelectSite(r, -1)
	if err != nil {
		return nil, "", xerrors.Errorf("SelectSite() error: %w", err)
	}

	pData, err := SelectPageData(r, id)
	if err != nil {
		return nil, "", xerrors.Errorf("SelectPageData() error: %w", err)
	}

	content := string(pData.Content)
	dto := HTMLDto{
		Site:     site,
		Page:     page,
		PageData: pData,
		Content:  content,
	}

	dir := "/manage/page/view/"
	top := "/manage/page/view/"

	//表示でない場合
	if !view {
		if page.Deleted {
			return nil, "", fmt.Errorf("Page is private or deleted[%s]", id)
		}
		dir = "/page/"
		top = "/"
	}

	childNum := 0
	//本番時には複数件取得して展開
	if view && page.Paging > 0 {
		childNum = page.Paging
	}

	children, next, err := SelectChildPages(r, id, cur, childNum, view)
	if err != nil {
		return nil, "", xerrors.Errorf("SelectChildPages() error: %w", err)
	}

	leng := len(children)
	last := 1
	dtoNum := 1

	if page.Paging > 0 {
		last = leng/page.Paging + 1
		mod := leng % page.Paging

		if mod == 0 {
			last -= 1
		}

		if !view {
			dtoNum = last
		}
	}

	dtos := make([]*HTMLDto, dtoNum)

	for idx := 0; idx < dtoNum; idx++ {

		cp := dto
		prevId := ""
		nextId := ""
		pNum := idx + 1

		/*
			if view {
				pNum = pageNum
			}
		*/

		if last != pNum || view {
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
		if !view && page.Paging > 0 {
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

	return dtos, next, nil
}

func createTemplate(r *http.Request, page *Page, mng bool) (*template.Template, error) {
	//テンプレートを取得
	siteTmp, err := SelectTemplateData(r, page.SiteTemplate)
	if err != nil {
		return nil, err
	}
	pageTmp, err := SelectTemplateData(r, page.PageTemplate)
	if err != nil {
		return nil, err
	}
	siteTmpData := string(siteTmp.Content)
	pageTmpData := string(pageTmp.Content)
	siteTmpData = "{{define \"" + api.SiteTemplateName + "\"}}" + "\n" + siteTmpData + "\n" + "{{end}}"
	pageTmpData = "{{define \"" + api.PageTemplateName + "\"}}" + "\n" + pageTmpData + "\n" + "{{end}}"

	pub := Public{
		request: r,
		manage:  mng,
	}
	//適用する
	tmpl, err := template.New(api.SiteTemplateName).Funcs(pub.funcMap()).Parse(siteTmpData)
	if err != nil {
		return nil, err
	}
	tmpl, err = tmpl.Parse(pageTmpData)
	if err != nil {
		return nil, err
	}
	return tmpl, nil
}

//Public template object
type Public struct {
	request *http.Request
	manage  bool
}

func (p Public) list(id string, num int) []Page {
	//TODO 1ページ目固定
	pages, _, err := SelectChildPages(p.request, id, "", num, p.manage)
	if err != nil {
		return make([]Page, 0)
	}
	return pages
}

var privateMark = `data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAEAAAABACAYAAACqaXHeAAAENklEQVR4XuVb240cMQzLtZIKkv6LSCpIKwkmgA9arSiSsuceuPzkHrYs0RTFGdy+fPvi/16m9f/58fPv99+/Xq7/pzGcfeusfGb8/vraiXmttTe8VcG5EBVsFwQZgPcq3L3RtV4FggLACo8HrbWZlijG2lvtmxZ+7avaBMWDADiF7yR7Ym8E8IpX5Y4YUQLAbuxE0nfGQAJdgSADoPbUnYU5sdVLfABA3eQk8t5rWTtQAD7bzWfAtwD47MUvMDoQXhmQF02Kn8TIghUd36n2eRMA0ChSgOxGbrV/geYAhEA4wgB08xmULuHODFVmK8aagnztKwFQAiLD0RXJ4sabZaBW5yjxM3D/AXB7F9GJucdlUxUmVA9DHeiKdlR12gCwItktsL5lerDjVcYAdIeyOcsKrn7fxdw5bwwAK6IaZWyP2gaZUbcAkHvLofFOQgwEVrz7NggyIDsmFYCdfqzOZGJc/X4yatfZD2MwskAFgDFHSY4VHVlSGSDHFN2iAahItzBmbE602nEAYvExOHrVhWwtKx7Nf4epyO/YPmAFyu/dKuPCbk3VEOYN1IlzGwOQa1MSy8xBfZ5ZougLOj/uHTEgK7hCYZZMF4MxSQG6yvl1CjiClVW5OtzpTbU4dR3zFnEEXl9vMyA+hKhJRpp3e5DITluuEuoxAJ2A5f6sCo6TAjm6Ks4pxq3zjwNwJVhNiGo0onbKa1dMZIocQcwsGAHQjSUEQEfbSoOYr2DgVUxpx6DzYDEFoDqDxXKnQ8UWNLUepoDyRqULlBF3BHHCmmvP0hZ1ikEGrGB5RDAaMdYoDk5hQDdWHRAkANgMz73p3jQCmcVFeXXTSLnAp7fCDgBM8VHLxD7tblBVd/USItPHYxAlXIlVtxY9MbomZ1J8ZKE1BhFNOy2It8ho7hbviHLWucXCIwDkVqiMTFwT20wRyqqXq6kzad8RADsF5iRdEVPAQGsqc2UDoKgxKlJ5I9QZmZ3ic7vYIrgELSYRf6Yqdt5fFcWofAKIFcNiQNV3OzfmmBg00x2wqks6BkAsRknqhPhVrOzYccQJIgorRXf0Z7b6JPsiE47/fQADYtf753nOzqtAjw9+2wDklx8sodyHCiDRue2MTakFHFGrpgDaX83gyp11PeyMVySa8eev7wOQTWTjpkK0MyGqD0AxKrY54xcywHmgcOc4U2q1BXLxuS3U1ruVAU77ZGfG2LZin7qsByd4ogXYDSgFdpqCNESNm9c9ATABwdEAJdE71yAAv8wfSyOj9fR5gWmP3Xl709hKLfADE67BmSZ5175MeTQu7c8MnRC7u4pmxqrK3QYgjqOPCIZrlduPzXXBVAOTjUelxtU0OXU2u6Ttzw26dEbvE904ynpW/BWDAlA5ts6Wsj5UEs+O0mWbUvjKQwYgJ37K4yNAYhEMgPh8rwK8DYB70Edd/w+BjJCMxE5c7wAAAABJRU5ErkJggg==`

func (p Public) mark() template.HTML {
	src := ""
	if p.manage {
		src = `<img src="` + privateMark + `" style="position: fixed; display: block; right: 0; bottom: 0; margin-right: 40px; margin-bottom: 40px; z-index: 900;" />`
	}
	return template.HTML(src)
}

func (p Public) funcMap() template.FuncMap {
	return template.FuncMap{
		"html":            api.ConvertHTML,
		"eraseBR":         api.EraseBR,
		"plane":           api.ConvertString,
		"convertDate":     api.ConvertDate,
		"list":            p.list,
		"mark":            p.mark,
		"templateContent": p.ConvertTemplate,
	}
}

//Contentの変換時にテンプレートを実現する
func (p Public) ConvertTemplate(data string) template.HTML {

	tmpl, err := template.New("").Funcs(p.funcMap()).Parse(data)
	if err != nil {
		return template.HTML(fmt.Sprintf("Template Parse Error[%s]", err))
	}

	dto := struct {
		Dir string
		Top string
	}{"/page/", "/"}
	if p.manage {
		dto.Dir = "/manage/page/view/"
		dto.Top = "/manage/page/view/"
	}

	buf := bytes.NewBuffer(make([]byte, 0, len(data)+30))
	err = tmpl.Execute(buf, dto)
	if err != nil {
		return template.HTML(fmt.Sprintf("Template Execute Error[%s]", err))
	}

	return template.HTML(buf.String())
}
