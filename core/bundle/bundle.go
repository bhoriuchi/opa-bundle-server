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
	"github.com/bhoriuchi/opa-bundle-server/plugins/deployer"
	"github.com/bhoriuchi/opa-bundle-server/plugins/publisher"
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
	mx          sync.Mutex
	dq          *lane.Deque
	Name        string
	Logger      logger.Logger
	Store       store.Store
	Webhooks    []string
	Subscribers []string
	Publishers  []publisher.Publisher
	Deployers   []deployer.Deployer
	Config      *config.Bundle
	data        []byte
	etag        string
	activated   bool
	pollCancel  context.CancelFunc
}

// Data returns the bundle data
func (b *Bundle) Data() []byte {
	b.mx.Lock()
	defer b.mx.Unlock()

	return b.data
}

// Etag returns the bundle's etag which is the md5 sum of its contents
func (b *Bundle) Etag() string {
	b.mx.Lock()
	defer b.mx.Unlock()

	return b.etag
}

// Rebuild rebuilds the bundle. Becuase requests to this function are made asynchronously
// At most 1 call be will queued up during execution. This ensures that any calls made
// to rebuild during a rebuild operation will still be processed but will be combined into
// a single queued up rebuild instead of n-rebuilds
func (b *Bundle) Rebuild(ctx context.Context) error {
	return utils.Enqueue(b.dq, "", func(id interface{}, args ...interface{}) error {
		b.mx.Lock()
		defer b.mx.Unlock()

		var err error
		b.Logger.Debug("request %s rebuilding bundle %s", id, b.Name)

		// create the bundle
		if b.data, err = b.Store.Bundle(ctx); err != nil {
			return err
		}

		// calculate the etag and determine if it changed
		lastEtag := b.etag
		b.etag = fmt.Sprintf("%x", md5.Sum(b.data))

		// if etag has changed, the bundle was updated
		// if the last etag is empty, this is the first update
		// so ignore publishing updates
		if lastEtag != b.etag && lastEtag != "" {
			// TODO: perform deployments

			// publish events on successful deployments
			payload := []byte(fmt.Sprintf(`{"etag":%q}`, b.etag))
			for _, pub := range b.Publishers {
				go pub.Publish(ctx, payload)
			}
		}

		return nil
	})
}

// Activate sets up the bundle, performs the initial build, and by
// default starts polling the store and rebuilding periodically
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

// Deactivate stops the polling loop and sets the bundle to deactivated
func (b *Bundle) Deactivate() error {
	if b.pollCancel != nil {
		b.pollCancel()
	}

	if !b.activated {
		return fmt.Errorf("bundle %s is not activated", b.Name)
	}

	b.activated = false
	return nil
}

// loop performs a polling operation if it is enabled. loop will always
// run at least once to perform the initial build of the bundle
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

		b.Logger.Debug("Waiting %v before next rebuild.", delay)

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
