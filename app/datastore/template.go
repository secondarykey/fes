package datastore

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/api/iterator"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
)

const KindTemplateName = "Template"

type Template struct {
	Name string
	Type int

	TargetVersion string `datastore:"-"`
	Meta
}

func (t *Template) Load(props []datastore.Property) error {
	return datastore.LoadStruct(t, props)
}

func (t *Template) Save() ([]datastore.Property, error) {
	t.update(t.TargetVersion)
	return datastore.SaveStruct(t)
}

func PutTemplate(r *http.Request) error {

	var err error

	vars := mux.Vars(r)
	id := vars["key"]

	ctx := r.Context()
	tmpKey := datastore.NameKey(KindTemplateName, id, nil)
	tmpDataKey := datastore.NameKey(KindTemplateDataName, id, nil)

	template := Template{}
	templateData := TemplateData{}

	ver := r.FormValue("version")
	version, err := strconv.Atoi(ver)
	if err != nil {
		return err
	}

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	if version > 0 {
		template.TargetVersion = fmt.Sprintf("%d", version)
	}

	//TODO Version
	err = cli.Get(ctx, tmpKey, &template)

	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return err
		}
	}

	template.LoadKey(tmpKey)
	templateData.LoadKey(tmpDataKey)

	template.Name = r.FormValue("name")
	template.Type, err = strconv.Atoi(r.FormValue("templateType"))
	if err != nil {
		return err
	}

	//TODO ByteStringからの変換
	templateData.Content = []byte(r.FormValue("template"))

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		_, err = tx.Put(template.GetKey(), &template)
		if err != nil {
			return err
		}

		_, err = tx.Put(templateData.GetKey(), &templateData)
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

func CreateTemplateKey() *datastore.Key {
	id := uuid.NewV4()
	return datastore.NameKey(KindTemplateName, id.String(), nil)
}

func SelectTemplate(r *http.Request, id string) (*Template, error) {
	temp := Template{}
	ctx := r.Context()
	//Method
	key := datastore.NameKey(KindTemplateName, id, nil)
	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.Get(ctx, key, &temp)
	if err != nil {
		if errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, err
		} else {
			return nil, nil
		}
	}
	return &temp, nil
}

func getTemplateCursor(p int) string {
	return "template_" + strconv.Itoa(p) + "_cursor"
}

func SelectTemplates(r *http.Request, p int) ([]Template, error) {

	var rtn []Template

	ctx := r.Context()
	cursor := ""

	q := datastore.NewQuery(KindTemplateName).Order("- UpdatedAt")
	if p > 0 {
		//TODO 新しい残し方
		log.Println("cursor not implemented", cursor)
	}

	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	t := cli.Run(ctx, q)
	for {
		var tmp Template
		key, err := t.Next(&tmp)
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, err
		}
		tmp.LoadKey(key)
		rtn = append(rtn, tmp)
	}

	if p > 0 {

		cur, err := t.Cursor()
		if err != nil {
			return nil, err
		}

		log.Println("cursor notimplemented", cur)

	}

	return rtn, nil
}

const KindTemplateDataName = "TemplateData"

type TemplateData struct {
	Key     *datastore.Key `datastore:"__key__"`
	Content []byte
}

func (d *TemplateData) GetKey() *datastore.Key {
	return d.Key
}

func (d *TemplateData) LoadKey(k *datastore.Key) error {
	d.Key = k
	return nil
}

func SelectTemplateData(r *http.Request, id string) (*TemplateData, error) {
	temp := TemplateData{}
	ctx := r.Context()
	//Method
	key := datastore.NameKey(KindTemplateDataName, id, nil)
	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.Get(ctx, key, &temp)
	if err != nil {
		return nil, err
	}
	return &temp, nil
}

func RemoveTemplate(r *http.Request, id string) error {

	var err error
	ctx := r.Context()

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		key := datastore.NameKey(KindTemplateName, id, nil)
		err = tx.Delete(key)
		if err != nil {
			return err
		}

		dataKey := datastore.NameKey(KindTemplateDataName, id, nil)
		err = tx.Delete(dataKey)
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
