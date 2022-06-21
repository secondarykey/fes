package logic

import (
	"app/api"
	"app/datastore"
	"context"
	"io"

	"bytes"
	"fmt"
	"html/template"
	"net/http"

	"golang.org/x/xerrors"
)

func WriteManageHTML(w io.Writer, r *http.Request, id string, page int, ve *ErrorDto) error {

	gen := newGenerator()
	var err error
	ctx := r.Context()

	htmls, _, err := gen.createHTMLs(ctx, true, ve, id)
	if err != nil {
		return xerrors.Errorf("createHTMLs() error: %w", err)
	}

	if len(htmls) >= page {
		page -= 1
	} else {
		page = len(htmls) - 1
	}

	fmt.Println("Len", len(htmls))
	fmt.Println("page", page)

	_, err = w.Write(htmls[page].Content)
	if err != nil {
		return xerrors.Errorf("writer Write() error: %w", err)
	}

	return nil
}

func PutHTMLs(ctx context.Context, ids ...string) error {
	gen := newGenerator()
	defer gen.dao.Close()

	htmls, page, err := gen.createHTMLs(ctx, false, nil, ids...)
	if err != nil {
		return xerrors.Errorf("datastore.PutHTML(): %w", err)
	}

	err = gen.dao.PutHTML(ctx, htmls, page)
	if err != nil {
		return xerrors.Errorf("PutHTML() error: %w", err)
	}

	return nil
}

type HTMLDto struct {
	Site     *datastore.Site
	Page     *datastore.Page
	PageData *datastore.PageData
	Content  string
	Children []datastore.Page
	Top      string
	Dir      string
	Prev     string
	Next     string

	Error *ErrorDto
}

type ErrorDto struct {
	No      int
	Message string
	Detail  string
}

func (gen *Generator) createHTMLDto(ctx context.Context, page *datastore.Page, pData *datastore.PageData, view bool, ve *ErrorDto) ([]*HTMLDto, error) {

	id := page.Key.Name

	site, err := gen.dao.SelectSite(ctx, -1)
	if err != nil {
		return nil, xerrors.Errorf("SelectSite() error: %w", err)
	}

	dir := "/manage/page/view/"
	top := "/manage/page/view/"
	//表示でない場合
	if !view {
		if page.Deleted {
			return nil, fmt.Errorf("Page is private or deleted[%s]", id)
		}
		dir = "/page/"
		top = "/"
	}

	content := string(pData.Content)
	dto := HTMLDto{
		Site:     site,
		Page:     page,
		PageData: pData,
		Content:  content,
		Error:    ve,
		Top:      top,
		Dir:      dir,
	}

	children, _, err := gen.dao.SelectChildrenPage(ctx, id, "", page.Paging, view)
	if err != nil {
		return nil, xerrors.Errorf("SelectChildPages() error: %w", err)
	}
	dto.Children = children

	dtos := make([]*HTMLDto, 0)
	dtos = append(dtos, &dto)

	return dtos, nil
}

func ClearTemplateCache() {
	startTemplateCache()
}

var (
	cacheTemplateData = make(map[string]string)
)

func startTemplateCache() {
	cacheTemplateData = make(map[string]string)
}

func deleteTemplateCache() {
	cacheTemplateData = nil
}

func (gen *Generator) createTemplate(ctx context.Context, page *datastore.Page, mng bool, dto interface{}) (*template.Template, error) {

	var ok bool
	var siteTmpData string
	var pageTmpData string

	if siteTmpData, ok = cacheTemplateData[page.SiteTemplate]; !ok {
		siteTmp, err := gen.dao.SelectTemplateData(ctx, page.SiteTemplate)
		if err != nil {
			return nil, xerrors.Errorf("datastore.SelectTemplateData(Site) error: %w", err)
		}
		siteTmpData = string(siteTmp.Content)
		cacheTemplateData[page.SiteTemplate] = siteTmpData
	}

	if pageTmpData, ok = cacheTemplateData[page.PageTemplate]; !ok {
		pageTmp, err := gen.dao.SelectTemplateData(ctx, page.PageTemplate)
		if err != nil {
			return nil, xerrors.Errorf("datastore.SelectTemplateData(Page) error: %w", err)
		}
		pageTmpData = string(pageTmp.Content)
		cacheTemplateData[page.PageTemplate] = pageTmpData
	}

	siteTmpData = fmt.Sprintf(`{{ define "%s" }}%s{{end}}`, api.SiteTemplateName, siteTmpData)
	pageTmpData = fmt.Sprintf(`{{ define "%s" }}%s{{end}}`, api.PageTemplateName, pageTmpData)

	helper := api.Helper{
		Ctx:         ctx,
		ID:          page.GetKey().Name,
		Manage:      mng,
		Dao:         gen.dao,
		TemplateDto: dto,
	}

	//適用する
	tmpl, err := template.New(api.SiteTemplateName).Funcs(helper.FuncMap()).Parse(siteTmpData)
	if err != nil {
		return nil, xerrors.Errorf("Template New() error: %w", err)
	}
	tmpl, err = tmpl.Parse(pageTmpData)
	if err != nil {
		return nil, xerrors.Errorf("Template Parse() error: %w", err)
	}
	return tmpl, nil
}

type Generator struct {
	dao *datastore.Dao
}

func newGenerator() *Generator {
	var gen Generator
	gen.dao = datastore.NewDao()
	return &gen
}

func (gen *Generator) createHTMLs(ctx context.Context, mng bool, ve *ErrorDto, ids ...string) ([]*datastore.HTML, *datastore.Page, error) {

	pages, err := gen.dao.SelectPages(ctx, ids...)
	if err != nil {
		return nil, nil, xerrors.Errorf("datasore.SelectPages() error: %w", err)
	}

	page := &pages[0]
	if len(pages) > 1 {
		page = nil
	}

	data, err := gen.dao.GetPageData(ctx, ids...)
	if err != nil {
		return nil, nil, xerrors.Errorf("datasore.GetPageData() error: %w", err)
	}

	//HTMLとを作成
	htmlData := make([][]byte, 0)
	keys := make([]string, 0)

	for idx, elm := range pages {

		if elm.Deleted {
			continue
		}
		pData := data[idx]

		dtos, err := gen.createHTMLDto(ctx, &elm, &pData, mng, nil)
		if err != nil {
			return nil, nil, xerrors.Errorf("createHTMLDto() error: %w", err)
		}

		tmpl, err := gen.createTemplate(ctx, &elm, mng, dtos[0])
		if err != nil {
			return nil, nil, xerrors.Errorf("createTemplate() error: %w", err)
		}

		var buf []byte
		w := bytes.NewBuffer(buf)
		err = tmpl.Execute(w, dtos[0])
		if err != nil {
			return nil, nil, xerrors.Errorf("Reference createTemplate() error: %w", err)
		}
		htmlData = append(htmlData, w.Bytes())
		keys = append(keys, elm.Key.Name)
	}

	htmls := make([]*datastore.HTML, len(keys))
	for idx, _ := range htmls {
		htmls[idx] = &datastore.HTML{}
		htmls[idx].LoadKey(datastore.CreateHTMLKey(keys[idx]))
		htmls[idx].Content = htmlData[idx]
	}

	return htmls, page, nil
}
