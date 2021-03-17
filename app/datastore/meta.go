package datastore

import (
	"strconv"
	"time"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
)

type HasKey interface {
	LoadKey(*datastore.Key) error
	GetKey() *datastore.Key
}

type Meta struct {
	Key       *datastore.Key `datastore:"__key__" json:"key"`
	Version   int            `datastore:",noindex" json:"version"`
	Deleted   bool           `json:"deleted"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

func (m *Meta) LoadKey(key *datastore.Key) error {
	m.Key = key
	return nil
}

func (m *Meta) GetKey() *datastore.Key {
	return m.Key
}

func (m *Meta) load(ver string) error {
	if ver == "" {
		return nil
	}
	v, err := strconv.Atoi(ver)
	if err != nil {
		return xerrors.Errorf("version strconv.Atoi() error: %w", err)
	}

	if v <= 0 && v != m.Version {
		return xerrors.Errorf("version optimistic lock error")
	}
	return nil
}

func (m *Meta) update(ver string) error {
	now := time.Now()
	if m.Version == 0 {
		m.CreatedAt = now
	} else {
		if ver != "" {
			v, err := strconv.Atoi(ver)
			if err != nil {
				return xerrors.Errorf("version strconv.Atoi() error: %w", err)
			}
			if v <= 0 && v != m.Version {
				return xerrors.Errorf("version optimistic lock error")
			}
		}
	}
	m.Version++
	m.UpdatedAt = now
	return nil
}
