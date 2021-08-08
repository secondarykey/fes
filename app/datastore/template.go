package datastore

import (
	"context"
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

const (
	TemplateTypeAll  = 0
	TemplateTypeSite = 1
	TemplateTypePage = 2
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
	tmpKey := SetTemplateKey(id)
	tmpDataKey := createTemplateDataKey(id)

	template := Template{}
	templateData := TemplateData{}

	ver := r.FormValue("version")
	version, err := strconv.Atoi(ver)
	if err != nil {
		return xerrors.Errorf("version strconv.Atoi() error: %w", err)
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
			return xerrors.Errorf("Template Get() error: %w", err)
		}
	}

	template.LoadKey(tmpKey)
	templateData.LoadKey(tmpDataKey)

	template.Name = r.FormValue("name")
	template.Type, err = strconv.Atoi(r.FormValue("templateType"))
	if err != nil {
		return xerrors.Errorf("TemplateType Atoi() error: %w", err)
	}

	templateData.Content = []byte(r.FormValue("template"))

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		_, err = tx.Put(template.GetKey(), &template)
		if err != nil {
			return xerrors.Errorf("Template Put() error: %w", err)
		}

		_, err = tx.Put(templateData.GetKey(), &templateData)
		if err != nil {
			return xerrors.Errorf("TemplateData Put() error: %w", err)
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
	return datastore.NameKey(KindTemplateName, id.String(), createSiteKey())
}

func SetTemplateKey(id string) *datastore.Key {
	return datastore.NameKey(KindTemplateName, id, createSiteKey())
}

func SelectTemplate(ctx context.Context, id string) (*Template, error) {
	temp := Template{}

	//Method
	key := SetTemplateKey(id)
	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.Get(ctx, key, &temp)
	if err != nil {
		if !errors.Is(err, datastore.ErrNoSuchEntity) {
			return nil, xerrors.Errorf("Template Get() error: %w", err)
		} else {
			return nil, nil
		}
	}
	return &temp, nil
}

func SelectTemplates(ctx context.Context, ty string, cur string) ([]Template, string, error) {

	var rtn []Template

	q := datastore.NewQuery(KindTemplateName).Order("- UpdatedAt")

	if ty != "all" {
		v, err := strconv.Atoi(ty)
		if err == nil {
			q = q.Filter("Type=", v)
		} else {
			log.Println("strconv parse error", ty)
		}
	}

	if cur != NoLimitCursor {
		q = q.Limit(10)
		if cur != "" {
			cursor, err := datastore.DecodeCursor(cur)
			if err != nil {
				return nil, "", xerrors.Errorf("datastore.DecodeCursor() error: %w", err)
			}
			q = q.Start(cursor)
		}
	}

	cli, err := createClient(ctx)
	if err != nil {
		return nil, "", xerrors.Errorf("createClient() error: %w", err)
	}

	t := cli.Run(ctx, q)
	for {
		var tmp Template
		key, err := t.Next(&tmp)
		if errors.Is(err, iterator.Done) {
			break
		}

		if err != nil {
			return nil, "", xerrors.Errorf("Template Next() error: %w", err)
		}
		tmp.LoadKey(key)
		rtn = append(rtn, tmp)
	}

	cursor, err := t.Cursor()
	if err != nil {
		return nil, "", xerrors.Errorf("Template Cursor() error: %w", err)
	}

	return rtn, cursor.String(), nil
}

const KindTemplateDataName = "TemplateData"

type TemplateData struct {
	Key     *datastore.Key `datastore:"__key__"`
	Content []byte         `datastore:",noindex"`
}

func (d *TemplateData) GetKey() *datastore.Key {
	return d.Key
}

func (d *TemplateData) LoadKey(k *datastore.Key) error {
	d.Key = k
	return nil
}

func createTemplateDataKey(id string) *datastore.Key {
	return datastore.NameKey(KindTemplateDataName, id, createSiteKey())
}

func SelectTemplateData(ctx context.Context, id string) (*TemplateData, error) {
	temp := TemplateData{}

	//Method
	key := createTemplateDataKey(id)
	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	err = cli.Get(ctx, key, &temp)
	if err != nil {
		return nil, xerrors.Errorf("TemplateData Get() error: %w", err)
	}
	return &temp, nil
}

func RemoveTemplate(ctx context.Context, id string) error {

	var err error

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		key := SetTemplateKey(id)
		err = tx.Delete(key)
		if err != nil {
			return xerrors.Errorf("Template Delete() error: %w", err)
		}

		dataKey := createTemplateDataKey(id)
		err = tx.Delete(dataKey)
		if err != nil {
			return xerrors.Errorf("TemplateData Delete() error: %w", err)
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}

	return nil
}
