package manage

import (
	"app/datastore"
	"app/handler/manage/form"
	"app/logic"

	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func viewDraftHandler(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query()
	cursor := q.Get("cursor")

	dao := datastore.NewDao()
	defer dao.Close()

	data, next, err := dao.SelectDrafts(r.Context(), cursor)
	if err != nil {
		errorPage(w, "Error Select Draft", err, 500)
		return
	}

	if data == nil {
		data = make([]datastore.Draft, 0)
	}

	current, err := GetDraftId(r)
	if err == nil && current != "" {
		for idx := range data {
			if current == data[idx].Key.Name {
				data[idx].Current = true
			}
		}
	}

	dto := struct {
		Drafts []datastore.Draft
		Now    string
		Next   string
	}{data, cursor, next}

	viewManage(w, "draft/view.tmpl", dto)
}

func addDraftHandler(w http.ResponseWriter, r *http.Request) {

	draft := &datastore.Draft{}
	draft.LoadKey(datastore.CreateDraftKey())

	draftPages := make([]*datastore.DraftPage, 0)

	//新規作成用のテンプレート
	dto := struct {
		Draft      *datastore.Draft
		DraftPages []*datastore.DraftPage
	}{draft, draftPages}

	viewManage(w, "draft/edit.tmpl", dto)
}

func editDraftHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Add("X-XSS-Protection", "1")

	dao := datastore.NewDao()
	defer dao.Close()

	ctx := r.Context()
	vars := mux.Vars(r)
	id := vars["key"]

	//POST
	if POST(r) {

		set := datastore.DraftSet{}

		draft, err := dao.SelectDraft(ctx, id)
		if err != nil {
			errorPage(w, "Error SelectDraft", err, 500)
			return
		}

		if draft == nil {
			draft = &datastore.Draft{}
		}

		set.Draft = draft

		err = form.SetDraftSet(r, &set)
		if err != nil {
			errorPage(w, "Error SetDraft()", err, 500)
			return
		}

		//更新
		err = dao.PutDraftSet(ctx, &set)
		if err != nil {
			errorPage(w, "Error Put Draft", err, 500)
			return
		}
	}

	set, err := dao.SelectDraftSet(ctx, id)
	if err != nil {
		errorPage(w, "Error SelectDraft", err, 500)
		return
	}
	if set == nil {
		errorPage(w, "NotFound Draft", fmt.Errorf(id), 404)
		return
	}

	dto := struct {
		Draft      *datastore.Draft
		DraftPages []*datastore.DraftPage
	}{set.Draft, set.Pages}

	viewManage(w, "draft/edit.tmpl", dto)
}

func deleteDraftHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	err := dao.RemoveDraft(ctx, id)
	if err != nil {
		errorPage(w, "Remove Draft Error", err, 500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/draft/", 302)
}

func publishDraftHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()

	dao := datastore.NewDao()
	defer dao.Close()
	pages, err := dao.SelectDraftPages(ctx, id)
	if err != nil {
		errorPage(w, "GetDraftPageKeys() Error", err, 500)
		return
	}

	ids := make([]string, len(pages))
	for idx, p := range pages {
		ids[idx] = p.PageID
	}

	err = logic.PutHTMLs(ctx, ids...)
	if err != nil {
		errorPage(w, "logic.PutHTMLs() Error", err, 500)
		return
	}

	err = dao.RemoveDraft(ctx, id)
	if err != nil {
		errorPage(w, "Remove Draft Error", err, 500)
		return
	}
	//リダイレクト
	http.Redirect(w, r, "/manage/draft/", 302)
}

func currentDraftHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	//クッキーに作業用の下書きを設定

	err := SetDraftId(w, r, id)
	if err != nil {
		errorPage(w, "Set DraftId() Error", err, 500)
		return
	}

	http.Redirect(w, r, "/manage/draft/", 302)
}

func addDraftPageHandler(w http.ResponseWriter, r *http.Request) {

	current, err := GetDraftId(r)
	if err != nil || current == "" {
		errorPage(w, "GetDraftId() Error(Draft select)", err, 500)
		return
	}

	ctx := r.Context()
	dao := datastore.NewDao()
	defer dao.Close()

	draft, err := dao.SelectDraft(ctx, current)
	if err != nil {
		errorPage(w, "SelectDraft() Error", err, 500)
		return
	}

	if draft == nil {
		errorPage(w, "SelectDraft(DraftSelect) Error", err, 404)
		return
	}

	vars := mux.Vars(r)
	id := vars["key"]

	err = dao.AddDraftPage(ctx, current, id)
	if err != nil {
		errorPage(w, "AddDraftPage() Error", err, 500)
		return
	}

	http.Redirect(w, r, "/manage/page/"+id, 302)
}

func deleteDraftPageHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)
	id := vars["key"]

	dao := datastore.NewDao()
	defer dao.Close()

	ctx := r.Context()

	draftId, err := dao.RemoveDraftPage(ctx, id)
	if err != nil {
		errorPage(w, "DeleteDraftPage() Error", err, 500)
		return
	}

	http.Redirect(w, r, "/manage/draft/edit/"+draftId, 302)
}
