package datastore

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"math"
	"reflect"
	"time"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type GobKind map[string][]byte
type BackupData map[string]GobKind

func init() {

	/*
		grpc.MaxConcurrentStreams(math.MaxInt32)

		_, err := createClient(ctx,
			option.WithGRPCDialOption(
				grpc.WithDefaultCallOptions(
					grpc.MaxCallRecvMsgSize(10_000_000_000_000),
					grpc.MaxCallSendMsgSize(10_000_000_000_000),
				)),
			option.WithGRPCDialOption(
				grpc.WithKeepaliveParams(keepalive.ClientParameters{
					Time:                time.Second * 10,
					Timeout:             time.Second * 50,
					PermitWithoutStream: true,
				})))
	*/

	gob.Register(Site{})
	gob.Register(File{})
	gob.Register(FileData{})
	gob.Register(Variable{})
	gob.Register(VariableData{})
	gob.Register(Template{})
	gob.Register(TemplateData{})
	gob.Register(Page{})
	gob.Register(PageData{})
	gob.Register(HTML{})
}

func (dao *Dao) CreateBackupData(ctx context.Context) (BackupData, error) {

	backup := make(BackupData)

	cli, err := dao.createClient(ctx, option.WithGRPCDialOption(grpc.WithMaxMsgSize(1024*1024*1000)))
	if err != nil {
		return nil, xerrors.Errorf("createClient() error: %w", err)
	}

	kinds := []string{KindSiteName,
		KindFileName, KindFileDataName,
		KindVariableName, KindVariableDataName,
		KindTemplateName, KindTemplateDataName,
		KindPageName, KindPageDataName,
		KindHTMLName}

	fmt.Println("*************** Backup")
	defer fmt.Println("**********************")

	for _, name := range kinds {
		fmt.Println("Kind:", name)
		backup[name], err = createGobKind(ctx, cli, name)
		if err != nil {
			return nil, xerrors.Errorf("createSiteGob() error: %w", err)
		}
		fmt.Println("Length:", len(backup[name]))
	}

	return backup, nil
}

var (
	typeOfPropertyLoadSaver = reflect.TypeOf((*datastore.PropertyLoadSaver)(nil)).Elem()
	typeOfPropertyList      = reflect.TypeOf(datastore.PropertyList(nil))
)

func printMultiArg(dst interface{}) {

	fmt.Println("printMultiArg() ----------")
	defer fmt.Println("--------------------------")

	//GetAll() check
	dv := reflect.ValueOf(dst)
	if dv.IsNil() {
		fmt.Println("dv.IsNil()")
		return
	} else if dv.Kind() != reflect.Ptr {
		fmt.Printf("dv.Kind() is not reflect.Ptr[%v]\n", dv.Kind())
		return
	}
	arg := dv.Elem()

	//checkMultiArg()
	fmt.Println("arg.Kind()", arg.Kind())
	if arg.Kind() != reflect.Slice {
		fmt.Printf("v.Kind() is not reflect.Slice[%v]\n", arg.Kind())
		return
	}

	fmt.Println("arg.Type()", arg.Type())
	if arg.Type() == typeOfPropertyList {
		fmt.Printf("v.Type() is  typeOfPropertyList[%v]\n")
		return
	}

	elemType := arg.Type().Elem()
	if reflect.PtrTo(elemType).Implements(typeOfPropertyLoadSaver) {
		fmt.Println("ok", "typeOfPropetyLoadSaver")
		return
	}

	fmt.Println("elem.Kind()", elemType.Kind())
	switch elemType.Kind() {
	case reflect.Struct:
		fmt.Println("ok", "return multiArgTypeStruct elemType")
		return
	case reflect.Interface:
		fmt.Println("ok", "return multiArgTypeInterface elemType")
		return
	case reflect.Ptr:
		elemType = elemType.Elem()
		if elemType.Kind() == reflect.Struct {
			fmt.Println("ok", "return multiArgTypeStructPtr elemType")
			return
		}
	}

	fmt.Println("ok", "return multiArgTypeInValid nil")
	return
}

func createGobKind(ctx context.Context, cli *datastore.Client, name string) (GobKind, error) {

	//TODO interface{} で受けると型情報がおかしくなる
	//entitySlice := createEntitySlice(name)

	q := datastore.NewQuery(name)
	var keys []*datastore.Key
	var err error

	var dst interface{}
	switch name {
	case KindSiteName:
		var ref []*Site
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	case KindFileName:
		var ref []*File
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	case KindFileDataName:
		var ref []*FileData
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	case KindVariableName:
		var ref []*Variable
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	case KindVariableDataName:
		var ref []*VariableData
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	case KindTemplateName:
		var ref []*Template
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	case KindTemplateDataName:
		var ref []*TemplateData
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	case KindPageName:
		var ref []*Page
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	case KindPageDataName:
		var ref []*PageData
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	case KindHTMLName:
		var ref []*HTML
		keys, err = cli.GetAll(ctx, q, &ref)
		dst = ref
	default:
		return nil, xerrors.Errorf("Type is not Found[%s]", name)
	}

	if err != nil {
		return nil, xerrors.Errorf("GetAll() [%s] error: %w", name, err)
	}

	rtn := make(GobKind)
	s := reflect.ValueOf(dst)

	if s.Len() != len(keys) {
		return rtn, fmt.Errorf("error: data and key length are different(%d != %d)", s.Len(), len(keys))
	}

	for idx := 0; idx < s.Len(); idx++ {
		rtn[keys[idx].Name] = convertGob(s.Index(idx).Interface())
	}
	return rtn, nil
}

func convertGob(dst interface{}) []byte {
	buf := bytes.NewBuffer(nil)
	gob.NewEncoder(buf).Encode(dst)
	return buf.Bytes()
}

type Kinds struct {
	Names []string
	Multi bool
}

func NewKinds(names ...string) *Kinds {
	var k Kinds
	k.Names = names
	k.Multi = true
	return &k
}

func (k *Kinds) NonMulti() *Kinds {
	k.Multi = false
	return k
}

func (k Kinds) String() string {
	return fmt.Sprintf("%v(Multi:%t)", k.Names, k.Multi)
}

func (dao *Dao) PutBackupData(ctx context.Context, backup BackupData) error {

	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	grpc.MaxConcurrentStreams(math.MaxInt32)
	grpc.MaxRecvMsgSize(1024 * 1024 * 20)

	cli, err := dao.createClient(ctx,
		option.WithGRPCDialOption(
			grpc.WithDefaultCallOptions(
				grpc.MaxCallRecvMsgSize(1024*1024*20),
				grpc.MaxCallSendMsgSize(1024*1024*20),
			)),
		option.WithGRPCDialOption(
			grpc.WithKeepaliveParams(keepalive.ClientParameters{
				Time:                time.Second * 10,
				Timeout:             time.Second * 50,
				PermitWithoutStream: true,
			})))
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	kg := []*Kinds{
		NewKinds(KindSiteName),
		NewKinds(KindFileName, KindFileDataName).NonMulti(),
		NewKinds(KindVariableName, KindVariableDataName),
		NewKinds(KindTemplateName, KindTemplateDataName),
		NewKinds(KindPageName, KindPageDataName),
		NewKinds(KindHTMLName),
	}

	fmt.Println("******** Restore Start")
	defer fmt.Println("**********************")

	for _, kinds := range kg {

		fmt.Println("**** Kind Group", kinds)

		keys, err := getKeys(ctx, cli, kinds.Names...)
		if err != nil {
			return xerrors.Errorf("getKeys() error: %w", err)
		}

		if kinds.Multi {
			_, err = cli.RunInTransaction(ctx, func(tx *datastore.Transaction) error {

				fmt.Println("DELETE", len(keys))
				err = tx.DeleteMulti(keys)
				if err != nil {
					return xerrors.Errorf("backup data DeleteMulti() error: %w", err)
				}

				for _, kind := range kinds.Names {

					elm := backup[kind]

					fmt.Println("Kind:", kind, len(elm))
					var entities []HasKey
					var keys []*datastore.Key

					byt := 0
					for key, data := range elm {

						byt += len(data)
						has, err := createEntity(kind, key, data)
						if err != nil {
							return xerrors.Errorf("createKind error: %w", err)
						}

						entities = append(entities, has)
						keys = append(keys, has.GetKey())
					}

					fmt.Println("Size:", byt)

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
		} else {
			//TODO gRPC cancel reson
			// ServerOption 4MB over
			fmt.Println("NonMulti Delete", len(keys))
			err = cli.DeleteMulti(ctx, keys)
			if err != nil {
				return xerrors.Errorf("cli DeleteMulti() error: %w", err)
			}
			for _, kind := range kinds.Names {
				elm := backup[kind]
				fmt.Println("Kind:", kind, len(elm))
				for key, data := range elm {
					has, err := createEntity(kind, key, data)
					if err != nil {
						return xerrors.Errorf("createKind error: %w", err)
					}
					_, err = cli.Put(ctx, has.GetKey(), has)
					if err != nil {
						return xerrors.Errorf("Put() error: %w", err)
					}
				}
			}
		}
	}

	return nil
}

func getKindKeys(c context.Context, cli *datastore.Client, name string) ([]*datastore.Key, error) {
	q := datastore.NewQuery(name).KeysOnly()
	keys, err := cli.GetAll(c, q, nil)
	if err != nil {
		return nil, xerrors.Errorf("GetAll() error: %w", err)
	}
	return keys, nil
}

func getKeys(ctx context.Context, cli *datastore.Client, kinds ...string) ([]*datastore.Key, error) {

	var rtn []*datastore.Key
	for _, kind := range kinds {
		if kind == KindSiteName {
			rtn = append(rtn, createSiteKey())
		} else {
			keys, err := getKindKeys(ctx, cli, kind)
			if err != nil {
				return nil, xerrors.Errorf("getKindKeys(): %w", err)
			}
			rtn = append(rtn, keys...)
		}
	}
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
	case KindVariableName:
		dst = &Variable{}
	case KindVariableDataName:
		dst = &VariableData{}
	default:
		return nil, fmt.Errorf("NotFound Kind[%s]", kind)
	}

	err = decoder.Decode(dst)
	if err != nil {
		return nil, xerrors.Errorf("gob Decode() error: %w", err)
	}
	dst.LoadKey(key)
	return dst, nil
}
