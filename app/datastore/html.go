package datastore

import (
	"golang.org/x/xerrors"

	"cloud.google.com/go/datastore"
)

const KindHTMLName = "HTML"

type HTML struct {
	Content  []byte `datastore:",noindex"`
	Children int    //ignore
	PageKey  string //added
	Meta
}

func (h *HTML) Load(props []datastore.Property) error {
	return datastore.LoadStruct(h, props)
}

func (h *HTML) Save() ([]datastore.Property, error) {
	err := h.update()
	if err != nil {
		return nil, xerrors.Errorf("Meta update() error: %w", err)
	}
	return datastore.SaveStruct(h)
}

func GetHTMLKey(id string) *datastore.Key {
	return datastore.NameKey(KindHTMLName, id, getSiteKey())
}
