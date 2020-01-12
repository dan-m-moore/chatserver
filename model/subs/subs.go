// Package subs provides an asynchronous subscription notification engine.  It
// allows clients to connect (subscribe) and will call Client interface functions
// when pieces of state are changed (via subscription actions).
package subs

import (
	"errors"
	"sync"
)

// Client provides an interface for subscription engine clients to fulfill in order
// to receive asynchronous subscription notifications.
type Client interface {
	OnUsersChanged()
	OnUserChanged(username string)
	OnChannelsChanged()
	OnChannelChanged(channelname string)
}

type clientInfo struct {
	client Client
}

// Engine provides the subscription engine functionality.  It contains information about
// clients that are connected.
type Engine struct {
	mutex   sync.Mutex
	clients map[Client]*clientInfo
}

// NewEngine creates/initializes/returns a new Engine.
func NewEngine() *Engine {
	engine := Engine{
		clients: make(map[Client]*clientInfo),
	}

	return &engine
}

// Connect allows a Client to subscribe to notifications.
func (e *Engine) Connect(client Client) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Make sure the client doesn't already exist
	if _, ok := e.clients[client]; ok {
		return errors.New("Client already exists")
	}

	// Create a new client
	newClient := clientInfo{
		client: client,
	}

	// Add the client to the list
	e.clients[client] = &newClient

	return nil
}

// Disconnect allows a Client to unsubscribe from notifications.
func (e *Engine) Disconnect(client Client) error {
	e.mutex.Lock()
	defer e.mutex.Unlock()

	// Make sure the client exists
	if _, ok := e.clients[client]; !ok {
		return errors.New("Client doesn't exist")
	}

	// Delete the client from the list
	delete(e.clients, client)

	return nil
}

// UsersChanged will notify subscribers (asynchronously) that the users have changed.
func (e *Engine) UsersChanged() {
	go func() {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		for client := range e.clients {
			client.OnUsersChanged()
		}
	}()
}

// UserChanged will notify subscribers (asynchronously) that a user has changed.
func (e *Engine) UserChanged(username string) {
	go func() {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		for client := range e.clients {
			client.OnUserChanged(username)
		}
	}()
}

// ChannelsChanged will notify subscribers (asynchronously) that the channels have changed.
func (e *Engine) ChannelsChanged() {
	go func() {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		for client := range e.clients {
			client.OnChannelsChanged()
		}
	}()
}

// ChannelChanged will notify subscribers (asynchronously) that a channel has changed.
func (e *Engine) ChannelChanged(channelname string) {
	go func() {
		e.mutex.Lock()
		defer e.mutex.Unlock()

		for client := range e.clients {
			client.OnChannelChanged(channelname)
		}
	}()
}
