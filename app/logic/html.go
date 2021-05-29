package logic

import (
	"app/api"
	"app/datastore"

	"bytes"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"golang.org/x/xerrors"
)

func createTemplate(r *http.Request, page *datastore.Page, mng bool) (*template.Template, error) {

	//テンプレートを取得
	siteTmp, err := datastore.SelectTemplateData(r, page.SiteTemplate)
	if err != nil {
		return nil, xerrors.Errorf("datastore.SelectTemplateData(Site) error: %w", err)
	}
	pageTmp, err := datastore.SelectTemplateData(r, page.PageTemplate)
	if err != nil {
		return nil, xerrors.Errorf("datastore.SelectTemplateData(Page) error: %w", err)
	}
	siteTmpData := string(siteTmp.Content)
	pageTmpData := string(pageTmp.Content)
	siteTmpData = "{{define \"" + api.SiteTemplateName + "\"}}" + "\n" + siteTmpData + "\n" + "{{end}}"
	pageTmpData = "{{define \"" + api.PageTemplateName + "\"}}" + "\n" + pageTmpData + "\n" + "{{end}}"

	helper := api.Helper{
		Request: r,
		Manage:  mng,
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

func WriteManageHTML(w http.ResponseWriter, r *http.Request, id string, p int) error {

	var err error
	page, err := datastore.SelectPage(r, id, -1)
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
		return xerrors.Errorf("NewDtos() error: %w", err)
	}

	//書き込み
	err = tmpl.Execute(w, dtos[0])
	if err != nil {
		return xerrors.Errorf("Template Execute() error: %w", err)
	}
	return err
}

func PutHTML(r *http.Request, id string) error {

	var err error

	page, err := datastore.SelectPage(r, id, -1)
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

	htmls := make([]*datastore.HTML, len(dtos))

	//pageを廃止する

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

		html, err := datastore.GetHTML(ctx, realId)
		if err != nil {
			return xerrors.Errorf("GetHTML() error: %w", err)
		}
		if html == nil {
			html = &datastore.HTML{}
			html.LoadKey(datastore.CreateHTMLKey(realId))
		}

		html.Content = w.Bytes()
		htmls[idx] = html
	}

	err = datastore.PutHTML(ctx, htmls, page)
	if err != nil {
		return xerrors.Errorf("datastore.PutHTML(): %w", err)
	}

	return nil
}

func PutHTMLs(r *http.Request, pages []datastore.Page) error {

	var err error

	//HTMLとを作成
	htmlData := make([][]byte, 0)
	keys := make([]string, 0)

	for _, elm := range pages {

		if elm.Deleted {
			continue
		}

		var buf []byte
		w := bytes.NewBuffer(buf)

		//TODO Referenceなので、NextはすでにDtoに埋め込まれている

		dtos, _, err := NewDtos(r, &elm, datastore.NoLimitCursor, false)
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
		keys = append(keys, elm.Key.Name)
	}

	htmls := make([]*datastore.HTML, len(keys))
	for idx, _ := range htmls {
		htmls[idx] = &datastore.HTML{}
		htmls[idx].LoadKey(datastore.CreateHTMLKey(keys[idx]))
		htmls[idx].Content = htmlData[idx]
	}

	//TODO 必要かな？(他の属性が変更される可能性)
	//err = datastore.GetMulti(ctx, htmls)
	//if err != nil {
	//return xerrors.Errorf("GetMulti() error: %w", err)
	//}

	ctx := r.Context()
	err = datastore.PutHTML(ctx, htmls, nil)
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
}

//TODO 出力時と表示時のテスト
func NewDtos(r *http.Request, page *datastore.Page, cur string, view bool) ([]*HTMLDto, string, error) {

	ctx := r.Context()
	id := page.Key.Name
	site, err := datastore.SelectSite(ctx, -1)
	if err != nil {
		return nil, "", xerrors.Errorf("SelectSite() error: %w", err)
	}

	pData, err := datastore.SelectPageData(r, id)
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

	children, next, err := datastore.SelectChildPages(r, id, cur, childNum, view)
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
