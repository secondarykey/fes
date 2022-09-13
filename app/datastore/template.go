package datastore

import (
	uuid "github.com/satori/go.uuid"

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

func CreateTemplateKey() *datastore.Key {
	id := uuid.NewV4()
	return datastore.NameKey(KindTemplateName, id.String(), getSiteKey())
}

func GetTemplateKey(id string) *datastore.Key {
	return datastore.NameKey(KindTemplateName, id, getSiteKey())
}

func (t *Template) Load(props []datastore.Property) error {
	return datastore.LoadStruct(t, props)
}

func (t *Template) Save() ([]datastore.Property, error) {
	err := t.update()
	if err != nil {
		return nil, xerrors.Errorf("Meta update() error: %w", err)
	}
	return datastore.SaveStruct(t)
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

func GetTemplateDataKey(id string) *datastore.Key {
	return datastore.NameKey(KindTemplateDataName, id, getSiteKey())
}
