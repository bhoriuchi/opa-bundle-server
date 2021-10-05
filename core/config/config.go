package config

import (
	"bytes"
	"os"
	"strings"
	"text/template"

	"github.com/bhoriuchi/opa-bundle-server/core/utils"
)

type TemplateData struct {
	Meta map[string]interface{} `json:"meta" yaml:"meta"`
	Env  map[string]string      `json:"env" yaml:"env"`
}

type Config struct {
	Server      *Server                `json:"server" yaml:"server"`
	Lock        *Lock                  `json:"lock" yaml:"lock"`
	Stores      map[string]*Store      `json:"stores" yaml:"stores"`
	Deployers   map[string]*Deployer   `json:"deployers" yaml:"deployers"`
	Webhooks    map[string]*Webhook    `json:"webhooks" yaml:"webhooks"`
	Subscribers map[string]*Subscriber `json:"subscribers" yaml:"subscribers"`
	Publishers  map[string]*Publisher  `json:"publishers" yaml:"publishers"`
	Bundles     map[string]*Bundle     `json:"bundles" yaml:"bundles"`
}

type Server struct {
	Address string `json:"address" yaml:"address"`
}

type Lock struct {
	Type   string      `json:"type" yaml:"type"`
	Config interface{} `json:"config" yaml:"config"`
}

type Deployer struct {
	Type   string      `json:"type" yaml:"type"`
	Config interface{} `json:"config" yaml:"config"`
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
	Store       string   `json:"store" yaml:"store"`
	Webhooks    []string `json:"webhooks" yaml:"webhooks"`
	Publishers  []string `json:"publishers" yaml:"publishers"`
	Subscribers []string `json:"subscribers" yaml:"subscribers"`
	Deployers   []string `json:"deployers" yaml:"deployers"`
	Polling     Polling  `json:"polling" yaml:"polling"`
}

type Polling struct {
	Disable         bool  `json:"disable" yaml:"disable"`
	MinDelaySeconds int64 `json:"min_delay_seconds" yaml:"min_delay_seconds"`
	MaxDelaySeconds int64 `json:"max_delay_seconds" yaml:"max_delay_seconds"`
}

// NewConfig creates a new config from config content
func NewConfig(content []byte) (*Config, error) {
	data := &TemplateData{}
	content = bytes.TrimSpace(content)

	if err := utils.Unmarshal(content, data); err != nil {
		return nil, err
	}

	data.Env = map[string]string{}
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			data.Env[parts[0]] = parts[1]
		}
	}

	tmpl, err := template.New("config").Parse(string(content))
	if err != nil {
		return nil, err
	}

	buf := bytes.NewBuffer([]byte{})
	if err := tmpl.Execute(buf, data); err != nil {
		return nil, err
	}

	c := &Config{}
	if err := utils.Unmarshal(buf.Bytes(), c); err != nil {
		return nil, err
	}

	return c, nil
}
