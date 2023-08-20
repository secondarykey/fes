package logic

import (
	"app/api"
	"app/datastore"
	"context"
	"io"
	"time"

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

	htmls, err := gen.createHTMLs(ctx, true, ve, id)
	if err != nil {
		return xerrors.Errorf("createHTMLs() error: %w", err)
	}

	if len(htmls) >= page {
		page -= 1
	} else {
		page = len(htmls) - 1
	}

	_, err = w.Write(htmls[page].Content)
	if err != nil {
		return xerrors.Errorf("writer Write() error: %w", err)
	}

	return nil
}

type PageInfo struct {
	ID      string
	Publish bool
}

func NewPageInfo(id string) *PageInfo {
	var info PageInfo
	info.ID = id
	return &info
}

func NewPageInfos(ids ...string) []*PageInfo {
	infos := make([]*PageInfo, len(ids))
	for idx, id := range ids {
		infos[idx] = NewPageInfo(id)
	}
	return infos
}

func PutHTMLs(ctx context.Context, infos ...*PageInfo) error {

	gen := newGenerator()
	defer gen.dao.Close()

	ids := make([]string, len(infos))
	for idx, info := range infos {
		ids[idx] = info.ID
	}

	//一旦公開日用にページに更新をかける
	ps, err := gen.dao.SelectPages(ctx, ids...)
	if err != nil {
		return xerrors.Errorf("datasore.SelectPages() error: %w", err)
	}

	now := time.Now()
	up := make([]*datastore.Page, 0)
	for idx, page := range ps {
		if page.Publish.IsZero() || infos[idx].Publish {
			page.Publish = now
		}
		up = append(up, &page)
	}

	//TODO 下書きの画像があるので結局更新
	if len(up) > 0 {
		//ページを更新
		err = gen.dao.PutPages(ctx, up)
		if err != nil {
			return xerrors.Errorf("datastore.PutPages(): %w", err)
		}
	}

	htmls, err := gen.createHTMLs(ctx, false, nil, ids...)
	if err != nil {
		return xerrors.Errorf("datastore.PutHTML(): %w", err)
	}

	err = gen.dao.PutHTML(ctx, htmls)
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

func (gen *Generator) createHTMLs(ctx context.Context, mng bool, ve *ErrorDto, ids ...string) ([]*datastore.HTML, error) {

	pages, err := gen.dao.SelectPages(ctx, ids...)
	if err != nil {
		return nil, xerrors.Errorf("datasore.SelectPages() error: %w", err)
	}

	data, err := gen.dao.GetPageData(ctx, ids...)
	if err != nil {
		return nil, xerrors.Errorf("datasore.GetPageData() error: %w", err)
	}

	//1 件の場合に同時に公開日付を設定する為に取得
	//一旦検索をかけるが、親子関係を見て、子から処理を行う

	//HTMLを作成
	htmlData := make([][]byte, 0)
	keys := make([]string, 0)

	//Page数回繰り返す
	for idx, elm := range pages {

		if elm.Deleted {
			continue
		}

		pData := data[idx]

		//TODO dtos -> ページ数の仕様を確認
		// 子の検索の時に、公開日を最新のものにする仕組みを考える
		dtos, err := gen.createHTMLDto(ctx, &elm, &pData, mng, nil)
		if err != nil {
			return nil, xerrors.Errorf("createHTMLDto() error: %w", err)
		}

		//テンプレートを作成
		tmpl, err := gen.createTemplate(ctx, &elm, mng, dtos[0])
		if err != nil {
			return nil, xerrors.Errorf("createTemplate() error: %w", err)
		}

		var buf []byte
		w := bytes.NewBuffer(buf)
		//テンプレートを作成
		err = tmpl.Execute(w, dtos[0])
		if err != nil {
			return nil, xerrors.Errorf("Reference createTemplate() error: %w", err)
		}

		htmlData = append(htmlData, w.Bytes())
		keys = append(keys, elm.Key.Name)
	}

	htmls := make([]*datastore.HTML, len(keys))
	for idx, _ := range htmls {
		htmls[idx] = &datastore.HTML{}
		htmls[idx].LoadKey(datastore.GetHTMLKey(keys[idx]))
		htmls[idx].Content = htmlData[idx]
	}

	return htmls, nil
}
