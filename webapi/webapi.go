// Package webapi provides a JSON RPC server connection handler that gets called
// when a new JSON RPC (over websocket) session is initiated.  It also provides
// the service API via the WebAPI public interface.
package webapi

import (
	"chatserver/model"
	"chatserver/model/subs"
	"chatserver/webconn"
	"log"
	"net/rpc"
	"net/rpc/jsonrpc"
	"sort"
	"time"

	"golang.org/x/net/websocket"
)

// NewConnectionHandler creates a new websocket Handler that will manage individual
// websocket connections.  It will serve a JSON RPC API on that connection.
func NewConnectionHandler(subsEngine *subs.Engine) websocket.Handler {
	connectionHandler := func(ws *websocket.Conn) {
		webConn := webconn.NewWebConn(ws)

		// Connect the subscriptions for this web conn
		err := subsEngine.Connect(webConn)
		if err != nil {
			log.Fatal(err)
		}

		// For a single connection, handle requests sequentially
		for {
			err := rpc.ServeRequest(jsonrpc.NewServerCodec(ws))
			if err != nil {
				break
			}
		}

		// Disconnect the subscriptions for this web conn
		err = subsEngine.Disconnect(webConn)
		if err != nil {
			log.Fatal(err)
		}
	}
	return connectionHandler
}

// WebAPI provides the JSON RPC service API.
type WebAPI struct {
	model *model.Model
}

// NewInstance creates/initializes/returns a new WebAPI instance.
func NewInstance(model *model.Model) *WebAPI {
	instance := WebAPI{
		model: model,
	}

	return &instance
}

// CreateUserArgs provides the input arguments for the CreateUser action.
type CreateUserArgs struct {
	Username string
}

// CreateUserResponse provides the output arguments for the CreateUser action.
type CreateUserResponse struct {
}

// CreateUser will create a new user.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.CreateUser",
//     "params": [{
//         "Username": "User1"
//     }]
// }
//
// Output
// {
// }
func (w *WebAPI) CreateUser(args *CreateUserArgs, response *CreateUserResponse) error {
	w.model.CreateUser(args.Username)

	return nil
}

// DeleteUserArgs provides the input arguments for the DeleteUser action.
type DeleteUserArgs struct {
	Username string
}

// DeleteUserResponse provides the output arguments for the DeleteUser action.
type DeleteUserResponse struct {
}

// DeleteUser will delete an existing user.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.DeleteUser",
//     "params": [{
//         "Username": "User1"
//     }]
// }
//
// Output
// {
// }
func (w *WebAPI) DeleteUser(args *DeleteUserArgs, response *DeleteUserResponse) error {
	w.model.DeleteUser(args.Username)

	return nil
}

// GetUserInfoArgs provides the input arguments for the GetUserInfo action.
type GetUserInfoArgs struct {
	Username string
}

// GetUserInfoResponse provides the output arguments for the GetUserInfo action.
type GetUserInfoResponse struct {
	User model.User
}

// GetUserInfo will get user info for a specified user.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.GetUserInfo",
//     "params": [{
//         "Username": "User1"
//     }]
// }
//
// Output
// {
//     "User": {
//         "Name": "User1",
//         "BlockedUsers": [
//             "User2",
//             "User3"
//         ]
//     }
// }
func (w *WebAPI) GetUserInfo(args *GetUserInfoArgs, response *GetUserInfoResponse) error {
	userInfo := w.model.GetUserInfo(args.Username)
	response.User = userInfo
	sort.Strings(response.User.BlockedUsers)

	return nil
}

// GetUsersArgs provides the input arguments for the GetUsers action.
type GetUsersArgs struct {
}

// GetUsersResponse provides the output arguments for the GetUsers action.
type GetUsersResponse struct {
	Users []string
}

// GetUsers will get a list of all users.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.GetUsers",
//     "params": [{
//     }]
// }
//
// Output
// {
//     "Users": [
//         "User1",
//         "User2"
//     ]
// }
func (w *WebAPI) GetUsers(args *GetUsersArgs, response *GetUsersResponse) error {
	users := w.model.GetUsers()

	// Sort the users alphabetically
	response.Users = make([]string, 0)
	for user := range users {
		response.Users = append(response.Users, user)
	}
	sort.Strings(response.Users)

	return nil
}

// BlockUserArgs provides the input arguments for the BlockUser action.
type BlockUserArgs struct {
	Username        string
	UsernameToBlock string
}

// BlockUserResponse provides the output arguments for the BlockUser action.
type BlockUserResponse struct {
}

// BlockUser will block an existing user for the given user.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.BlockUser",
//     "params": [{
//         "Username": "User1",
//         "UsernameToBlock": "User2"
//     }]
// }
//
// Output
// {
// }
func (w *WebAPI) BlockUser(args *BlockUserArgs, response *BlockUserResponse) error {
	w.model.BlockUser(args.Username, args.UsernameToBlock)

	return nil
}

// UnblockUserArgs provides the input arguments for the UnblockUser action.
type UnblockUserArgs struct {
	Username          string
	UsernameToUnblock string
}

// UnblockUserResponse provides the output arguments for the UnblockUser action.
type UnblockUserResponse struct {
}

// UnblockUser will unblock an existing user for the given user.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.UnblockUser",
//     "params": [{
//         "Username": "User1",
//         "UsernameToUnblock": "User2"
//     }]
// }
//
// Output
// {
// }
func (w *WebAPI) UnblockUser(args *UnblockUserArgs, response *UnblockUserResponse) error {
	w.model.UnblockUser(args.Username, args.UsernameToUnblock)

	return nil
}

// CreateChannelArgs provides the input arguments for the CreateChannel action.
type CreateChannelArgs struct {
	Channelname string
}

// CreateChannelResponse provides the output arguments for the CreateChannel action.
type CreateChannelResponse struct {
}

// CreateChannel will create a new channel.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.CreateChannel",
//     "params": [{
//         "Channelname": "Channel1"
//     }]
// }
//
// Output
// {
// }
func (w *WebAPI) CreateChannel(args *CreateChannelArgs, response *CreateChannelResponse) error {
	w.model.CreateChannel(args.Channelname)

	return nil
}

// DeleteChannelArgs provides the input arguments for the DeleteChannel action.
type DeleteChannelArgs struct {
	Channelname string
}

// DeleteChannelResponse provides the output arguments for the DeleteChannel action.
type DeleteChannelResponse struct {
}

// DeleteChannel will delete an existing channel.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.DeleteChannel",
//     "params": [{
//         "Channelname": "Channel1"
//     }]
// }
//
// Output
// {
// }
func (w *WebAPI) DeleteChannel(args *DeleteChannelArgs, response *DeleteChannelResponse) error {
	w.model.DeleteChannel(args.Channelname)

	return nil
}

// GetChannelHistoryArgs provides the input arguments for the GetChannelHistory action.
type GetChannelHistoryArgs struct {
	Channelname string
	Username    string
	NumMessages int
}

// ChannelHistoryMessage provides a translation of the model.Message struct
type ChannelHistoryMessage struct {
	Username  string
	Timestamp string
	Text      string
}

// GetChannelHistoryResponse provides the output arguments for the GetChannelHistory action.
type GetChannelHistoryResponse struct {
	Messages []ChannelHistoryMessage
}

// GetChannelHistory will get channel history for a channel (filtered for a user) up to a number of messages.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.GetChannelHistory",
//     "params": [{
//         "Channelname": "Channel1",
//         "Username": "User1",
//         "NumMessages": 12
//     }]
// }
//
// Output
// {
//     "Messages": [{
//         "Username": "User1",
//         "Timestamp": "2020-01-12...",
//         "Text": "Message1"
//     }]
// }
func (w *WebAPI) GetChannelHistory(args *GetChannelHistoryArgs, response *GetChannelHistoryResponse) error {
	messages := w.model.GetChannelHistory(args.Channelname, args.Username, args.NumMessages)
	response.Messages = make([]ChannelHistoryMessage, len(messages))
	for i, message := range messages {
		response.Messages[i].Username = message.Username
		response.Messages[i].Timestamp = message.Timestamp.Format("2006-01-02 15:04:05")
		response.Messages[i].Text = message.Text
	}

	return nil
}

// GetChannelInfoArgs provides the input arguments for the GetChannelInfo action.
type GetChannelInfoArgs struct {
	Channelname string
}

// GetChannelInfoResponse provides the output arguments for the GetChannelInfo action.
type GetChannelInfoResponse struct {
	Channel model.ChannelInfo
}

// GetChannelInfo will get channel info for a specified channel.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.GetChannelInfo",
//     "params": [{
//         "Channelname": "Channel1"
//     }]
// }
//
// Output
// {
//     "Channel": {
//         "Name": "Channel1",
//         "NumMessages": 12
//     }
// }
func (w *WebAPI) GetChannelInfo(args *GetChannelInfoArgs, response *GetChannelInfoResponse) error {
	channelInfo := w.model.GetChannelInfo(args.Channelname)
	response.Channel = channelInfo

	return nil
}

// GetChannelsArgs provides the input arguments for the GetChannels action.
type GetChannelsArgs struct {
}

// GetChannelsResponse provides the output arguments for the GetChannels action.
type GetChannelsResponse struct {
	Channels []string
}

// GetChannels will get a list of all channels.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.GetChannels",
//     "params": [{
//     }]
// }
//
// Output
// {
//     "Channels": [
//         "Channel1",
//         "Channel2"
//     ]
// }
func (w *WebAPI) GetChannels(args *GetChannelsArgs, response *GetChannelsResponse) error {
	channels := w.model.GetChannels()

	// Sort the channels alphabetically
	response.Channels = make([]string, 0)
	for channel := range channels {
		response.Channels = append(response.Channels, channel)
	}
	sort.Strings(response.Channels)

	return nil
}

// PostMessageArgs provides the input arguments for the PostMessage action.
type PostMessageArgs struct {
	Channelname string
	Username    string
	Text        string
}

// PostMessageResponse provides the output arguments for the PostMessage action.
type PostMessageResponse struct {
}

// PostMessage will post a message to a channel by a user.
//
// JSON RPC Definition
// -------------------
//
// Input
// {
//     "method": "<registeredAPI>.PostMessage",
//     "params": [{
//         "Channelname": "Channel1",
//         "Username": "User1",
//         "Text": "Message1"
//     }]
// }
//
// Output
// {
// }
func (w *WebAPI) PostMessage(args *PostMessageArgs, response *PostMessageResponse) error {
	w.model.PostMessage(args.Channelname, args.Username, time.Now(), args.Text)

	return nil
}
