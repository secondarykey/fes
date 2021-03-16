package datastore

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"

	"golang.org/x/xerrors"

	"cloud.google.com/go/datastore"
)

type GobKind map[string][]byte
type BackupData map[string]GobKind

func init() {
	gob.Register(HTML{})
	gob.Register(Page{})
	gob.Register(PageData{})
	gob.Register(File{})
	gob.Register(FileData{})
	gob.Register(Template{})
	gob.Register(TemplateData{})
}

func GetBackupData(r *http.Request) (BackupData, error) {

	backup := make(BackupData)
	ctx := r.Context()
	cli, err := createClient(ctx)
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	backup[KindSiteName], err = createSiteGob(ctx, cli)
	if err != nil {
		return nil, xerrors.Errorf("createSiteGob() error: %w", err)
	}
	backup[KindHTMLName], err = createHTMLGob(ctx, cli)
	if err != nil {
		return nil, xerrors.Errorf("createHTMLGob() error: %w", err)
	}
	backup[KindPageName], err = createPageGob(ctx, cli)
	if err != nil {
		return nil, xerrors.Errorf("createPageGob() error: %w", err)
	}
	backup[KindPageDataName], err = createPageDataGob(ctx, cli)
	if err != nil {
		return nil, xerrors.Errorf("createPageDataGob() error: %w", err)
	}
	backup[KindFileName], err = createFileGob(ctx, cli)
	if err != nil {
		return nil, xerrors.Errorf("createFileGob() error: %w", err)
	}
	backup[KindFileDataName], err = createFileDataGob(ctx, cli)
	if err != nil {
		return nil, xerrors.Errorf("createFileDataGob() error: %w", err)
	}
	backup[KindTemplateName], err = createTemplateGob(ctx, cli)
	if err != nil {
		return nil, xerrors.Errorf("createTemplateGob() error: %w", err)
	}
	backup[KindTemplateDataName], err = createTemplateDataGob(ctx, cli)
	if err != nil {
		return nil, xerrors.Errorf("createTemplateDataGob() error: %w", err)
	}
	return backup, nil
}

func createSiteGob(ctx context.Context, cli *datastore.Client) (GobKind, error) {
	rtn := make(GobKind)

	var data []*Site
	keys, err := getAllKind(ctx, cli, KindSiteName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].Name] = convertGob(data)
	}
	return rtn, nil
}

func createHTMLGob(ctx context.Context, cli *datastore.Client) (GobKind, error) {
	rtn := make(GobKind)
	var data []*HTML
	keys, err := getAllKind(ctx, cli, KindHTMLName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].Name] = convertGob(data)
	}
	return rtn, nil
}

func createPageGob(ctx context.Context, cli *datastore.Client) (GobKind, error) {
	rtn := make(GobKind)
	var data []*Page
	keys, err := getAllKind(ctx, cli, KindPageName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].Name] = convertGob(data)
	}
	return rtn, nil
}

func createPageDataGob(ctx context.Context, cli *datastore.Client) (GobKind, error) {
	rtn := make(GobKind)
	var data []*PageData
	keys, err := getAllKind(ctx, cli, KindPageDataName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].Name] = convertGob(data)
	}
	return rtn, nil
}

func createFileGob(ctx context.Context, cli *datastore.Client) (GobKind, error) {
	rtn := make(GobKind)
	var data []*File
	keys, err := getAllKind(ctx, cli, KindFileName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].Name] = convertGob(data)
	}
	return rtn, nil
}

func createFileDataGob(ctx context.Context, cli *datastore.Client) (GobKind, error) {
	rtn := make(GobKind)
	var data []*FileData
	keys, err := getAllKind(ctx, cli, KindFileDataName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].Name] = convertGob(data)
	}
	return rtn, nil
}

func createTemplateGob(ctx context.Context, cli *datastore.Client) (GobKind, error) {
	rtn := make(GobKind)
	var data []*Template
	keys, err := getAllKind(ctx, cli, KindTemplateName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].Name] = convertGob(data)
	}
	return rtn, nil
}

func createTemplateDataGob(ctx context.Context, cli *datastore.Client) (GobKind, error) {
	rtn := make(GobKind)
	var data []*TemplateData
	keys, err := getAllKind(ctx, cli, KindTemplateDataName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].Name] = convertGob(data)
	}
	return rtn, nil
}

func getAllKind(ctx context.Context, cli *datastore.Client, name string, dst interface{}) ([]*datastore.Key, error) {
	q := datastore.NewQuery(name)
	keys, err := cli.GetAll(ctx, q, dst)
	if err != nil {
		return nil, xerrors.Errorf("GetAll() error: %w", err)
	}
	return keys, nil
}

func convertGob(dst interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	gob.NewEncoder(buf).Encode(dst)
	return buf.Bytes()
}

func PutBackupData(r *http.Request, backup BackupData) error {

	ctx := r.Context()

	cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	keys, err := getAllKey(ctx, cli)
	if err != nil {
		return xerrors.Errorf("getAllKey() error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		err := tx.DeleteMulti(keys)
		if err != nil {
			return xerrors.Errorf("backup data DeleteMulti() error: %w", err)
		}

		for kind, elm := range backup {
			log.Println(kind)
			for key, data := range elm {
				err = putKind(tx, kind, key, data)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	if err != nil {
		return xerrors.Errorf("transaction error: %w", err)
	}
	return nil
}

func getAllKindKey(c context.Context, cli *datastore.Client, name string) ([]*datastore.Key, error) {
	q := datastore.NewQuery(name).KeysOnly()
	keys, err := cli.GetAll(c, q, nil)
	if err != nil {
		return nil, xerrors.Errorf("GetAll() error: %w", err)
	}
	return keys, nil
}

func getAllKey(ctx context.Context, cli *datastore.Client) ([]*datastore.Key, error) {

	var rtn []*datastore.Key

	keys, err := getAllKindKey(ctx, cli, KindSiteName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)

	keys, err = getAllKindKey(ctx, cli, KindHTMLName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)

	keys, err = getAllKindKey(ctx, cli, KindPageName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)

	keys, err = getAllKindKey(ctx, cli, KindPageDataName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	keys, err = getAllKindKey(ctx, cli, KindFileName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	keys, err = getAllKindKey(ctx, cli, KindFileDataName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	keys, err = getAllKindKey(ctx, cli, KindTemplateName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	keys, err = getAllKindKey(ctx, cli, KindTemplateDataName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	return rtn, nil
}

func putKind(tx *datastore.Transaction, kind string, id string, data []byte) error {

	key := datastore.NameKey(kind, id, nil)
	reader := bytes.NewBuffer(data)

	var err error
	switch kind {
	case KindSiteName:
		dst := &Site{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.LoadKey(key)
		_, err = tx.Put(key, dst)
		if err != nil {
			return err
		}
	case KindHTMLName:
		dst := &HTML{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.LoadKey(key)
		_, err = tx.Put(key, dst)
		if err != nil {
			return err
		}
	case KindPageName:
		dst := &Page{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.LoadKey(key)
		_, err = tx.Put(key, dst)
		if err != nil {
			return err
		}
	case KindPageDataName:
		dst := &PageData{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.LoadKey(key)
		_, err = tx.Put(key, dst)
		if err != nil {
			return err
		}
	case KindFileName:
		dst := &File{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.LoadKey(key)
		_, err = tx.Put(key, dst)
		if err != nil {
			return err
		}
	case KindFileDataName:
		dst := &FileData{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.LoadKey(key)
		_, err = tx.Put(key, dst)
		if err != nil {
			return err
		}
	case KindTemplateName:
		dst := &Template{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.LoadKey(key)
		_, err = tx.Put(key, dst)
		if err != nil {
			return err
		}
	case KindTemplateDataName:
		dst := &TemplateData{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.LoadKey(key)
		_, err = tx.Put(key, dst)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("NotFound Kind[%s]", kind)
	}

	return nil
}
