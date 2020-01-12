// Package telnetconn manages state associated with a single telnet view connection. It
// handles telnet commands for that connection, interacts with the model, and handles
// subscription updates.
package telnetconn

import (
	"chatserver/model"
	"sort"
	"strconv"
	"sync"
	"time"
)

const defaultHistoricalMessages int = 10
const defaultSeparator string = "-----------------"

// PrintLinesCallback is the function signature that clients will provide in order
// to give the TelnetConn the ability to output text data.
type PrintLinesCallback = func(lines []string)

// TelnetConn manages data associated with a single telnet view connection.  This
// includes things like which user the connection is currently using and which
// channel is currently being viewed.
type TelnetConn struct {
	model                      *model.Model
	printLinesCallback         PrintLinesCallback
	currentUser                string
	currentChannel             string
	currentChannelMessageIndex int
	mutex                      sync.Mutex
}

// NewTelnetConn creates/initializes/returns a new TelnetConn.  It will default the
// connection to the "Anonymous" user as well as the "General" channel.
func NewTelnetConn(model *model.Model, printLinesCallback PrintLinesCallback) *TelnetConn {
	telnetConn := TelnetConn{
		model:                      model,
		printLinesCallback:         printLinesCallback,
		currentUser:                "None",
		currentChannel:             "None",
		currentChannelMessageIndex: 0,
	}

	// Default to the Anonymous user
	telnetConn.SwitchUser("Anonymous")

	return &telnetConn
}

// OnUsersChanged is called whenever the users state changes in the model.
func (t *TelnetConn) OnUsersChanged() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	users := t.model.GetUsers()

	// If our current user has been deleted, switch to Anonymous
	if _, ok := users[t.currentUser]; !ok {
		t.switchUser("Anonymous")
	}
}

// OnUserChanged is called whenever a particular user's state changes in the model.
func (t *TelnetConn) OnUserChanged(username string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// If our current user has changed, we may need to reprint channel
	// history to hide/show newly blocked/unblocked messages
	if t.currentUser == username {
		t.showChannelHistory(defaultHistoricalMessages)
	}
}

// OnChannelsChanged is called whenever the channels state changes in the model.
func (t *TelnetConn) OnChannelsChanged() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	channels := t.model.GetChannels()

	// If our current channel has been deleted, switch to General
	if _, ok := channels[t.currentChannel]; !ok {
		t.switchChannel("General")
	}
}

// OnChannelChanged is called whenever a particular channel's state changes in the model.
func (t *TelnetConn) OnChannelChanged(channelname string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// If our current channel has changed, then see if we need to post any new messages
	if t.currentChannel == channelname {
		channelInfo := t.model.GetChannelInfo(channelname)
		numNewMessages := channelInfo.NumMessages - t.currentChannelMessageIndex
		t.showChannelHistory(numNewMessages)
	}
}

// ShowUsers will print a list of all of the users in the model.
func (t *TelnetConn) ShowUsers() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	users := t.model.GetUsers()

	// Sort the users alphabetically
	sortedUsers := make([]string, 0)
	for user := range users {
		sortedUsers = append(sortedUsers, user)
	}
	sort.Strings(sortedUsers)

	// Tell the client about the users
	msg := make([]string, 0)
	msg = append(msg, defaultSeparator)
	for _, user := range sortedUsers {
		if user == t.currentUser {
			msg = append(msg, "--> "+user+" <--")
		} else {
			msg = append(msg, user)
		}
	}
	msg = append(msg, defaultSeparator)
	t.printLinesCallback(msg)
}

// SwitchUser will change the user that is associated with the current telnet view connection.
func (t *TelnetConn) SwitchUser(username string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Call the private (lock held) version
	t.switchUser(username)
}

// ShowUserInfo will print information associated with the current user.
func (t *TelnetConn) ShowUserInfo() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	userInfo := t.model.GetUserInfo(t.currentUser)

	// Sort the blocked users alphabetically
	sort.Strings(userInfo.BlockedUsers)

	// Tell the client about the user info
	msg := make([]string, 0)
	msg = append(msg, defaultSeparator)
	msg = append(msg, "User: "+userInfo.Name)
	msg = append(msg, "Blocked Users:")
	for _, blockedUser := range userInfo.BlockedUsers {
		msg = append(msg, "    "+blockedUser)
	}
	msg = append(msg, defaultSeparator)
	t.printLinesCallback(msg)
}

// CreateUser will create a new user.
func (t *TelnetConn) CreateUser(username string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	users := t.model.GetUsers()

	// Validate the user input
	if _, ok := users[username]; ok {
		msg := make([]string, 0)
		msg = append(msg, "error: <user> already exists")
		t.printLinesCallback(msg)
		return
	}

	// Tell the model about the new user
	t.model.CreateUser(username)
}

// DeleteUser will delete an existing user.
func (t *TelnetConn) DeleteUser(username string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	users := t.model.GetUsers()

	// Validate the user input
	if _, ok := users[username]; !ok {
		msg := make([]string, 0)
		msg = append(msg, "error: <user> not found")
		t.printLinesCallback(msg)
		return
	}

	// Delete the user in the model
	t.model.DeleteUser(username)
}

// BlockUser will add a new user to the current user's blocked user list.
func (t *TelnetConn) BlockUser(username string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	users := t.model.GetUsers()

	// Validate the user input
	if _, ok := users[username]; !ok {
		msg := make([]string, 0)
		msg = append(msg, "error: <user> not found")
		t.printLinesCallback(msg)
		return
	}

	t.model.BlockUser(t.currentUser, username)
}

// UnblockUser will delete an existing user from the current user's blocked user list.
func (t *TelnetConn) UnblockUser(username string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	users := t.model.GetUsers()

	// Validate the user input
	if _, ok := users[username]; !ok {
		msg := make([]string, 0)
		msg = append(msg, "error: <user> not found")
		t.printLinesCallback(msg)
		return
	}

	t.model.UnblockUser(t.currentUser, username)
}

// ShowChannels will print a list of all of the channels in the model.
func (t *TelnetConn) ShowChannels() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	channels := t.model.GetChannels()

	// Sort the channels alphabetically
	sortedChannels := make([]string, 0)
	for channel := range channels {
		sortedChannels = append(sortedChannels, channel)
	}
	sort.Strings(sortedChannels)

	// Tell the client about the channels
	msg := make([]string, 0)
	msg = append(msg, defaultSeparator)
	for _, channel := range sortedChannels {
		if channel == t.currentChannel {
			msg = append(msg, "--> "+channel+" <--")
		} else {
			msg = append(msg, channel)
		}
	}
	msg = append(msg, defaultSeparator)
	t.printLinesCallback(msg)
}

// SwitchChannel will change the channel that the current user is viewing.
func (t *TelnetConn) SwitchChannel(channelname string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Call the private (lock held) version
	t.switchChannel(channelname)
}

// ShowChannelInfo will print information associated with the current channel.
func (t *TelnetConn) ShowChannelInfo() {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	channelInfo := t.model.GetChannelInfo(t.currentChannel)

	// Tell the client about the channel info
	msg := make([]string, 0)
	msg = append(msg, defaultSeparator)
	msg = append(msg, "Channel: "+channelInfo.Name)
	msg = append(msg, "Messages: "+strconv.Itoa(channelInfo.NumMessages))
	msg = append(msg, defaultSeparator)
	t.printLinesCallback(msg)
}

// ShowChannelHistory will print up to 'numMessages' worth of history from the current channel
// (NOTE: '-1' will print all messages).
func (t *TelnetConn) ShowChannelHistory(numMessages int) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	// Call the private (lock held) version
	t.showChannelHistory(numMessages)
}

// CreateChannel will create a new channel.
func (t *TelnetConn) CreateChannel(channelname string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	channels := t.model.GetChannels()

	// Validate the user input
	if _, ok := channels[channelname]; ok {
		msg := make([]string, 0)
		msg = append(msg, "error: <channel> already exists")
		t.printLinesCallback(msg)
		return
	}

	// Tell the model about the new channel
	t.model.CreateChannel(channelname)
}

// DeleteChannel will delete an existing channel.
func (t *TelnetConn) DeleteChannel(channelname string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	channels := t.model.GetChannels()

	// Validate the user input
	if _, ok := channels[channelname]; !ok {
		msg := make([]string, 0)
		msg = append(msg, "error: <channel> not found")
		t.printLinesCallback(msg)
		return
	}

	// Delete the channel in the model
	t.model.DeleteChannel(channelname)
}

// PostMessage will post a new message to the current channel by the current user.
func (t *TelnetConn) PostMessage(text string) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.model.PostMessage(t.currentChannel, t.currentUser, time.Now(), text)
}

func (t *TelnetConn) showChannelHistory(numMessages int) {
	// This will always bring us up to date with the channel messages
	channelInfo := t.model.GetChannelInfo(t.currentChannel)
	t.currentChannelMessageIndex = channelInfo.NumMessages

	messages := t.model.GetChannelHistory(t.currentChannel, t.currentUser, numMessages)

	// Tell the client about the messages
	msg := make([]string, 0)
	for _, message := range messages {
		timestamp := message.Timestamp.Format("2006-01-02 15:04:05")
		msg = append(msg, "["+timestamp+" - "+message.Username+"] "+message.Text)
	}
	t.printLinesCallback(msg)
}

func (t *TelnetConn) switchUser(username string) {
	users := t.model.GetUsers()

	// Validate the user input
	if _, ok := users[username]; !ok {
		msg := make([]string, 0)
		msg = append(msg, "error: <user> not found")
		t.printLinesCallback(msg)
		return
	}

	// Update the current user
	t.currentUser = username

	// Switch channels
	t.switchChannel("General")
}

func (t *TelnetConn) switchChannel(channelname string) {
	channels := t.model.GetChannels()

	// Validate the user input
	if _, ok := channels[channelname]; !ok {
		msg := make([]string, 0)
		msg = append(msg, "error: <channel> not found")
		t.printLinesCallback(msg)
		return
	}

	// Update the current channel
	t.currentChannel = channelname

	// Tell the client about the new channel
	msg := make([]string, 0)
	msg = append(msg, defaultSeparator)
	msg = append(msg, "User: "+t.currentUser)
	msg = append(msg, "Channel: "+t.currentChannel)
	msg = append(msg, defaultSeparator)
	t.printLinesCallback(msg)

	// Show channel history
	t.showChannelHistory(defaultHistoricalMessages)
}
