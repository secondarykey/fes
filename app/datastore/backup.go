package datastore

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"

	"github.com/knightso/base/gae/ds"
	"google.golang.org/appengine"
	"google.golang.org/appengine/datastore"
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
	var err error
	backup[KindSiteName], err = createSiteGob(r)
	if err != nil {
		return nil, err
	}
	backup[KindHTMLName], err = createHTMLGob(r)
	if err != nil {
		return nil, err
	}
	backup[KindPageName], err = createPageGob(r)
	if err != nil {
		return nil, err
	}
	backup[KindPageDataName], err = createPageDataGob(r)
	if err != nil {
		return nil, err
	}
	backup[KindFileName], err = createFileGob(r)
	if err != nil {
		return nil, err
	}
	backup[KindFileDataName], err = createFileDataGob(r)
	if err != nil {
		return nil, err
	}
	backup[KindTemplateName], err = createTemplateGob(r)
	if err != nil {
		return nil, err
	}
	backup[KindTemplateDataName], err = createTemplateDataGob(r)
	if err != nil {
		return nil, err
	}
	return backup, nil
}

func createSiteGob(r *http.Request) (GobKind, error) {
	rtn := make(GobKind)
	var data []*Site
	keys, err := getAllKind(r, KindSiteName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].StringID()] = convertGob(data)
	}
	return rtn, nil
}

func createHTMLGob(r *http.Request) (GobKind, error) {
	rtn := make(GobKind)
	var data []*HTML
	keys, err := getAllKind(r, KindHTMLName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].StringID()] = convertGob(data)
	}
	return rtn, nil
}

func createPageGob(r *http.Request) (GobKind, error) {
	rtn := make(GobKind)
	var data []*Page
	keys, err := getAllKind(r, KindPageName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].StringID()] = convertGob(data)
	}
	return rtn, nil
}

func createPageDataGob(r *http.Request) (GobKind, error) {
	rtn := make(GobKind)
	var data []*PageData
	keys, err := getAllKind(r, KindPageDataName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].StringID()] = convertGob(data)
	}
	return rtn, nil
}

func createFileGob(r *http.Request) (GobKind, error) {
	rtn := make(GobKind)
	var data []*File
	keys, err := getAllKind(r, KindFileName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].StringID()] = convertGob(data)
	}
	return rtn, nil
}

func createFileDataGob(r *http.Request) (GobKind, error) {
	rtn := make(GobKind)
	var data []*FileData
	keys, err := getAllKind(r, KindFileDataName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].StringID()] = convertGob(data)
	}
	return rtn, nil
}

func createTemplateGob(r *http.Request) (GobKind, error) {
	rtn := make(GobKind)
	var data []*Template
	keys, err := getAllKind(r, KindTemplateName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].StringID()] = convertGob(data)
	}
	return rtn, nil
}

func createTemplateDataGob(r *http.Request) (GobKind, error) {
	rtn := make(GobKind)
	var data []*TemplateData
	keys, err := getAllKind(r, KindTemplateDataName, &data)
	if err != nil {
		return nil, err
	}
	for idx, data := range data {
		rtn[keys[idx].StringID()] = convertGob(data)
	}
	return rtn, nil
}

func getAllKind(r *http.Request, name string, dst interface{}) ([]*datastore.Key, error) {
	c := appengine.NewContext(r)
	q := datastore.NewQuery(name)
	keys, err := q.GetAll(c, dst)
	return keys, err
}

func convertGob(dst interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	gob.NewEncoder(buf).Encode(dst)
	return buf.Bytes()
}

func PutBackupData(r *http.Request, backup BackupData) error {

	c := appengine.NewContext(r)

	option := &datastore.TransactionOptions{XG: true}
	keys, err := getAllKey(c)
	if err != nil {
		return err
	}

	err = datastore.RunInTransaction(c, func(ctx context.Context) error {
		for _, key := range keys {
			err := ds.Delete(c, key)
			if err != nil {
				return err
			}
		}

		for kind, elm := range backup {
			log.Println(kind)
			for key, data := range elm {
				log.Println(key)
				err = putKind(c, kind, key, data)
				if err != nil {
					return err
				}
			}
		}

		return nil
	}, option)
	return err
}

func getAllKindKey(c context.Context, name string) ([]*datastore.Key, error) {
	q := datastore.NewQuery(name).KeysOnly()
	return q.GetAll(c, nil)
}

func getAllKey(ctx context.Context) ([]*datastore.Key, error) {

	var rtn []*datastore.Key

	keys, err := getAllKindKey(ctx, KindSiteName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)

	keys, err = getAllKindKey(ctx, KindHTMLName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)

	keys, err = getAllKindKey(ctx, KindPageName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)

	keys, err = getAllKindKey(ctx, KindPageDataName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	keys, err = getAllKindKey(ctx, KindFileName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	keys, err = getAllKindKey(ctx, KindFileDataName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	keys, err = getAllKindKey(ctx, KindTemplateName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	keys, err = getAllKindKey(ctx, KindTemplateDataName)
	if err != nil {
		return nil, err
	}
	rtn = append(rtn, keys...)
	return rtn, nil
}

func putKind(ctx context.Context, kind string, id string, data []byte) error {

	key := datastore.NewKey(ctx, kind, id, 0, nil)
	reader := bytes.NewBuffer(data)

	var err error
	switch kind {
	case KindSiteName:
		dst := &Site{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.SetKey(key)
		err = ds.Put(ctx, dst)
		if err != nil {
			return err
		}
	case KindHTMLName:
		dst := &HTML{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.SetKey(key)
		err = ds.Put(ctx, dst)
		if err != nil {
			return err
		}
	case KindPageName:
		dst := &Page{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.SetKey(key)
		err = ds.Put(ctx, dst)
		if err != nil {
			return err
		}
	case KindPageDataName:
		dst := &PageData{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.SetKey(key)
		err = ds.Put(ctx, dst)
		if err != nil {
			return err
		}
	case KindFileName:
		dst := &File{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.SetKey(key)
		err = ds.Put(ctx, dst)
		if err != nil {
			return err
		}
	case KindFileDataName:
		dst := &FileData{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.SetKey(key)
		err = ds.Put(ctx, dst)
		if err != nil {
			return err
		}
	case KindTemplateName:
		dst := &Template{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.SetKey(key)
		err = ds.Put(ctx, dst)
		if err != nil {
			return err
		}
	case KindTemplateDataName:
		dst := &TemplateData{}
		err = gob.NewDecoder(reader).Decode(dst)
		if err != nil {
			return err
		}
		dst.SetKey(key)
		err = ds.Put(ctx, dst)
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("NotFound Kind[%s]", kind)
	}

	return nil
}
