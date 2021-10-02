package bundle

import (
	"context"
	"crypto/md5"
	"fmt"
	"sync"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/plugins/publisher"
	"github.com/bhoriuchi/opa-bundle-server/plugins/remote"
	"github.com/bhoriuchi/opa-bundle-server/plugins/store"
	"github.com/bhoriuchi/opa-bundle-server/plugins/subscriber"
)

type Bundle struct {
	mx         sync.Mutex
	Name       string
	Logger     logger.Logger
	Store      store.Store
	Webhook    string
	Subscriber subscriber.Subscriber
	Publisher  publisher.Publisher
	Remotes    []remote.Remote
	data       []byte
	etag       string
}

func (b *Bundle) Data() []byte {
	b.mx.Lock()
	defer b.mx.Unlock()

	return b.data
}

func (b *Bundle) Etag() string {
	b.mx.Lock()
	defer b.mx.Unlock()

	return b.etag
}

func (b *Bundle) Rebuild(ctx context.Context) error {
	b.mx.Lock()
	defer b.mx.Unlock()

	var err error
	b.Logger.Debugf("rebuilding bundle %s", b.Name)

	if b.data, err = b.Store.Bundle(ctx); err != nil {
		return err
	}

	b.etag = fmt.Sprintf("%x", md5.Sum(b.data))
	return nil
}

func (b *Bundle) Activate() error {
	return nil
}

func (b *Bundle) Deactivate() error {
	return nil
}
