package config

import (
	"bytes"
	"encoding/json"

	"github.com/ghodss/yaml"
)

type Config struct {
	Server      *Server                `json:"server" yaml:"server"`
	Stores      map[string]*Store      `json:"stores" yaml:"stores"`
	Webhooks    map[string]*Webhook    `json:"webhooks" yaml:"webhooks"`
	Subscribers map[string]*Subscriber `json:"subscribers" yaml:"subscribers"`
	Publishers  map[string]*Publisher  `json:"publishers" yaml:"publishers"`
	Bundles     map[string]*Bundle     `json:"bundles" yaml:"bundles"`
}

type Server struct {
	Address string `json:"address" yaml:"address"`
}

type Subscriber struct {
	Type   string      `json:"type" yaml:"type"`
	Config interface{} `json:"config" yaml:"config"`
}

type Publisher struct {
	Type   string      `json:"type" yaml:"type"`
	Config interface{} `json:"config" yaml:"config"`
}

type Webhook struct {
	Type   string      `json:"type" yaml:"type"`
	Config interface{} `json:"config" yaml:"config"`
}

type Store struct {
	Type   string      `json:"type" yaml:"type"`
	Config interface{} `json:"config" yaml:"config"`
}

type Bundle struct {
	Store      string   `json:"store" yaml:"store"`
	Webhook    string   `json:"webhook" yaml:"webhook"`
	Publisher  string   `json:"publisher" yaml:"publisher"`
	Subscriber string   `json:"subscriber" yaml:"subscriber"`
	Remotes    []string `json:"remotes" yaml:"remotes"`
	Polling    Polling  `json:"polling" yaml:"polling"`
}

type Polling struct {
	Disable         bool  `json:"disable" yaml:"disable"`
	MinDelaySeconds int64 `json:"min_delay_seconds" yaml:"min_delay_seconds"`
	MaxDelaySeconds int64 `json:"max_delay_seconds" yaml:"max_delay_seconds"`
}

// NewConfig creates a new config from config content
func NewConfig(content []byte) (*Config, error) {
	c := &Config{}
	content = bytes.TrimSpace(content)
	if bytes.HasPrefix(content, []byte("{")) && bytes.HasSuffix(content, []byte("}")) {
		if err := json.Unmarshal(content, c); err != nil {
			return nil, err
		}
		return c, nil
	}

	if err := yaml.Unmarshal(content, c); err != nil {
		return nil, err
	}

	return c, nil
}
