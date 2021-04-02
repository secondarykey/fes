package datastore

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/http"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
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
		return nil, xerrors.Errorf("GetAll() [%s] error: %w", name, err)
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

	err = cli.DeleteMulti(ctx, keys)
	if err != nil {
		return xerrors.Errorf("backup data DeleteMulti() error: %w", err)
	}

	for kind, elm := range backup {

		//Siteの場合を行わない

		var entities []HasKey
		var keys []*datastore.Key
		for key, data := range elm {
			has, err := createEntity(kind, key, data)
			if err != nil {
				return xerrors.Errorf("createKind error: %w", err)
			}
			entities = append(entities, has)
			keys = append(keys, has.GetKey())
		}

		if kind != "FileData" {
			_, err = cli.PutMulti(ctx, keys, entities)
			if err != nil {
				return xerrors.Errorf("entities PutMulti() error: %w", err)
			}
		} else {
			width := 10
			flag := true
			leng := len(keys)
			idx := 0

			for flag {
				last := idx + width
				if last >= leng {
					flag = false
					last = leng
				}
				wkk := keys[idx:last]
				wke := entities[idx:last]

				_, err = cli.PutMulti(ctx, wkk, wke)
				if err != nil {
					return xerrors.Errorf("entities PutMulti() error: %w", err)
				}
				idx = last
			}
		}
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

	rtn = append(rtn, createSiteKey())

	keys, err := getAllKindKey(ctx, cli, KindHTMLName)
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

func createEntity(kind string, id string, data []byte) (HasKey, error) {

	key := datastore.NameKey(kind, id, createSiteKey())
	reader := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(reader)

	var err error
	var dst HasKey
	switch kind {
	case KindSiteName:
		//TODO Siteは単独で行う
		dst = &Site{}
	case KindHTMLName:
		dst = &HTML{}
	case KindPageName:
		dst = &Page{}
	case KindPageDataName:
		dst = &PageData{}
	case KindFileName:
		dst = &File{}
	case KindFileDataName:
		dst = &FileData{}
	case KindTemplateName:
		dst = &Template{}
	case KindTemplateDataName:
		dst = &TemplateData{}
	default:
		return nil, fmt.Errorf("NotFound Kind[%s]", kind)
	}

	err = decoder.Decode(dst)
	if err != nil {
		return nil, err
	}
	dst.LoadKey(key)
	return dst, nil
}
