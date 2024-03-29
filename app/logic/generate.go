package logic

import (
	"app/datastore"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"context"
	"fmt"

	"golang.org/x/xerrors"
)

func CreateStaticSite(dir string) error {

	var err error

	err = os.Mkdir(dir, 0777)
	if err != nil {
		return xerrors.Errorf("make directory error: %w", err)
	}

	fmt.Println("Pages create.")
	//ディレクトリの作成
	renameP, err := createPageFiles(dir)
	if err != nil {
		return xerrors.Errorf("createPageFiles() error: %w", err)
	}

	fmt.Println("Assets create.")
	renameF, err := createAssetFiles(dir)
	if err != nil {
		return xerrors.Errorf("createAssetFiles() error: %w", err)
	}

	fmt.Println("Convert HTML")

	//_, _ = renameP, renameF
	err = convertHTML(dir, renameP, renameF)
	if err != nil {
		return xerrors.Errorf("convertHTML() error: %w", err)
	}
	return nil
}

func createPageFiles(dir string) (map[string]string, error) {

	//Pageをすべて検索
	tree, err := datastore.PageTree(context.Background())
	if err != nil {
		return nil, xerrors.Errorf("datastore.PageTree() error: %w", err)
	}

	name := tree.Page.GetKey().Name
	rtn := make(map[string]string)

	parent := "/" + dir + "/"
	url := parent + "index.html"
	path := filepath.Join(dir, "index.html")

	rtn["/"] = url
	rtn["/page/"+name] = url

	err = createPageFile(name, path)
	if err != nil {
		return nil, xerrors.Errorf("createPageFile() error: %w", err)
	}

	err = setRenameMap(parent, dir, tree.Children, rtn)
	if err != nil {
		return nil, xerrors.Errorf("setRenameMap() error: %w", err)
	}

	return rtn, nil
}

func setRenameMap(urlPath string, realPath string, trees []*datastore.Tree, rtn map[string]string) error {

	if len(trees) == 0 {
		return nil
	}

	var err error
	if _, err := os.Stat(realPath); os.IsNotExist(err) {
		err = os.Mkdir(realPath, 0777)
		if err != nil {
			return xerrors.Errorf("assets make directory error: %w", err)
		}
	}

	for idx, tree := range trees {

		p := tree.Page

		id := p.GetKey().Name
		num := fmt.Sprintf("%d", idx+1)

		name := num + ".html"
		url := urlPath + name
		rtn["/page/"+id] = url

		err = createPageFile(id, filepath.Join(realPath, name))
		if err != nil {
			return xerrors.Errorf("createPageFile() error: %w", err)
		}

		parent := urlPath + num + "/"
		rpath := filepath.Join(realPath, num)

		err = setRenameMap(parent, rpath, tree.Children, rtn)
		if err != nil {
			return xerrors.Errorf("setRenameMap() error: %w", err)
		}

	}

	return nil
}

var exts = map[string]string{
	"image/png":       "png",
	"image/jpeg":      "jpg",
	"text/css":        "css",
	"image/x-icon":    "ico",
	"application/pdf": "pdf",
}

func createAssetFiles(dir string) (map[string]string, error) {

	parent := filepath.Join(dir, "assets")
	err := os.Mkdir(parent, 0777)
	if err != nil {
		return nil, xerrors.Errorf("assets make directory error: %w", err)
	}

	files, err := datastore.GetAllFiles(context.Background())
	if err != nil {
		return nil, xerrors.Errorf("datastore.GetAllFiles() error: %w", err)
	}

	rtn := make(map[string]string)

	for idx, file := range files {

		name := file.GetKey().Name
		//FileData検索
		data, err := datastore.GetFileData(context.Background(), name)
		if err != nil {
			return nil, xerrors.Errorf("datastore.GetFileData() error: %w", err)
		}

		mime := data.Mime

		rename := name
		if strings.Index(name, ".") == -1 {
			if v, ok := exts[mime]; ok {
				rename = fmt.Sprintf("%d.%s", idx, v)
			} else {
				log.Printf("Not Found mime[%s]", mime)
			}
		}

		r := filepath.Join(parent, rename)
		err = createFile(r, data.Content)
		if err != nil {
			return nil, xerrors.Errorf("createAssetFile() error: %w", err)
		}

		rtn["/file/"+name] = "/" + dir + "/assets/" + rename
	}

	return rtn, nil
}

func createPageFile(id string, name string) error {
	//HTML検索
	html, err := datastore.GetHTML(context.Background(), id)
	if err != nil {
		return xerrors.Errorf("datastore.GetHTML() error: %w", err)
	}
	//名称でファイルを作成
	return createFile(name, html.Content)
}

func createFile(name string, data []byte) error {
	fo, err := os.Create(name)
	if err != nil {
		return xerrors.Errorf("os.Create() error: %w", err)
	}
	defer fo.Close()

	_, err = fo.Write(data)
	if err != nil {
		return xerrors.Errorf("file Write() error: %w", err)
	}

	return nil
}

func convertHTML(dir string, htmlMap, fileMap map[string]string) error {

	htmls, err := filepath.Glob(dir + "/*.html")
	if err != nil {
		return xerrors.Errorf("error: %w", err)
	}

	re, err := filepath.Glob(dir + "/**/*.html")
	if err != nil {
		return xerrors.Errorf("error: %w", err)
	}

	htmls = append(htmls, re...)

	for _, v := range htmls {
		fmt.Println("convert:" + v)
		f, err := createChangeFile(v, htmlMap, fileMap)
		if err != nil {
			return xerrors.Errorf("createChangeFile() error: %w", err)
		}

		err = os.Remove(v)
		if err != nil {
			return xerrors.Errorf("os.Remove() error: %w", err)
		}

		err = os.Rename(f, v)
		if err != nil {
			return xerrors.Errorf("os.Rename() error: %w", err)
		}
	}

	return nil
}

func createChangeFile(name string, htmlMap, fileMap map[string]string) (string, error) {

	f, err := os.Open(name)
	if err != nil {
		return "", xerrors.Errorf("os.Open() error: %w", err)
	}
	defer f.Close()

	var builder strings.Builder
	_, err = io.Copy(&builder, f)
	if err != nil {
		return "", xerrors.Errorf("io.Copy() error: %w", err)
	}

	buf := builder.String()
	top := ""
	for key, v := range htmlMap {
		if key != "/" {
			buf = strings.ReplaceAll(buf, key, v)
		} else {
			top = v
		}
	}

	for key, v := range fileMap {
		buf = strings.ReplaceAll(buf, key, v)
	}

	buf = strings.ReplaceAll(buf, `="/"`, fmt.Sprintf(`="%s"`, top))

	//TODO キャッシュを利用した場合の日付変換ができてない

	tmpName := name + ".tmp"
	tmp, err := os.Create(tmpName)
	if err != nil {
		return "", xerrors.Errorf("os.Open() error: %w", err)
	}
	defer tmp.Close()

	_, err = tmp.Write([]byte(buf))
	if err != nil {
		return "", xerrors.Errorf("Write() error: %w", err)
	}

	return tmpName, nil
}

func GenerateFiles(dir string) error {

	err := os.Mkdir(dir, 0777)
	if err != nil {
		return xerrors.Errorf("make directory error: %w", err)
	}

	files, err := datastore.GetAllFiles(context.Background())
	if err != nil {
		return xerrors.Errorf("datastore.GetAllFiles() error: %w", err)
	}

	for _, file := range files {

		name := file.GetKey().Name
		//FileData検索
		data, err := datastore.GetFileData(context.Background(), name)
		if err != nil {
			return xerrors.Errorf("datastore.GetFileData() error: %w", err)
		}

		mime := data.Mime

		rename := name
		if strings.Index(name, ".") == -1 {
			if v, ok := exts[mime]; ok {
				rename = fmt.Sprintf("%s.%s", name, v)
			} else {
				log.Printf("Not Found mime[%s]", mime)
			}
		}

		fmt.Println("Generate", rename)
		path := filepath.Join(dir, rename)
		err = createFile(path, data.Content)
		if err != nil {
			return xerrors.Errorf("createFile() error: %w", err)
		}
	}
	return nil
}
