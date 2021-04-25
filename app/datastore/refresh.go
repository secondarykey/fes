package datastore

import (
	"context"
	"math"
	"time"

	"golang.org/x/xerrors"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

//TODO Once
func RefreshSite(ctx context.Context) error {

	grpc.MaxConcurrentStreams(math.MaxInt32)

	cli, err := createClient(ctx,
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
	//cli, err := createClient(ctx)
	if err != nil {
		return xerrors.Errorf("createClient() error: %w", err)
	}

	var htmls []*HTML
	keys, err := getAllKind(ctx, cli, KindHTMLName, &htmls)
	if err != nil {
		return xerrors.Errorf("get all HTML error: %w", err)
	}
	err = cli.DeleteMulti(ctx, keys)
	if err != nil {
		return xerrors.Errorf("htmls delete error: %w", err)
	}
	for _, html := range htmls {
		key := CreateHTMLKey(html.GetKey().Name)
		html.LoadKey(key)
		_, err = cli.Put(ctx, key, html)
		if err != nil {
			return xerrors.Errorf("html put error: %w", err)
		}
	}

	var pages []*Page
	keys, err = getAllKind(ctx, cli, KindPageName, &pages)
	if err != nil {
		return xerrors.Errorf("get all Page error: %w", err)
	}
	err = cli.DeleteMulti(ctx, keys)
	if err != nil {
		return xerrors.Errorf("htmls delete error: %w", err)
	}
	for _, page := range pages {
		key := CreatePageKey(page.GetKey().Name)
		page.LoadKey(key)
		_, err = cli.Put(ctx, key, page)
		if err != nil {
			return xerrors.Errorf("html put error: %w", err)
		}
	}

	var pageData []*PageData
	keys, err = getAllKind(ctx, cli, KindPageDataName, &pageData)
	if err != nil {
		return xerrors.Errorf("get all PageData error: %w", err)
	}
	err = cli.DeleteMulti(ctx, keys)
	if err != nil {
		return xerrors.Errorf("page data delete error: %w", err)
	}
	for _, data := range pageData {
		key := CreatePageDataKey(data.GetKey().Name)
		data.LoadKey(key)
		_, err = cli.Put(ctx, key, data)
		if err != nil {
			return xerrors.Errorf("html put error: %w", err)
		}
	}

	var files []*File
	keys, err = getAllKind(ctx, cli, KindFileName, &files)
	if err != nil {
		return xerrors.Errorf("get all files error: %w", err)
	}
	err = cli.DeleteMulti(ctx, keys)
	if err != nil {
		return xerrors.Errorf("files delete error: %w", err)
	}
	for _, data := range files {
		key := createFileKey(data.GetKey().Name)
		data.LoadKey(key)
		_, err = cli.Put(ctx, key, data)
		if err != nil {
			return xerrors.Errorf("html put error: %w", err)
		}
	}

	var fileData []*FileData
	keys, err = getAllKind(ctx, cli, KindFileDataName, &fileData)
	if err != nil {
		return xerrors.Errorf("get all FileData error: %w", err)
	}
	err = cli.DeleteMulti(ctx, keys)
	if err != nil {
		return xerrors.Errorf("file data delete error: %w", err)
	}
	for _, data := range fileData {
		key := createFileDataKey(data.GetKey().Name)
		data.LoadKey(key)
		_, err = cli.Put(ctx, key, data)
		if err != nil {
			return xerrors.Errorf("html put error: %w", err)
		}
	}

	var templates []*Template
	keys, err = getAllKind(ctx, cli, KindTemplateName, &templates)
	if err != nil {
		return xerrors.Errorf("get all templates error: %w", err)
	}
	err = cli.DeleteMulti(ctx, keys)
	if err != nil {
		return xerrors.Errorf("templates delete error: %w", err)
	}
	for _, data := range templates {
		key := SetTemplateKey(data.GetKey().Name)
		data.LoadKey(key)
		_, err = cli.Put(ctx, key, data)
		if err != nil {
			return xerrors.Errorf("html put error: %w", err)
		}
	}

	var templateData []*TemplateData
	keys, err = getAllKind(ctx, cli, KindTemplateDataName, &templateData)
	if err != nil {
		return xerrors.Errorf("get all Page error: %w", err)
	}
	err = cli.DeleteMulti(ctx, keys)
	if err != nil {
		return xerrors.Errorf("htmls delete error: %w", err)
	}
	for _, data := range templateData {
		key := createTemplateDataKey(data.GetKey().Name)
		data.LoadKey(key)
		_, err = cli.Put(ctx, key, data)
		if err != nil {
			return xerrors.Errorf("html put error: %w", err)
		}
	}

	return nil
}
