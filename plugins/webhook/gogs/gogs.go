package gogs

import (
	"fmt"
	"net/http"

	"github.com/bhoriuchi/opa-bundle-server/core/logger"
	"github.com/bhoriuchi/opa-bundle-server/core/utils"
	"github.com/bhoriuchi/opa-bundle-server/plugins/webhook"
	gogsh "github.com/go-playground/webhooks/v6/gogs"
)

const (
	ProviderName = "gogs"
)

func init() {
	webhook.Providers[ProviderName] = NewWebhook
}

type Webhook struct {
	name   string
	events []gogsh.Event
	logger logger.Logger
	cb     func()
	Config *Config
	hook   *gogsh.Webhook
}

type Config struct {
	Secret string   `json:"secret"`
	Events []string `json:"events"`
}

func NewWebhook(opts webhook.Options) (webhook.Webhook, error) {
	h := &Webhook{
		name:   opts.Name,
		logger: opts.Logger,
		cb:     opts.Callback,
		events: []gogsh.Event{},
		Config: &Config{},
	}

	err := utils.ReMarshal(opts.Config, h.Config)
	if err != nil {
		return nil, err
	}

	if len(h.Config.Events) == 0 {
		return nil, fmt.Errorf("at least one event is required for webhook %s", opts.Name)
	}

	// convert events to gogs events
	for _, event := range h.Config.Events {
		e, err := parseEvent(event)
		if err != nil {
			return nil, fmt.Errorf("invalid event %s for webhook %s", event, opts.Name)
		}
		h.events = append(h.events, e)
	}

	hookOpts := []gogsh.Option{}
	if h.Config.Secret != "" {
		hookOpts = append(hookOpts, gogsh.Options.Secret(h.Config.Secret))
	}

	if h.hook, err = gogsh.New(hookOpts...); err != nil {
		return nil, err
	}

	return h, nil
}

func (h *Webhook) Handle(w http.ResponseWriter, r *http.Request) {
	if _, err := h.hook.Parse(r, h.events...); err != nil {
		h.logger.Errorf("error parsing webhook %s: %s", h.name, err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(err.Error()))
		return
	}

	go h.cb()
	w.WriteHeader(http.StatusOK)
}

func parseEvent(event string) (gogsh.Event, error) {
	switch e := gogsh.Event(event); e {
	case gogsh.CreateEvent,
		gogsh.DeleteEvent,
		gogsh.ForkEvent,
		gogsh.IssueCommentEvent,
		gogsh.IssuesEvent,
		gogsh.PullRequestEvent,
		gogsh.PushEvent,
		gogsh.ReleaseEvent:
		return e, nil
	default:
		return gogsh.Event(""), fmt.Errorf("invalid event")
	}
}
