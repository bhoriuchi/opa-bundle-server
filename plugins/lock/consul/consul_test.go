package consul_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	client "github.com/bhoriuchi/opa-bundle-server/core/clients/consul"
	"github.com/bhoriuchi/opa-bundle-server/plugins/lock"
	"github.com/bhoriuchi/opa-bundle-server/plugins/lock/consul"
	"github.com/open-policy-agent/opa/logging"
)

func TestLock(t *testing.T) {
	addr := os.Getenv("CONSUL_ADDR")
	if addr == "" {
		t.Log("no CONSUL_ADDR environment variable specified, skipping test")
		return
	}

	logger := logging.NewStandardLogger()
	logger.SetLevel(logging.Debug)

	opts := &lock.Options{
		Config: consul.Config{
			Key: "test-lock",
			Consul: &client.Config{
				Address: addr,
			},
		},
		Logger: logger,
	}

	var wg sync.WaitGroup
	wg.Add(2)

	// run a test
	runner := func(to time.Duration) {
		l, err := consul.NewLock(opts)
		if err != nil {
			t.Errorf("failed to create new consul lock: %s", err)
			return
		}

		if err := l.Connect(context.Background()); err != nil {
			t.Errorf("connect error %s", err)
			return
		}

		// cancel the lock operations
		timer := time.AfterFunc(to, func() {
			t.Log("calling unlock")
			if err := l.Unlock(context.Background()); err != nil {
				t.Errorf("unlock error: %s", err)
			}
			l.Disconnect(context.Background())
			wg.Done()
		})
		defer timer.Stop()

		if err := lock.Acquire(context.Background(), l); err != nil {
			t.Errorf("acquire error: %s", err)
			return
		}
	}

	go runner(5 * time.Second)
	time.Sleep(500 * time.Millisecond)
	go runner(10 * time.Second)

	wg.Wait()
}
