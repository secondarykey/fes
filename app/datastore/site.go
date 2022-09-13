package datastore

import (
	"fmt"

	"golang.org/x/xerrors"

	"cloud.google.com/go/datastore"
)

const KindSiteName = "Site"
const SiteEntityKey = "fixing"

var (
	SiteNotFoundError = fmt.Errorf("site not found")
)

type Site struct {
	Name        string
	Description string
	Root        string
	ManageURL   string
	Managers    []string

	TargetVersion string `datastore:"-"`
	Meta

	//Deprecated
	HTMLCache     bool
	TemplateCache bool
	FileCache     bool
	PageCache     bool
}

func (s *Site) Load(props []datastore.Property) error {
	return datastore.LoadStruct(s, props)
}

func (s *Site) Save() ([]datastore.Property, error) {
	err := s.update()
	if err != nil {
		return nil, xerrors.Errorf("Meta update() error: %w", err)
	}
	return datastore.SaveStruct(s)
}

func getSiteKey() *datastore.Key {
	return datastore.NameKey(KindSiteName, SiteEntityKey, nil)
}
