package config

import (
	"os"
	"strconv"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/xerrors"
)

type Option func(*Config) error

func SetPort() Option {
	return func(c *Config) error {
		p := os.Getenv("PORT")
		if p == "" {
			p = "8080"
		}
		v, err := strconv.Atoi(p)
		if err != nil {
			return xerrors.Errorf("strconv.Atoi() error: %w", err)
		}
		c.Port = v
		return nil
	}
}

const (
	DatastoreEmulatorHost = "DATASTORE_EMULATOR_HOST"
	DatastoreProjectID    = "DATASTORE_PROJECT_ID"
	DatastoreDataset      = "DATASTORE_DATASET"
)

func SetProjectID() Option {
	return func(c *Config) error {
		if !c.DevelopMode {
			p, err := metadata.ProjectID()
			if err != nil {
				return xerrors.Errorf("metadata.ProjectID() error: %w", err)
			}
			c.ProjectID = p
		}
		return nil
	}
}

func SetDatastore() Option {
	return func(c *Config) error {
		if c.DevelopMode {
			//emulatorを設定
			host := os.Getenv(DatastoreEmulatorHost)
			if host == "" {
				host = "localhost:8081"
			}
			os.Setenv(DatastoreEmulatorHost, host)
		}

		ds := os.Getenv(DatastoreDataset)
		if ds == "" {
			ds = c.ProjectID
		}
		os.Setenv(DatastoreDataset, ds)

		p := os.Getenv(DatastoreProjectID)
		if p == "" {
			p = c.ProjectID
		}
		os.Setenv(DatastoreProjectID, p)

		return nil
	}
}
