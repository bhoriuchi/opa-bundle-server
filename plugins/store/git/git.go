package git

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/store"
	getter "github.com/hashicorp/go-getter/v2"
	"github.com/open-policy-agent/opa/bundle"
)

const (
	ProviderName = "git"
)

func init() {
	store.Providers[ProviderName] = NewStore
}

type Store struct {
	name   string
	config *Config
	logger logger.Logger
}

type Config struct {
	Source  string `json:"source" yaml:"source"`
	TempDir string `json:"temp_dir" yaml:"temp_dir"`
}

// NewStore creates a new store
func NewStore(opts *store.Options) (store.Store, error) {
	s := &Store{
		name:   opts.Name,
		config: &Config{},
		logger: opts.Logger,
	}

	if opts.Config == nil {
		return nil, fmt.Errorf("invalid configuration for store %s", opts.Name)
	}

	if err := utils.ReMarshal(opts.Config, s.config); err != nil {
		return nil, err
	}

	return s, nil
}

// Connect is noop but required to implement store interface
func (s *Store) Connect(ctx context.Context) (err error) {
	s.logger.Debugf("connecting to git store %s", s.name)
	return
}

// Disconnect is noop but required to implement store interface
func (s *Store) Disconnect(ctx context.Context) (err error) {
	return
}

// Bundle
func (s *Store) Bundle(ctx context.Context) ([]byte, error) {
	var err error

	// Get the pwd
	pwd, err := os.Getwd()
	if err != nil {
		s.logger.Errorf("error getting wd: %s", err)
	}

	parentDir := os.TempDir()
	if s.config.TempDir != "" {
		if parentDir, err = filepath.Abs(s.config.TempDir); err != nil {
			return nil, err
		}
	}

	dir, err := ioutil.TempDir(parentDir, "opabs-"+s.name+"-*")
	if err != nil {
		return nil, err
	}

	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			s.logger.Errorf("failed to clean up temp directory %s for git store %s: %s", dir, s.name, err)
		}
	}()

	req := &getter.Request{
		Src:     s.config.Source,
		Dst:     dir,
		Pwd:     pwd,
		GetMode: getter.ModeAny,
	}

	client := getter.DefaultClient
	client.Getters = []getter.Getter{
		new(getter.GitGetter),
	}

	res, err := getter.DefaultClient.Get(ctx, req)
	if err != nil {
		return nil, err
	}

	s.logger.Debugf("successfully cloned %s in store %s", res.Dst, s.name)

	loader := bundle.NewDirectoryLoader(res.Dst)
	return store.Bundle(ctx, loader)
}
