package store

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"strings"
)

var (
	Providers = map[string]NewStoreFunc{}
)

type NewStoreFunc func(config interface{}) (Store, error)

type Entry struct {
	Key   string `json:"key"`
	Value []byte `json:"value"`
}

type Manifest struct {
	Revision string   `json:"revision"`
	Roots    []string `json:"roots"`
}

type Store interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
	Get(ctx context.Context, bundle, key string) (out *Entry, err error)
	Set(ctx context.Context, bundle, key string, entry *Entry) (err error)
	Del(ctx context.Context, bundle, key string) (err error)
	List(ctx context.Context, bundle, prefix string) (entries []*Entry, err error)
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
