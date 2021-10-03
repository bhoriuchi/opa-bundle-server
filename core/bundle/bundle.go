package bundle

import (
	"context"
	"crypto/md5"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/bhoriuchi/opa-bundle-server/core/config"
	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/publisher"
	"github.com/bhoriuchi/opa-bundle-server/plugins/remote"
	"github.com/bhoriuchi/opa-bundle-server/plugins/store"
	"github.com/oleiade/lane"
	"github.com/open-policy-agent/opa/util"
)

const (
	DefaultPollingMinDelaySeconds int64 = 60
	DefaultPollingMaxDelaySeconds int64 = 120
	minRetryDelay                       = time.Millisecond * 100
)

type Bundle struct {
	mx         sync.Mutex
	dq         *lane.Deque
	Name       string
	Logger     logger.Logger
	Store      store.Store
	Webhook    string
	Subscriber string
	Publisher  publisher.Publisher
	Remotes    []remote.Remote
	Config     *config.Bundle
	data       []byte
	etag       string
	activated  bool
	pollCancel context.CancelFunc
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
	return utils.Enqueue(b.dq, "", func(id interface{}, args ...interface{}) error {
		b.mx.Lock()
		defer b.mx.Unlock()

		var err error
		b.Logger.Debugf("request %s rebuilding bundle %s", id, b.Name)

		if b.data, err = b.Store.Bundle(ctx); err != nil {
			return err
		}

		b.etag = fmt.Sprintf("%x", md5.Sum(b.data))
		return nil
	})
}

func (b *Bundle) Activate() error {
	var ctx context.Context

	if b.activated {
		return fmt.Errorf("bundle %s already activated", b.Name)
	}

	b.dq = lane.NewCappedDeque(1)

	ctx, b.pollCancel = context.WithCancel(context.Background())
	go b.loop(ctx)

	b.activated = true
	return nil
}

func (b *Bundle) Deactivate() error {
	if b.pollCancel != nil {
		b.pollCancel()
	}

	if !b.activated {
		return fmt.Errorf("bundle %s is not activated", b.Name)
	}
	return nil
}

func (b *Bundle) loop(ctx context.Context) {
	var retry int

	for {
		var delay time.Duration

		if ctx.Err() != nil {
			return
		}

		err := b.Rebuild(ctx)

		// if polling is disabled, dont try to rebuild
		if b.Config.Polling.Disable {
			return
		}

		minDelay := b.Config.Polling.MinDelaySeconds
		if minDelay == 0 {
			minDelay = DefaultPollingMinDelaySeconds
		}

		maxDelay := b.Config.Polling.MaxDelaySeconds
		if maxDelay == 0 {
			maxDelay = DefaultPollingMaxDelaySeconds
		}

		if err != nil {
			delay = util.DefaultBackoff(float64(minRetryDelay), float64(maxDelay), retry) * time.Second
		} else {
			min := float64(minDelay)
			max := float64(maxDelay)
			delay = time.Duration(((max-min)*rand.Float64())+min) * time.Second
		}

		b.Logger.Debugf("Waiting %v before next rebuild.", delay)

		select {
		case <-time.After(delay):
			if err != nil {
				retry++
			} else {
				retry = 0
			}
		case <-ctx.Done():
			return
		}
	}
}
