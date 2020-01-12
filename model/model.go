// Package model provides model functionality for the chat server.  It responds to model actions
// and requests and tracks these changes in its local state.  It will also notify the subscription
// engine and logger as required.
package model

import (
	"chatserver/model/actions"
	"strings"
	"sync"
	"time"
)

// User provides information about a user.
type User struct {
	Name         string
	BlockedUsers []string
}

// Message provides data contained by a message.
type Message struct {
	Username  string
	Timestamp time.Time
	Text      string
}

// ChannelInfo provides information about a channel.
type ChannelInfo struct {
	Name        string
	NumMessages int
}

// Channel provides data contained by a channel.
type Channel struct {
	Name     string
	Messages []Message
}

// ActionsReplayer is the interface required to replay actions.
type ActionsReplayer interface {
	Replay(actor actions.Actor) error
}

// SubsEngine is the interface required to note subscription state changes.
type SubsEngine interface {
	UsersChanged()
	UserChanged(username string)
	ChannelsChanged()
	ChannelChanged(channelname string)
}

// Model provides an in memory store of the current state of the chat server.
type Model struct {
	actionsLogger actions.Actor
	subsEngine    SubsEngine
	mutex         sync.Mutex
	users         map[string]*User
	channels      map[string]*Channel
}

// NewModel creates/initializes/returns a new Model.
func NewModel(actionsReplayer ActionsReplayer, actionsLogger actions.Actor, subsEngine SubsEngine) (*Model, error) {
	model := Model{
		actionsLogger: actionsLogger,
		subsEngine:    subsEngine,
		users:         make(map[string]*User),
		channels:      make(map[string]*Channel),
	}

	if actionsReplayer == nil {
		// We are not restoring from an existing log, we need to create a new default state
		model.CreateUser("Anonymous")
		model.CreateChannel("General")
	} else {
		// Disable logging and subscriptions
		model.actionsLogger = nil
		model.subsEngine = nil

		// We've been given an actions replayer, replay the actions to initialize our state
		err := actionsReplayer.Replay(&model)
		if err != nil {
			return nil, err
		}

		// Enable logging and subscriptions
		model.actionsLogger = actionsLogger
		model.subsEngine = subsEngine
	}

	return &model, nil
}

// CreateUser creates a new user in the model.
func (m *Model) CreateUser(username string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If the user already exists, do nothing
	if _, ok := m.users[username]; ok {
		return
	}

	// Disallow adding of empty user
	if username == "" {
		return
	}

	// Disallow adding of user with space in username
	if strings.Contains(username, " ") {
		return
	}

	// Add the new user
	newUser := User{
		Name:         username,
		BlockedUsers: make([]string, 0),
	}
	m.users[newUser.Name] = &newUser

	// Handle logging and subscriptions
	if m.actionsLogger != nil {
		m.actionsLogger.CreateUser(username)
	}

	if m.subsEngine != nil {
		m.subsEngine.UsersChanged()
	}
}

// DeleteUser deletes an existing user from the model.
func (m *Model) DeleteUser(username string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If the user doesn't exist, do nothing
	if _, ok := m.users[username]; !ok {
		return
	}

	// Disallow deleting of Anonymous user
	if username == "Anonymous" {
		return
	}

	// Remove the user
	delete(m.users, username)

	// Remove the user from all other users' blockedUsers list
	for _, user := range m.users {
		removalIndex := -1
		for i, blockedUsername := range user.BlockedUsers {
			if blockedUsername == username {
				removalIndex = i
				break
			}
		}

		if removalIndex != -1 {
			user.BlockedUsers = append(user.BlockedUsers[:removalIndex], user.BlockedUsers[removalIndex+1:]...)
		}
	}

	// Handle logging and subscriptions
	if m.actionsLogger != nil {
		m.actionsLogger.DeleteUser(username)
	}

	if m.subsEngine != nil {
		m.subsEngine.UsersChanged()
	}
}

// GetUserInfo returns information about a requested user.
func (m *Model) GetUserInfo(username string) User {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If the user doesn't exist, do nothing
	if _, ok := m.users[username]; !ok {
		return User{}
	}

	// Copy and return the user
	user := m.users[username]
	userInfo := User{
		Name:         user.Name,
		BlockedUsers: make([]string, len(user.BlockedUsers)),
	}
	copy(userInfo.BlockedUsers, user.BlockedUsers)

	return userInfo
}

// GetUsers returns a list of all users.
func (m *Model) GetUsers() map[string]struct{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	users := make(map[string]struct{})
	for _, user := range m.users {
		users[user.Name] = struct{}{}
	}

	return users
}

// BlockUser blocks a user for a requested user.
func (m *Model) BlockUser(username string, usernameToBlock string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If the user doesn't exist, do nothing
	if _, ok := m.users[username]; !ok {
		return
	}

	// If the user to block doesn't exist, do nothing
	if _, ok := m.users[usernameToBlock]; !ok {
		return
	}

	// Don't allow the anonymous user to block
	if username == "Anonymous" {
		return
	}

	// Don't allow blocking yourself
	if username == usernameToBlock {
		return
	}

	// Look through the user's blockedUsers list and add the username if new
	user := m.users[username]

	found := false
	for _, blockedUser := range user.BlockedUsers {
		if blockedUser == usernameToBlock {
			found = true
			break
		}
	}

	if !found {
		user.BlockedUsers = append(user.BlockedUsers, usernameToBlock)
	}

	// Handle logging and subscriptions
	if m.actionsLogger != nil {
		m.actionsLogger.BlockUser(username, usernameToBlock)
	}

	if m.subsEngine != nil {
		m.subsEngine.UserChanged(username)
	}
}

// UnblockUser unblocks a user for a requested user.
func (m *Model) UnblockUser(username string, usernameToUnblock string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If the user doesn't exist, do nothing
	if _, ok := m.users[username]; !ok {
		return
	}

	// If the user to block doesn't exist, do nothing
	if _, ok := m.users[usernameToUnblock]; !ok {
		return
	}

	// Look through the user's blockedUsers list and add the username if new
	user := m.users[username]

	foundIndex := -1
	for i, blockedUser := range user.BlockedUsers {
		if blockedUser == usernameToUnblock {
			foundIndex = i
			break
		}
	}

	if foundIndex != -1 {
		user.BlockedUsers = append(user.BlockedUsers[:foundIndex], user.BlockedUsers[foundIndex+1:]...)
	}

	// Handle logging and subscriptions
	if m.actionsLogger != nil {
		m.actionsLogger.UnblockUser(username, usernameToUnblock)
	}

	if m.subsEngine != nil {
		m.subsEngine.UserChanged(username)
	}
}

// CreateChannel creates a new channel in the model.
func (m *Model) CreateChannel(channelname string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If the channel already exists, do nothing
	if _, ok := m.channels[channelname]; ok {
		return
	}

	// Disallow adding of empty channel
	if channelname == "" {
		return
	}

	// Disallow adding of channel with space in channelname
	if strings.Contains(channelname, " ") {
		return
	}

	// Add the channel
	newChannel := Channel{
		Name:     channelname,
		Messages: make([]Message, 0),
	}
	m.channels[channelname] = &newChannel

	// Handle logging and subscriptions
	if m.actionsLogger != nil {
		m.actionsLogger.CreateChannel(channelname)
	}

	if m.subsEngine != nil {
		m.subsEngine.ChannelsChanged()
	}
}

// DeleteChannel deletes an existing channel from the model.
func (m *Model) DeleteChannel(channelname string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If the channel doesn't exist, do nothing
	if _, ok := m.channels[channelname]; !ok {
		return
	}

	// Disallow deleting of the General channel
	if channelname == "General" {
		return
	}

	// Remove the channel
	delete(m.channels, channelname)

	// Handle logging and subscriptions
	if m.actionsLogger != nil {
		m.actionsLogger.DeleteChannel(channelname)
	}

	if m.subsEngine != nil {
		m.subsEngine.ChannelsChanged()
	}
}

// GetChannelInfo returns information about a requested channel.
func (m *Model) GetChannelInfo(channelname string) ChannelInfo {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// If the user doesn't exist, do nothing
	if _, ok := m.channels[channelname]; !ok {
		return ChannelInfo{}
	}

	// Copy and return the channel info
	channel := m.channels[channelname]
	channelInfo := ChannelInfo{
		Name:        channel.Name,
		NumMessages: len(channel.Messages),
	}

	return channelInfo
}

// GetChannelHistory returns message history for a requested channel
// filtered for a requested user up to some requested number of messages
// (-1 for all).
func (m *Model) GetChannelHistory(channelname string, username string, numMessages int) []Message {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate that channel exists
	if _, ok := m.channels[channelname]; !ok {
		return make([]Message, 0)
	}

	// Validate that user exists
	if _, ok := m.users[username]; !ok {
		return make([]Message, 0)
	}

	// Figure out which message to start copying from
	channel := m.channels[channelname]
	user := m.users[username]

	startingMessageIndex := len(channel.Messages) - numMessages
	if startingMessageIndex < 0 {
		startingMessageIndex = 0
	}

	// Copy all messages when numMessages is -1
	if numMessages == -1 {
		startingMessageIndex = 0
	}

	// Copy messages
	messages := make([]Message, 0)
	for i := startingMessageIndex; i < len(channel.Messages); i++ {
		fromBlockedUser := false
		for _, blockedUser := range user.BlockedUsers {
			if channel.Messages[i].Username == blockedUser {
				fromBlockedUser = true
				break
			}
		}

		if !fromBlockedUser {
			messages = append(messages, channel.Messages[i])
		}
	}

	return messages
}

// GetChannels returns a list of all channels.
func (m *Model) GetChannels() map[string]struct{} {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	channels := make(map[string]struct{})
	for _, channel := range m.channels {
		channels[channel.Name] = struct{}{}
	}

	return channels
}

// PostMessage posts a message to a requested channel for a requested user.
func (m *Model) PostMessage(channelname string, username string, timestamp time.Time, text string) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Validate that channel exists
	if _, ok := m.channels[channelname]; !ok {
		return
	}

	// Validate that user exists
	if _, ok := m.users[username]; !ok {
		return
	}

	// Disregard empty messages
	if len(text) == 0 {
		return
	}

	// Create the new message
	newMessage := Message{
		Username:  username,
		Timestamp: timestamp,
		Text:      text,
	}

	// Add the new message to the channel
	channel := m.channels[channelname]
	channel.Messages = append(channel.Messages, newMessage)

	// Handle logging and subscriptions
	if m.actionsLogger != nil {
		m.actionsLogger.PostMessage(channelname, username, timestamp, text)
	}

	if m.subsEngine != nil {
		m.subsEngine.ChannelChanged(channelname)
	}
}
