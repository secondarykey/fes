package datastore

import (
	uuid "github.com/satori/go.uuid"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
)

const KindDraftName = "Draft"

type Draft struct {
	Name    string
	Current bool
	Meta
}

func (t *Draft) Load(props []datastore.Property) error {
	return datastore.LoadStruct(t, props)
}

func (t *Draft) Save() ([]datastore.Property, error) {
	err := t.update()
	if err != nil {
		return nil, xerrors.Errorf("Meta update() error: %w", err)
	}
	return datastore.SaveStruct(t)
}

func CreateDraftKey() *datastore.Key {
	id := uuid.NewV4()
	return datastore.NameKey(KindDraftName, id.String(), getSiteKey())
}

func GetDraftKey(id string) *datastore.Key {
	return datastore.NameKey(KindDraftName, id, getSiteKey())
}

const KindDraftPageName = "DraftPage"

type DraftPage struct {
	DraftID string
	Name    string
	PageID  string
	Seq     int
	Meta
}

func (t *DraftPage) Load(props []datastore.Property) error {
	return datastore.LoadStruct(t, props)
}

func (t *DraftPage) Save() ([]datastore.Property, error) {
	err := t.update()
	if err != nil {
		return nil, xerrors.Errorf("Meta update() error: %w", err)
	}
	return datastore.SaveStruct(t)
}

func CreateDraftPageKey() *datastore.Key {
	id := uuid.NewV4()
	return datastore.NameKey(KindDraftPageName, id.String(), getSiteKey())
}

func GetDraftPageKey(id string) *datastore.Key {
	return datastore.NameKey(KindDraftName, id, getSiteKey())
}
