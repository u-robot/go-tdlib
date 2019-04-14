package client

import (
	"errors"
	"sync"
	"time"
)

// Client is a main object of the package which allow set up and manage connection to Telegram API.
type Client struct {
	tdClient       *TDClient
	extraGenerator ExtraGenerator
	catcher        chan *Response
	listenerStore  *listenerStore
	catchersStore  *sync.Map
	updatesTimeout time.Duration
	catchTimeout   time.Duration
}

// Option is a function type which adjusts client's configuration.
type Option func(*Client)

// WithExtraGenerator configures the client to use existing extra generator.
func WithExtraGenerator(extraGenerator ExtraGenerator) Option {
	return func(client *Client) {
		client.extraGenerator = extraGenerator
	}
}

// WithCatchTimeout configures the client to use specified timeout for catch loop.
func WithCatchTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		client.catchTimeout = timeout
	}
}

// WithUpdatesTimeout configures the client to use specified timeout for listen updates loop.
func WithUpdatesTimeout(timeout time.Duration) Option {
	return func(client *Client) {
		client.updatesTimeout = timeout
	}
}

// WithProxy configures the client to use specified proxy settings.
func WithProxy(request *AddProxyRequest) Option {
	return func(client *Client) {
		client.AddProxy(request)
	}
}

// NewClient creates new client.
func NewClient(authorizationStateHandler AuthorizationStateHandler, options ...Option) (*Client, error) {

	client := &Client{
		tdClient:       NewTDClient(),
		extraGenerator: UUIDV4Generator(),
		listenerStore:  newListenerStore(),
		catchersStore:  &sync.Map{},
		catcher:        make(chan *Response, 1024),
		catchTimeout:   60 * time.Second,
		updatesTimeout: 60 * time.Second,
	}

	for _, option := range options {
		option(client)
	}

	go client.receive()
	go client.catch(client.catcher)

	err := Authorize(client, authorizationStateHandler)
	if err != nil {
		client.ForceStopAndDestroy()
		return nil, err
	}

	return client, nil
}

// GetListener creates and returns update events listener.
func (client *Client) GetListener() *Listener {
	listener := &Listener{
		isActive: true,
		Updates:  make(chan Type, 1024),
	}
	client.listenerStore.Add(listener)

	return listener
}

// Send sends request to TDLib client and waits response.
func (client *Client) Send(request Request) (*Response, error) {
	request.Extra = client.extraGenerator()

	catcher := make(chan *Response, 1)

	client.catchersStore.Store(request.Extra, catcher)

	defer func() {
		close(catcher)
		client.catchersStore.Delete(request.Extra)
	}()

	client.tdClient.Send(request)

	select {
	case response := <-catcher:
		return response, nil

	case <-time.After(client.catchTimeout):
		return nil, errors.New("response catching timeout")
	}
}

// Stop safely closes TDLib client and destroy all unnecessary resources.
func (client *Client) Stop() {
	client.Close()
	client.tdClient.Destroy()
}

// ForceStopAndDestroy closes TDLib client forcefully and unsafe and destroy all unnecessary resources.
func (client *Client) ForceStopAndDestroy() {
	client.Destroy()
	client.tdClient.Destroy()
}

// Lock locks client's mutex.
func (client *Client) Lock(str string) {
	client.tdClient.Lock(str)
}

// Unlock unlocks client's mutex.
func (client *Client) Unlock(str string) {
	client.tdClient.Unlock(str)
}

func (client *Client) receive() {
	for {
		response, err := client.tdClient.Receive(client.updatesTimeout)
		if err != nil {
			continue
		}
		client.catcher <- response

		typ, err := UnmarshalType(response.Data)
		if err != nil {
			continue
		}

		needGc := false
		listeners := client.listenerStore.Listeners()
		for _, listener := range listeners {
			if listener.IsActive() {
				listener.Updates <- typ
			} else {
				needGc = true
			}
		}
		if needGc {
			client.listenerStore.gc()
		}
	}
}

func (client *Client) catch(updates chan *Response) {
	for update := range updates {
		if update.Extra != "" {
			value, ok := client.catchersStore.Load(update.Extra)
			if ok {
				value.(chan *Response) <- update
			}
		}
	}
}
