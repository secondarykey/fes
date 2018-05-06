package datastore

import (
	"net/http"

	"github.com/satori/go.uuid"
	"github.com/gorilla/mux"
	kerr "github.com/knightso/base/errors"
	"github.com/knightso/base/gae/ds"

	"google.golang.org/appengine/datastore"
	"google.golang.org/appengine"
	"strconv"
)

func PutTemplate(r *http.Request) error {

	var err error

	vars := mux.Vars(r)
	id := vars["key"]

	c := appengine.NewContext(r)
	tmpKey := datastore.NewKey(c,KIND_TEMPLATE,id,0,nil)
	tmpDataKey := datastore.NewKey(c,KIND_TEMPLATEDATA,id,0,nil)

	template := Template{}
	templateData := TemplateData{}

	template.SetKey(tmpKey)
	templateData.SetKey(tmpDataKey)

	template.Name = r.FormValue("name")
	template.Type,err = strconv.Atoi(r.FormValue("type"))
	if err != nil {
		return err
	}
	templateData.Content = datastore.ByteString(r.FormValue("template"))

	err = ds.Put(c,&template)
	if err != nil {
		return err
	}

	err = ds.Put(c,&templateData)
	if err != nil {
		return err
	}

	return nil
}

const KIND_TEMPLATE = "Template"
type Template struct {
	Name string
	Type int
	ds.Meta
}

func CreateTemplateKey(r *http.Request) *datastore.Key {
	c := appengine.NewContext(r)
	id,err := uuid.NewV4()
	if err != nil {
	}
	return datastore.NewKey(c,KIND_TEMPLATE,id.String(),0,nil)
}

func SelectTemplate(r *http.Request,id string) (*Template,error){
	temp := Template{}
	c := appengine.NewContext(r)
	//Method
	key := datastore.NewKey(c,KIND_TEMPLATE,id,0,nil)
	err := ds.Get(c,key,&temp)
	if err != nil {
		if kerr.Root(err) != datastore.ErrNoSuchEntity {
			return nil,err
		} else {
			return nil,nil
		}
	}
	return &temp,nil
}

func SelectTemplates(r *http.Request) ([]Template,error){

	var rtn []Template

	c := appengine.NewContext(r)

	q := datastore.NewQuery(KIND_TEMPLATE)
	t := q.Run(c)

	for {
		var tmp Template
		key , err := t.Next(&tmp)
		if err == datastore.Done {
			break
		}

		if err != nil {
			return nil,err
		}
		tmp.SetKey(key)
		rtn = append(rtn,tmp)
	}
	return rtn,nil
}

const KIND_TEMPLATEDATA = "TemplateData"

type TemplateData struct {
	key     *datastore.Key
	Content datastore.ByteString `datastore:",noindex"`
}

func (d *TemplateData) GetKey() *datastore.Key {
	return d.key
}

func (d *TemplateData) SetKey(k *datastore.Key) {
	d.key = k
}

func SelectTemplateData(r *http.Request,id string) (*TemplateData,error){
	temp := TemplateData{}
	c := appengine.NewContext(r)
	//Method
	key := datastore.NewKey(c,KIND_TEMPLATEDATA,id,0,nil)
	err := ds.Get(c,key,&temp)
	if err != nil {
		return nil,err
	}
	return &temp,nil
}

func RemoveTemplate(r *http.Request,id string) error {

	var err error
	c := appengine.NewContext(r)
	key := datastore.NewKey(c,KIND_TEMPLATE,id,0,nil)
	err = ds.Delete(c,key)
	if err != nil {
		return err
	}

	dataKey := datastore.NewKey(c,KIND_TEMPLATEDATA,id,0,nil)
	err = ds.Delete(c,dataKey)
	if err != nil {
		return err
	}

	return nil
}
