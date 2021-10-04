# opa-bundle-server

A highly scalable bundle server with pluggable storage, webhooks, deployers, publishers, and subscribers.

OPA Bundle Server provides the ability to store bundle data in various backends like git, consul, postgres, etc. and set up triggers to rebuild and deploy bundles. Additionally the bundle server serves the bundle over a REST endpoint.

## Components

### Store

Stores provide a way to generate a bundle from any system that can store data. Stores can be anything from a git repository to a database. The `Bundle` method generates an OPA bundle. By default bundles are updated on a periodic basis. Stores are linked to bundle objects

```go
// Store interface
type Store interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
	Bundle(ctx context.Context) ([]byte, error)
}
```

### Subscriber

Subscribers provide a way to subscribe to changes on a store. When a message is recieved, the subscriber will trigger a rebuild on any bundles linked to it. Subscribers can watch or subscribe to events on event brokers like NATS, Kafka, or even a Consul watch

```go
// Subscriber interface
type Subscriber interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
	Subscribe(ctx context.Context) (err error)
	Unsubscribe(ctx context.Context) (err error)
}
```

### Webhook

Webhooks are similar to subscribers. They provide a way to trigger a bundle rebuild on bundles linked to them. Webhooks are typically used with the git store to signal a rebuild on a push event

```go
// Webhook interface
type Webhook interface {
	Handle(w http.ResponseWriter, r *http.Request)
}
```

### Deployer

Deployers provide a way to distribute bundles to external locations. Deployers can deploy to locations like an ftp server, cloud storage, or http server via ssh

```go
// Deployer Interface
type Deployer interface {
	Deploy(ctx context.Context) (err error)
}
```

### Publisher

Publishers provide a way to signal external systems that a bundle update has occured. This can potentially trigger OPA bundle updates with something like [opa-plugin-subscribe](https://github.com/bhoriuchi/opa-plugin-subscribe). Publishers can publish to event brokers like NATS, Kafka, RabbitMQ, etc.

```go
type Subscriber interface {
	Connect(ctx context.Context) (err error)
	Disconnect(ctx context.Context) (err error)
	Subscribe(ctx context.Context) (err error)
	Unsubscribe(ctx context.Context) (err error)
}
```