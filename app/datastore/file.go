package datastore

import (
	_ "image/gif"
	_ "image/png"

	"golang.org/x/xerrors"

	"cloud.google.com/go/datastore"
)

const (
	FileTypeData      = 1
	FileTypePageImage = 2
	FileTypeSystem    = 3
)

const (
	SystemFaviconID        = "system-favicon"
	draftPageImageIDSuffix = "DRAFT"
)

const KindFileName = "File"

type File struct {
	Size int64
	Type int

	Meta
}

func (f *File) Load(props []datastore.Property) error {
	return datastore.LoadStruct(f, props)
}

func (f *File) Save() ([]datastore.Property, error) {
	err := f.update()
	if err != nil {
		return nil, xerrors.Errorf("File.Save() error: %w", err)
	}
	return datastore.SaveStruct(f)
}

func getFileKey(name string) *datastore.Key {
	return datastore.NameKey(KindFileName, name, getSiteKey())
}

const KindFileDataName = "FileData"

type FileData struct {
	Key     *datastore.Key `datastore:"__key__"`
	Mime    string
	Content []byte `datastore:",noindex"`
}

func (d *FileData) GetKey() *datastore.Key {
	return d.Key
}

func (d *FileData) LoadKey(k *datastore.Key) error {
	d.Key = k
	return nil
}

func getFileDataKey(name string) *datastore.Key {
	return datastore.NameKey(KindFileDataName, name, getSiteKey())
}
