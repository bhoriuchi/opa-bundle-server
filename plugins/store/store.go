package store

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/open-policy-agent/opa/ast"
	"github.com/open-policy-agent/opa/bundle"
	"github.com/open-policy-agent/opa/compile"
)

var (
	Providers = map[string]NewStoreFunc{}
)

type NewStoreFunc func(opts *Options) (Store, error)

type Options struct {
	Name   string
	Config interface{}
	Logger logger.Logger
}

type EntryList []*Entry

func (e EntryList) Len() int {
	return len(e)
}
func (e EntryList) Swap(i, j int) {
	e[i], e[j] = e[j], e[i]
}
func (e EntryList) Less(i, j int) bool {
	return e[i].Key < e[j].Key
}

// Entry type implements os.FileInfo
type Entry struct {
	isDir bool
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

func (e *Entry) Name() string {
	return path.Base(e.Key)
}

func (e *Entry) Size() int64 {
	return int64(len(e.Value))
}

func (e *Entry) Mode() os.FileMode {
	return 0777
}

func (e *Entry) ModTime() time.Time {
	return time.Now() // TODO: support actual modtime
}

func (e *Entry) IsDir() bool {
	return e.isDir
}

func (e *Entry) Sys() interface{} {
	return nil
}

type Manifest struct {
	Revision string   `json:"revision"`
	Roots    []string `json:"roots"`
}

type Store interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
	Bundle(ctx context.Context) ([]byte, error)
}

type Err struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Detail string `json:"detail"`
}

func (e *Err) Error() string {
	return fmt.Sprintf(`{"code":"%d","status":%q,"detail":%q}`, e.Code, e.Status, e.Detail)
}

func NotFoundError(format string, args ...interface{}) error {
	return &Err{
		Code:   http.StatusNotFound,
		Status: http.StatusText(http.StatusNotFound),
		Detail: fmt.Sprintf(format, args...),
	}
}

func ParseError(err error) (e *Err, parseErr error) {
	e = &Err{}
	if parseErr := json.Unmarshal([]byte(err.Error()), e); parseErr != nil {
		return nil, parseErr
	}

	return e, nil
}

func IsErrorCode(err error, code int) bool {
	e, parseErr := ParseError(err)
	if parseErr != nil {
		return false
	}

	return e.Code == code
}

func NormalizePath(p string) string {
	p = path.Clean(p)
	p = strings.TrimLeft(p, "/")
	p = strings.TrimRight(p, "/")
	return p
}

func Archive(ctx context.Context, entries EntryList) ([]byte, error) {
	files := append(EntryList{}, entries...)
	sort.Sort(files) // sort because we want a consistent archive to generate an etag

	b := []byte{}
	buf := bytes.NewBuffer(b)
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	for _, entry := range files {
		file := strings.ReplaceAll(path.Clean("/"+entry.Key), "/", string(filepath.Separator))
		header, err := tar.FileInfoHeader(entry, file)
		if err != nil {
			return nil, err
		}

		header.Name = file
		if err := tw.WriteHeader(header); err != nil {
			return nil, err
		}

		if _, err := tw.Write(entry.Value); err != nil {
			return nil, err
		}
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}

	if err := zr.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func Bundle(ctx context.Context, loader bundle.DirectoryLoader) ([]byte, error) {
	reader := bundle.NewCustomReader(loader)

	b, err := reader.Read()
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer([]byte{})
	compiler := compile.New().
		WithCapabilities(ast.CapabilitiesForThisVersion()).
		WithOutput(buf).
		WithBundle(&b)

	if err := compiler.Build(ctx); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
