package consul

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/bhoriuchi/opa-bundle-server/core/clients/consul"
	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/lock"
	"github.com/google/uuid"
	"github.com/hashicorp/consul/api"
)

const (
	ProviderName = "consul"
)

func init() {
	lock.Providers[ProviderName] = NewLock
}

type Lock struct {
	id      string
	mx      sync.Mutex
	cc      chan struct{}
	stop    chan struct{}
	hasLock bool
	client  *consul.Client
	config  *Config
	logger  logger.Logger
	lock    *api.Lock
	wait    time.Duration
	ttl     string
}

type Config struct {
	Key    string         `json:"key" yaml:"key"`
	Consul *consul.Config `json:"consul" yaml:"consul"`
}

// NewLock creates a new lock maanger
func NewLock(opts *lock.Options) (lock.Lock, error) {
	l := &Lock{
		id:     uuid.NewString(),
		cc:     make(chan struct{}),
		config: &Config{},
		logger: opts.Logger,
		wait:   api.DefaultLockWaitTime,
		ttl:    api.DefaultLockSessionTTL,
	}

	if opts.Config == nil {
		return nil, fmt.Errorf("node %s invalid configuration for consul lock", l.id)
	}

	if err := utils.ReMarshal(opts.Config, l.config); err != nil {
		return nil, err
	}

	if l.config.Consul == nil {
		return nil, fmt.Errorf("node %s no consul configuration provided for consul lock", l.id)
	}

	return l, nil
}

// HasLock lock is held by this node
func (l *Lock) HasLock() bool {
	l.mx.Lock()
	defer l.mx.Unlock()
	return l.hasLock
}

// Connect connects to consul
func (l *Lock) Connect(ctx context.Context) (err error) {
	l.logger.Debug("node %s connecting to consul lock %s", l.id, l.config.Consul.Address)
	if l.client != nil {
		return fmt.Errorf("node %s already connected", l.id)
	}

	l.client, err = consul.NewClient(l.config.Consul)
	if err != nil {
		return
	}

	return
}

// Disconnect disconnects from consul
func (l *Lock) Disconnect(ctx context.Context) (err error) {
	if l.client == nil {
		err = fmt.Errorf("node %s not connected", l.id)
		return
	}

	l.client = nil
	return
}

// setHasLock sets the has lock property
func (l *Lock) setHasLock(hasLock bool) {
	l.mx.Lock()
	prev := l.hasLock
	l.hasLock = hasLock
	l.mx.Unlock()

	if hasLock {
		l.logger.Debug("node %s acquired lock", l.id)
	} else if prev {
		l.logger.Debug("node %s lock lost", l.id)
	} else {
		l.logger.Debug("node %s failed to acquire lock", l.id)
	}
}

// Lock creates a new lock
func (l *Lock) Lock(ctx context.Context) (err error) {
	var lc <-chan struct{}

	if l.lock, err = l.client.Consul().LockOpts(&api.LockOptions{
		Key:          l.config.Key,
		LockWaitTime: l.wait,
		SessionTTL:   l.ttl,
	}); err != nil {
		return
	}

	l.stop = make(chan struct{})

	if l.lock == nil {
		l.setHasLock(false)
		err = fmt.Errorf("node %s no lock manager", l.id)
		return
	}

	if lc, err = l.lock.Lock(l.stop); err != nil {
		l.setHasLock(false)
		err = lock.ErrLockFailed
		return
	}

	l.stop = nil
	l.setHasLock(true)

	select {
	case <-l.cc:

		l.setHasLock(false)
		err = lock.ErrLockClosed
		return
	case <-lc:
		l.setHasLock(false)
		err = lock.ErrLockFailed
		return
	}
}

// Unlock unlocks a lock by id
func (l *Lock) Unlock(ctx context.Context) (err error) {
	l.logger.Debug("node %s unlocking consul lock", l.id)

	if l.stop != nil {
		l.logger.Debug("node %s sending message to lock abort channel", l.id)
		go func() { l.stop <- struct{}{} }()
	}
	if l.cc != nil {
		l.logger.Debug("node %s sending message to close channel", l.id)
		go func() { l.cc <- struct{}{} }()
	}

	if err = l.lock.Unlock(); err != nil {
		if err == api.ErrLockNotHeld {
			err = nil
			return
		}
		l.logger.Error("node %s failed to unlock consul lock: %s", l.id, err)
		return
	}

	l.logger.Debug("node %s unlocked consul lock", l.id)
	return
}
