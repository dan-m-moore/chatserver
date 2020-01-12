// Package webconn manages state associated with a single web view connection.  As most of the
// web view connection state is held in the web client, this only handles forwarding model
// subscription updates to the open websocket.
package webconn

import (
	"golang.org/x/net/websocket"
)

// WebConn manages data associated with a single web client connection (over websocket).
type WebConn struct {
	ws *websocket.Conn
}

// NewWebConn creates/initializes/returns a new WebConn.
func NewWebConn(ws *websocket.Conn) *WebConn {
	webConn := WebConn{
		ws: ws,
	}

	return &webConn
}

// OnUsersChanged is called whenever the users state changes in the model.  It will forward this
// update to the websocket.
func (w *WebConn) OnUsersChanged() {
	msg := "{\"id\":-1,\"result\":{\"method\":\"OnUsersChanged\"},\"error\":null}"
	_, err := w.ws.Write([]byte(msg))
	if err != nil {
		// Assume this error means the client went away and will be cleaned up eventually
		return
	}
}

// OnUserChanged is called whenever a particular user's state changes in the model.  It will forward
// this update to the websocket.
func (w *WebConn) OnUserChanged(username string) {
	msg := "{\"id\":-1,\"result\":{\"method\":\"OnUserChanged\",\"username\":\"" + username + "\"},\"error\":null}"
	_, err := w.ws.Write([]byte(msg))
	if err != nil {
		// Assume this error means the client went away and will be cleaned up eventually
		return
	}
}

// OnChannelsChanged is called whenever the channels state changes in the model.  It will forward
// this update to the websocket.
func (w *WebConn) OnChannelsChanged() {
	msg := "{\"id\":-1,\"result\":{\"method\":\"OnChannelsChanged\"},\"error\":null}"
	_, err := w.ws.Write([]byte(msg))
	if err != nil {
		// Assume this error means the client went away and will be cleaned up eventually
		return
	}
}

// OnChannelChanged is called whenever a particular channel's state changes in the model.  It will
// forward this update to the websocket.
func (w *WebConn) OnChannelChanged(channelname string) {
	msg := "{\"id\":-1,\"result\":{\"method\":\"OnChannelChanged\",\"channelname\":\"" + channelname + "\"},\"error\":null}"
	_, err := w.ws.Write([]byte(msg))
	if err != nil {
		// Assume this error means the client went away and will be cleaned up eventually
		return
	}
}
