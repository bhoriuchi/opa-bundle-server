package server

import (
	_ "github.com/bhoriuchi/opa-bundle-server/plugins/publisher/consul"
	_ "github.com/bhoriuchi/opa-bundle-server/plugins/store/consul"
	_ "github.com/bhoriuchi/opa-bundle-server/plugins/store/directory"
	_ "github.com/bhoriuchi/opa-bundle-server/plugins/store/git"
	_ "github.com/bhoriuchi/opa-bundle-server/plugins/subscriber/consul"
	_ "github.com/bhoriuchi/opa-bundle-server/plugins/webhook/gogs"
)
