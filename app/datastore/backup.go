package datastore

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
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
	cli, err := createClient(ctx, option.WithGRPCDialOption(grpc.WithMaxMsgSize(10_000_000)))
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

	ctx, cancel := context.WithTimeout(r.Context(), time.Second*10)
	defer cancel()

	cli, err := createClient(ctx, option.WithGRPCDialOption(grpc.WithMaxMsgSize(1024*1024*1000)))
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	keys, err := getAllKey(ctx, cli)
	if err != nil {
		return xerrors.Errorf("getAllKey() error: %w", err)
	}

	//サイトデータのみ抜き出す
	siteData, ok := backup[KindSiteName]
	if !ok {
		return xerrors.Errorf("getAllKey() error: %w", err)
	}
	if len(siteData) != 1 {
		return xerrors.Errorf("site data is once error: %w", err)
	}

	site, err := createEntity(KindSiteName, SiteEntityKey, siteData[SiteEntityKey])
	if err != nil {
		return xerrors.Errorf("createKind error: %w", err)
	}

	_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

		err = tx.DeleteMulti(keys)
		if err != nil {
			return xerrors.Errorf("backup data DeleteMulti() error: %w", err)
		}

		_, err = tx.Put(site.GetKey(), site)
		if err != nil {
			return xerrors.Errorf("site data Put() error: %w", err)
		}

		for kind, elm := range backup {

			//TODO gRPC cancel reson
			if kind == KindSiteName || kind == KindFileDataName {
				continue
			}
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

			_, err = tx.PutMulti(keys, entities)
			if err != nil {
				return xerrors.Errorf("entities PutMulti() error: %w", err)
			}
		}
		return nil
	})

	if err != nil {
		return xerrors.Errorf("restore tx error: %w", err)
	}

	//TODO gRPC cancel reson
	data := backup[KindFileDataName]
	for key, entity := range data {
		has, err := createEntity(KindFileDataName, key, entity)
		if err != nil {
			return xerrors.Errorf("createEntiry error: %w", err)
		}
		_, err = cli.Put(ctx, has.GetKey(), has)
		if err != nil {
			return xerrors.Errorf("FileData Put() error: %w", err)
		}
	}

	//TODO gRPC cancel reson

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

	key := datastore.NameKey(kind, id, nil)
	if kind != KindSiteName {
		key = datastore.NameKey(kind, id, createSiteKey())
	}

	reader := bytes.NewBuffer(data)
	decoder := gob.NewDecoder(reader)

	var err error
	var dst HasKey
	switch kind {
	case KindSiteName:
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
