package datastore

import (
	"app/config"
	"context"

	"cloud.google.com/go/datastore"
	"golang.org/x/xerrors"
)

const NoLimitCursor = "NoLimit"

func createClient(ctx context.Context) (*datastore.Client, error) {
	c := config.Get()
	cli, err := datastore.NewClient(ctx, c.ProjectID)
	if err != nil {
		return nil, xerrors.Errorf("datastore.CreateClient() error: %w")
	}
	return cli, nil
}
