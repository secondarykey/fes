package datastore

import (
	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
)

const KindVariableName = "Variable"

type Variable struct {
	Meta
}

func (t *Variable) Load(props []datastore.Property) error {
	return datastore.LoadStruct(t, props)
}

func (t *Variable) Save() ([]datastore.Property, error) {
	err := t.update()
	if err != nil {
		return nil, xerrors.Errorf("Meta update() error: %w", err)
	}
	return datastore.SaveStruct(t)
}

func getVariableKey(id string) *datastore.Key {
	return datastore.NameKey(KindVariableName, id, getSiteKey())
}

const KindVariableDataName = "VariableData"

type VariableData struct {
	Key     *datastore.Key `datastore:"__key__"`
	Content []byte         `datastore:",noindex"`
}

func (d *VariableData) GetKey() *datastore.Key {
	return d.Key
}

func (d *VariableData) LoadKey(k *datastore.Key) error {
	d.Key = k
	return nil
}

func getVariableDataKey(id string) *datastore.Key {
	return datastore.NameKey(KindVariableDataName, id, getSiteKey())
}
