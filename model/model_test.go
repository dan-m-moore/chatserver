package model_test

import (
	"chatserver/model"
	"chatserver/model/actions"
	"chatserver/model/subs"
	"errors"
	"testing"
	"time"
)

func TestEmptyModelSetup(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	users := testModel.GetUsers()
	if len(users) != 1 {
		t.Error("Incorrect number of users")
	}

	if _, ok := users["Anonymous"]; !ok {
		t.Error("No Anonymous user")
	}

	channels := testModel.GetChannels()
	if len(channels) != 1 {
		t.Error("Incorrect number of channels")
	}

	if _, ok := channels["General"]; !ok {
		t.Error("No General channel")
	}
}

func TestCreateUserInputChecking(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	testModel.CreateUser("")
	users := testModel.GetUsers()
	if len(users) != 1 {
		t.Error("Incorrect number of users")
	}

	testModel.CreateUser("user 1")
	users = testModel.GetUsers()
	if len(users) != 1 {
		t.Error("Incorrect number of users")
	}

	testModel.CreateUser("Anonymous")
	users = testModel.GetUsers()
	if len(users) != 1 {
		t.Error("Incorrect number of users")
	}
}

func TestCreateAndDeleteUser(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	// Create a single user and verify that it is added
	testModel.CreateUser("user1")
	users := testModel.GetUsers()
	if len(users) != 2 {
		t.Error("Incorrect number of users")
	}

	if _, ok := users["user1"]; !ok {
		t.Error("Failed to CreateUser(user1)")
	}

	// Create another user with the same name and verify that it is not added again
	testModel.CreateUser("user1")
	users = testModel.GetUsers()
	if len(users) != 2 {
		t.Error("Incorrect number of users")
	}

	if _, ok := users["user1"]; !ok {
		t.Error("Failed to CreateUser(user1)")
	}

	// Delete the user and verify that it is deleted
	testModel.DeleteUser("user1")
	users = testModel.GetUsers()
	if len(users) != 1 {
		t.Error("Incorrect number of users")
	}

	if _, ok := users["user1"]; ok {
		t.Error("Failed to DeleteUser(user1)")
	}

	// Delete the user again and verify that it is not deleted again
	testModel.DeleteUser("user1")
	users = testModel.GetUsers()
	if len(users) != 1 {
		t.Error("Incorrect number of users")
	}

	if _, ok := users["user1"]; ok {
		t.Error("Failed to DeleteUser(user1)")
	}
}

func TestCreateAndDeleteAnonymousUser(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	// Ensure that we can't create or delete the Anonymous user
	testModel.CreateUser("Anonymous")
	users := testModel.GetUsers()
	if len(users) != 1 {
		t.Error("Incorrect number of users")
	}

	if _, ok := users["Anonymous"]; !ok {
		t.Error("Messed up Anonymous user")
	}

	testModel.DeleteUser("Anonymous")
	users = testModel.GetUsers()
	if len(users) != 1 {
		t.Error("Incorrect number of users")
	}

	if _, ok := users["Anonymous"]; !ok {
		t.Error("Messed up Anonymous user")
	}
}

func TestGetUserInfo(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	userInfo := testModel.GetUserInfo("user1")
	if userInfo.Name != "" {
		t.Error("Failed to return empty user info")
	}

	userInfo = testModel.GetUserInfo("Anonymous")
	if userInfo.Name != "Anonymous" {
		t.Error("Failed to return Anonymous user info")
	}
}

func TestBlockUserInputChecking(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	testModel.BlockUser("user1", "Anonymous")
	userInfo := testModel.GetUserInfo("user1")
	if userInfo.Name != "" {
		t.Error("Failed to disregard block call for unknown user")
	}

	testModel.CreateUser("user1")
	testModel.BlockUser("user1", "user2")
	userInfo = testModel.GetUserInfo("user1")
	if len(userInfo.BlockedUsers) != 0 {
		t.Error("Failed to disallow blocking of unknown user")
	}

	testModel.BlockUser("Anonymous", "user1")
	userInfo = testModel.GetUserInfo("Anonymous")
	if len(userInfo.BlockedUsers) != 0 {
		t.Error("Failed to disallow blocking for Anonymous user")
	}

	testModel.BlockUser("user1", "user1")
	userInfo = testModel.GetUserInfo("user1")
	if len(userInfo.BlockedUsers) != 0 {
		t.Error("Failed to disallow blocking for same user")
	}
}

func TestUnblockUserInputChecking(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	testModel.UnblockUser("user1", "Anonymous")
	userInfo := testModel.GetUserInfo("user1")
	if userInfo.Name != "" {
		t.Error("Failed to disregard unblock call for unknown user")
	}

	testModel.CreateUser("user1")
	testModel.CreateUser("user2")
	testModel.BlockUser("user1", "user2")
	testModel.UnblockUser("user1", "user3")
	userInfo = testModel.GetUserInfo("user1")
	if len(userInfo.BlockedUsers) != 1 || userInfo.BlockedUsers[0] != "user2" {
		t.Error("Failed to disallow unblocking of unknown user")
	}
}

func TestBlockingAndUnblockingUsers(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	// Add 2 users
	testModel.CreateUser("user1")
	testModel.CreateUser("user2")

	// Verify that their user info is correct
	user1Info := testModel.GetUserInfo("user1")
	if user1Info.Name != "user1" || len(user1Info.BlockedUsers) != 0 {
		t.Error("Invalid initial user info for user1")
	}

	user2Info := testModel.GetUserInfo("user2")
	if user2Info.Name != "user2" || len(user2Info.BlockedUsers) != 0 {
		t.Error("Invalid initial user info for user2")
	}

	// Add user2 to user1's BlockedUsers list and verify that both user infos are correct
	testModel.BlockUser("user1", "user2")
	user1Info = testModel.GetUserInfo("user1")
	if len(user1Info.BlockedUsers) != 1 || user1Info.BlockedUsers[0] != "user2" {
		t.Error("Failed to block user2 for user1")
	}

	user2Info = testModel.GetUserInfo("user2")
	if len(user2Info.BlockedUsers) != 0 {
		t.Error("Invalid user info for user2")
	}

	// Attempt to block user2 again and ensure it's not added twice
	testModel.BlockUser("user1", "user2")
	user1Info = testModel.GetUserInfo("user1")
	if len(user1Info.BlockedUsers) != 1 || user1Info.BlockedUsers[0] != "user2" {
		t.Error("Failed to block user2 for user1")
	}

	// Remove user2 from user1's BlockedUsers list
	testModel.UnblockUser("user1", "user2")
	user1Info = testModel.GetUserInfo("user1")
	if len(user1Info.BlockedUsers) != 0 {
		t.Error("Failed to unblock user2 for user1")
	}

	user2Info = testModel.GetUserInfo("user2")
	if len(user2Info.BlockedUsers) != 0 {
		t.Error("Invalid user info for user2")
	}

	// Attempt to unblock user2 again and ensure it's not removed twice
	testModel.UnblockUser("user1", "user2")
	user1Info = testModel.GetUserInfo("user1")
	if len(user1Info.BlockedUsers) != 0 {
		t.Error("Failed to unblock user2 for user1")
	}
}

func TestBlockingAndDeletingUsers(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	testModel.CreateUser("user1")
	testModel.CreateUser("user2")
	testModel.CreateUser("user3")
	testModel.CreateUser("user4")
	testModel.CreateUser("user5")

	users := testModel.GetUsers()
	if len(users) != 6 {
		t.Error("Failed to create 5 users")
	}

	testModel.BlockUser("user1", "user3")
	testModel.BlockUser("user1", "user4")
	testModel.BlockUser("user1", "user5")

	testModel.BlockUser("user2", "user3")
	testModel.BlockUser("user2", "user4")
	testModel.BlockUser("user2", "user5")

	user1Info := testModel.GetUserInfo("user1")
	if len(user1Info.BlockedUsers) != 3 {
		t.Error("Failed to block 3 users for user1")
	}

	user2Info := testModel.GetUserInfo("user2")
	if len(user2Info.BlockedUsers) != 3 {
		t.Error("Failed to block 3 users for user2")
	}

	testModel.DeleteUser("user3")
	testModel.DeleteUser("user4")
	testModel.DeleteUser("user5")

	users = testModel.GetUsers()
	if len(users) != 3 {
		t.Error("Failed to delete 3 users")
	}

	user1Info = testModel.GetUserInfo("user1")
	if len(user1Info.BlockedUsers) != 0 {
		t.Error("Failed to clean up blocked users on delete for user1")
	}

	user2Info = testModel.GetUserInfo("user2")
	if len(user2Info.BlockedUsers) != 0 {
		t.Error("Failed to clean up blocked users on delete for user2")
	}
}

func TestCreateChannelInputChecking(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	testModel.CreateChannel("")
	channels := testModel.GetChannels()
	if len(channels) != 1 {
		t.Error("Incorrect number of channels")
	}

	testModel.CreateChannel("channel 1")
	channels = testModel.GetChannels()
	if len(channels) != 1 {
		t.Error("Incorrect number of channels")
	}

	testModel.CreateChannel("General")
	channels = testModel.GetChannels()
	if len(channels) != 1 {
		t.Error("Incorrect number of channels")
	}
}

func TestCreateAndDeleteChannel(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	// Create a single channel and verify that it is added
	testModel.CreateChannel("channel1")
	channels := testModel.GetChannels()
	if len(channels) != 2 {
		t.Error("Incorrect number of channels")
	}

	if _, ok := channels["channel1"]; !ok {
		t.Error("Failed to create channel1")
	}

	// Create another channel with the same name and verify that it is not added again
	testModel.CreateChannel("channel1")
	channels = testModel.GetChannels()
	if len(channels) != 2 {
		t.Error("Incorrect number of channels")
	}

	if _, ok := channels["channel1"]; !ok {
		t.Error("Failed to create channel1")
	}

	// Delete the channel and verify that it is deleted
	testModel.DeleteChannel("channel1")
	channels = testModel.GetChannels()
	if len(channels) != 1 {
		t.Error("Incorrect number of channels")
	}

	if _, ok := channels["channels"]; ok {
		t.Error("Failed to delete channel1")
	}

	// Delete the channel again and verify that it is not deleted again
	testModel.DeleteChannel("channel1")
	channels = testModel.GetChannels()
	if len(channels) != 1 {
		t.Error("Incorrect number of channels")
	}

	if _, ok := channels["channel1"]; ok {
		t.Error("Failed to delete channel1")
	}
}

func TestCreateAndDeleteGeneralChannel(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	// Ensure that we can't create or delete the General channel
	testModel.CreateChannel("General")
	channels := testModel.GetChannels()
	if len(channels) != 1 {
		t.Error("Incorrect number of channels")
	}

	if _, ok := channels["General"]; !ok {
		t.Error("Messed up General channel")
	}

	testModel.DeleteChannel("General")
	channels = testModel.GetChannels()
	if len(channels) != 1 {
		t.Error("Incorrect number of channels")
	}

	if _, ok := channels["General"]; !ok {
		t.Error("Messed up General channel")
	}
}

func TestGetChannelInfo(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	channelInfo := testModel.GetChannelInfo("channel1")
	if channelInfo.Name != "" {
		t.Error("Failed to return empty channel info")
	}

	channelInfo = testModel.GetChannelInfo("General")
	if channelInfo.Name != "General" {
		t.Error("Failed to return General channel info")
	}
}

func TestCreatingAndDeletingMultipleChannels(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	testModel.CreateChannel("channel1")
	testModel.CreateChannel("channel2")
	testModel.CreateChannel("channel3")
	testModel.CreateChannel("channel4")
	testModel.CreateChannel("channel5")

	channels := testModel.GetChannels()
	if len(channels) != 6 {
		t.Error("Failed to create 5 channels")
	}

	testModel.DeleteChannel("channel2")
	testModel.DeleteChannel("channel4")
	testModel.DeleteChannel("channel5")

	channels = testModel.GetChannels()
	if len(channels) != 3 {
		t.Error("Failed to delete 3 channels")
	}

	channel1Info := testModel.GetChannelInfo("channel1")
	if channel1Info.Name != "channel1" || channel1Info.NumMessages != 0 {
		t.Error("Messed up channel1 info")
	}

	channel3Info := testModel.GetChannelInfo("channel3")
	if channel3Info.Name != "channel3" || channel3Info.NumMessages != 0 {
		t.Error("Messed up channel3 info")
	}
}

func TestGetChannelHistoryInputChecking(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	messages := testModel.GetChannelHistory("", "Anonymous", -1)
	if len(messages) != 0 {
		t.Error("Failed to disregard GetChannelHistory for unknown channel")
	}

	messages = testModel.GetChannelHistory("General", "", -1)
	if len(messages) != 0 {
		t.Error("Failed to disregard GetChannelHistory for unknown user")
	}
}

func TestPostMessageInputChecking(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	testModel.PostMessage("", "Anonymous", time.Now(), "message1")
	channelInfo := testModel.GetChannelInfo("General")
	if channelInfo.NumMessages != 0 {
		t.Error("Failed to disregard PostMessage for unknown channel")
	}

	testModel.PostMessage("General", "", time.Now(), "message1")
	channelInfo = testModel.GetChannelInfo("General")
	if channelInfo.NumMessages != 0 {
		t.Error("Failed to disregard PostMessage for unknown user")
	}

	testModel.PostMessage("General", "Anonymous", time.Now(), "")
	channelInfo = testModel.GetChannelInfo("General")
	if channelInfo.NumMessages != 0 {
		t.Error("Failed to disregard PostMessage for empty message")
	}
}

func TestPostMessage(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	testModel.CreateChannel("channel1")
	testModel.CreateUser("user1")

	testModel.PostMessage("channel1", "user1", time.Now(), "message1")
	testModel.PostMessage("channel1", "user1", time.Now(), "message2")
	testModel.PostMessage("channel1", "user1", time.Now(), "message3")
	testModel.PostMessage("channel1", "user1", time.Now(), "message4")

	channel1Info := testModel.GetChannelInfo("channel1")
	if channel1Info.NumMessages != 4 {
		t.Error("Failed to count message after PostMessage")
	}

	// Ensure that we get the newest messages
	messages := testModel.GetChannelHistory("channel1", "Anonymous", 1)
	if len(messages) != 1 || messages[0].Username != "user1" || messages[0].Text != "message4" {
		t.Error("Failed to get message after PostMessage")
	}

	// Ensure that we can get all of the messages
	messages = testModel.GetChannelHistory("channel1", "Anonymous", 5)
	if len(messages) != 4 {
		t.Error("Failed to get multiple messages after PostMessage")
	}

	if messages[0].Text != "message1" || messages[1].Text != "message2" || messages[2].Text != "message3" || messages[3].Text != "message4" {
		t.Error("Failed to get correct messages after PostMessage")
	}

	// Ensure that we can get all of the messages
	messages = testModel.GetChannelHistory("channel1", "Anonymous", -1)
	if len(messages) != 4 {
		t.Error("Failed to get multiple messages after PostMessage")
	}

	if messages[0].Text != "message1" || messages[1].Text != "message2" || messages[2].Text != "message3" || messages[3].Text != "message4" {
		t.Error("Failed to get correct messages after PostMessage")
	}
}

func TestFilteringBlockedUserMessages(t *testing.T) {
	testModel, err := model.NewModel(nil, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	testModel.CreateChannel("channel1")
	testModel.CreateUser("user1")

	testModel.BlockUser("user1", "Anonymous")

	testModel.PostMessage("channel1", "user1", time.Now(), "message1")
	testModel.PostMessage("channel1", "Anonymous", time.Now(), "message2")
	testModel.PostMessage("channel1", "Anonymous", time.Now(), "message3")
	testModel.PostMessage("channel1", "user1", time.Now(), "message4")
	testModel.PostMessage("channel1", "Anonymous", time.Now(), "message5")

	channel1Info := testModel.GetChannelInfo("channel1")
	if channel1Info.NumMessages != 5 {
		t.Error("Failed to count message after PostMessage")
	}

	messages := testModel.GetChannelHistory("channel1", "user1", 1)
	if len(messages) != 0 {
		t.Error("Failed to filter messages for user1")
	}

	messages = testModel.GetChannelHistory("channel1", "Anonymous", 10)
	if len(messages) != 5 {
		t.Error("Failed to get multiple messages after PostMessage")
	}

	messages = testModel.GetChannelHistory("channel1", "user1", 10)
	if len(messages) != 2 {
		t.Error("Failed to filter messages for user1")
	}

	if messages[0].Text != "message1" || messages[1].Text != "message4" {
		t.Error("Failed to get correct messages after PostMessage")
	}

	testModel.UnblockUser("user1", "Anonymous")

	messages = testModel.GetChannelHistory("channel1", "user1", 3)
	if len(messages) != 3 {
		t.Error("Failed to filter messages for user1")
	}

	if messages[0].Text != "message3" || messages[1].Text != "message4" || messages[2].Text != "message5" {
		t.Error("Failed to get correct messages after PostMessage")
	}
}

type TestSubsEngine struct {
	UsersChangedCalled        int
	UserChangedCalled         int
	UserChangedUsername       []string
	ChannelsChangedCalled     int
	ChannelChangedCalled      int
	ChannelChangedChannelname []string
}

func NewTestSubsEngine() *TestSubsEngine {
	t := TestSubsEngine{}
	t.Reset()

	return &t
}

func (t *TestSubsEngine) Reset() {
	t.UsersChangedCalled = 0
	t.UserChangedCalled = 0
	t.UserChangedUsername = make([]string, 0)
	t.ChannelsChangedCalled = 0
	t.ChannelChangedCalled = 0
	t.ChannelChangedChannelname = make([]string, 0)
}

func (t *TestSubsEngine) Connect(client subs.Client) error {
	return nil
}

func (t *TestSubsEngine) Disconnect(client subs.Client) error {
	return nil
}

func (t *TestSubsEngine) UsersChanged() {
	t.UsersChangedCalled++
}

func (t *TestSubsEngine) UserChanged(username string) {
	t.UserChangedCalled++
	t.UserChangedUsername = append(t.UserChangedUsername, username)
}

func (t *TestSubsEngine) ChannelsChanged() {
	t.ChannelsChangedCalled++
}

func (t *TestSubsEngine) ChannelChanged(channelname string) {
	t.ChannelChangedCalled++
	t.ChannelChangedChannelname = append(t.ChannelChangedChannelname, channelname)
}

func TestSubscriptions(t *testing.T) {
	testSubsEngine := NewTestSubsEngine()
	testModel, err := model.NewModel(nil, nil, testSubsEngine)
	if err != nil {
		t.Error("Failed to create model")
	}

	if testSubsEngine.UsersChangedCalled != 1 {
		t.Error("Didn't create Anonymous user")
	}

	if testSubsEngine.ChannelsChangedCalled != 1 {
		t.Error("Didn't create General channel")
	}

	testSubsEngine.Reset()
	testModel.CreateUser("user1")
	if testSubsEngine.UsersChangedCalled != 1 {
		t.Error("CreateUser didn't correctly notify subscriptions")
	}

	testSubsEngine.Reset()
	testModel.DeleteUser("user1")
	if testSubsEngine.UsersChangedCalled != 1 {
		t.Error("DeleteUser didn't correctly notify subscriptions")
	}

	testModel.CreateUser("user1")
	testSubsEngine.Reset()
	testModel.BlockUser("user1", "Anonymous")
	if testSubsEngine.UserChangedCalled != 1 || testSubsEngine.UserChangedUsername[0] != "user1" {
		t.Error("BlockUser didn't correctly notify subscriptions")
	}

	testSubsEngine.Reset()
	testModel.UnblockUser("user1", "Anonymous")
	if testSubsEngine.UserChangedCalled != 1 || testSubsEngine.UserChangedUsername[0] != "user1" {
		t.Error("UnblockUser didn't correctly notify subscriptions")
	}

	testSubsEngine.Reset()
	testModel.CreateChannel("channel1")
	if testSubsEngine.ChannelsChangedCalled != 1 {
		t.Error("CreateChannel didn't correctly notify subscriptions")
	}

	testSubsEngine.Reset()
	testModel.DeleteChannel("channel1")
	if testSubsEngine.ChannelsChangedCalled != 1 {
		t.Error("DeleteChannel didn't correctly notify subscriptions")
	}

	testModel.CreateChannel("channel1")
	testSubsEngine.Reset()
	testModel.PostMessage("channel1", "user1", time.Now(), "message1")
	if testSubsEngine.ChannelChangedCalled != 1 || testSubsEngine.ChannelChangedChannelname[0] != "channel1" {
		t.Error("PostMessage didn't correctly notify subscriptions")
	}
}

type TestActionsReplayer struct {
	ReplayCalled int
	ReplayActor  []actions.Actor
	ReplayError  error
}

func NewTestActionsReplayer() *TestActionsReplayer {
	t := TestActionsReplayer{}
	t.Reset()

	return &t
}

func (t *TestActionsReplayer) Reset() {
	t.ReplayCalled = 0
	t.ReplayActor = make([]actions.Actor, 0)
	t.ReplayError = nil
}

func (t *TestActionsReplayer) Replay(actor actions.Actor) error {
	t.ReplayCalled++
	t.ReplayActor = append(t.ReplayActor, actor)
	return t.ReplayError
}

func TestActionReplay(t *testing.T) {
	testActionsReplayer := NewTestActionsReplayer()

	testActionsReplayer.ReplayError = errors.New("Failed replay")
	testModel, err := model.NewModel(testActionsReplayer, nil, nil)
	if err == nil {
		t.Error("NewModel didn't fail when replayer did")
	}

	testActionsReplayer.Reset()
	testModel, err = model.NewModel(testActionsReplayer, nil, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	if testActionsReplayer.ReplayCalled != 1 || testActionsReplayer.ReplayActor[0] != testModel {
		t.Error("Incorrect usage of the actionsReplayer")
	}
}

type TestActionsLogger struct {
	CreateUserCalled             int
	CreateUserUsername           []string
	DeleteUserCalled             int
	DeleteUserUsername           []string
	BlockUserCalled              int
	BlockUserUsername            []string
	BlockUserUsernameToBlock     []string
	UnblockUserCalled            int
	UnblockUserUsername          []string
	UnblockUserUsernameToUnblock []string
	CreateChannelCalled          int
	CreateChannelChannelname     []string
	DeleteChannelCalled          int
	DeleteChannelChannelname     []string
	PostMessageCalled            int
	PostMessageChannelname       []string
	PostMessageUsername          []string
	PostMessageTimestamp         []time.Time
	PostMessageText              []string
}

func NewTestActionsLogger() *TestActionsLogger {
	t := TestActionsLogger{}
	t.Reset()

	return &t
}

func (t *TestActionsLogger) Reset() {
	t.CreateUserCalled = 0
	t.CreateUserUsername = make([]string, 0)
	t.DeleteUserCalled = 0
	t.DeleteUserUsername = make([]string, 0)
	t.BlockUserCalled = 0
	t.BlockUserUsername = make([]string, 0)
	t.BlockUserUsernameToBlock = make([]string, 0)
	t.UnblockUserCalled = 0
	t.UnblockUserUsername = make([]string, 0)
	t.UnblockUserUsernameToUnblock = make([]string, 0)
	t.CreateChannelCalled = 0
	t.CreateChannelChannelname = make([]string, 0)
	t.DeleteChannelCalled = 0
	t.DeleteChannelChannelname = make([]string, 0)
	t.PostMessageCalled = 0
	t.PostMessageChannelname = make([]string, 0)
	t.PostMessageUsername = make([]string, 0)
	t.PostMessageTimestamp = make([]time.Time, 0)
	t.PostMessageText = make([]string, 0)
}

func (t *TestActionsLogger) CreateUser(username string) {
	t.CreateUserCalled++
	t.CreateUserUsername = append(t.CreateUserUsername, username)
}

func (t *TestActionsLogger) DeleteUser(username string) {
	t.DeleteUserCalled++
	t.DeleteUserUsername = append(t.DeleteUserUsername, username)
}

func (t *TestActionsLogger) BlockUser(username string, usernameToBlock string) {
	t.BlockUserCalled++
	t.BlockUserUsername = append(t.BlockUserUsername, username)
	t.BlockUserUsernameToBlock = append(t.BlockUserUsernameToBlock, usernameToBlock)
}

func (t *TestActionsLogger) UnblockUser(username string, usernameToUnblock string) {
	t.UnblockUserCalled++
	t.UnblockUserUsername = append(t.UnblockUserUsername, username)
	t.UnblockUserUsernameToUnblock = append(t.UnblockUserUsernameToUnblock, usernameToUnblock)
}

func (t *TestActionsLogger) CreateChannel(channelname string) {
	t.CreateChannelCalled++
	t.CreateChannelChannelname = append(t.CreateChannelChannelname, channelname)
}

func (t *TestActionsLogger) DeleteChannel(channelname string) {
	t.DeleteChannelCalled++
	t.DeleteChannelChannelname = append(t.DeleteChannelChannelname, channelname)
}

func (t *TestActionsLogger) PostMessage(channelname string, username string, timestamp time.Time, text string) {
	t.PostMessageCalled++
	t.PostMessageChannelname = append(t.PostMessageChannelname, channelname)
	t.PostMessageUsername = append(t.PostMessageUsername, username)
	t.PostMessageTimestamp = append(t.PostMessageTimestamp, timestamp)
	t.PostMessageText = append(t.PostMessageText, text)
}

func TestActionLogging(t *testing.T) {
	testActionsLogger := NewTestActionsLogger()
	testModel, err := model.NewModel(nil, testActionsLogger, nil)
	if err != nil {
		t.Error("Failed to create model")
	}

	if testActionsLogger.CreateUserCalled != 1 || testActionsLogger.CreateUserUsername[0] != "Anonymous" {
		t.Error("Didn't create Anonymous user")
	}

	if testActionsLogger.CreateChannelCalled != 1 || testActionsLogger.CreateChannelChannelname[0] != "General" {
		t.Error("Didn't create General channel")
	}

	testActionsLogger.Reset()
	testModel.CreateUser("user1")
	if testActionsLogger.CreateUserCalled != 1 || testActionsLogger.CreateUserUsername[0] != "user1" {
		t.Error("CreateUser didn't correctly log action")
	}

	testActionsLogger.Reset()
	testModel.DeleteUser("user1")
	if testActionsLogger.DeleteUserCalled != 1 || testActionsLogger.DeleteUserUsername[0] != "user1" {
		t.Error("DeleteUser didn't correctly log action")
	}

	testModel.CreateUser("user1")
	testActionsLogger.Reset()
	testModel.BlockUser("user1", "Anonymous")
	if testActionsLogger.BlockUserCalled != 1 || testActionsLogger.BlockUserUsername[0] != "user1" || testActionsLogger.BlockUserUsernameToBlock[0] != "Anonymous" {
		t.Error("BlockUser didn't correctly log action")
	}

	testActionsLogger.Reset()
	testModel.UnblockUser("user1", "Anonymous")
	if testActionsLogger.UnblockUserCalled != 1 || testActionsLogger.UnblockUserUsername[0] != "user1" || testActionsLogger.UnblockUserUsernameToUnblock[0] != "Anonymous" {
		t.Error("UnblockUser didn't correctly log action")
	}

	testActionsLogger.Reset()
	testModel.CreateChannel("channel1")
	if testActionsLogger.CreateChannelCalled != 1 || testActionsLogger.CreateChannelChannelname[0] != "channel1" {
		t.Error("CreateChannel didn't correctly log action")
	}

	testActionsLogger.Reset()
	testModel.DeleteChannel("channel1")
	if testActionsLogger.DeleteChannelCalled != 1 || testActionsLogger.DeleteChannelChannelname[0] != "channel1" {
		t.Error("DeleteChannel didn't correctly log action")
	}

	testModel.CreateChannel("channel1")
	testActionsLogger.Reset()
	timestamp := time.Now()
	testModel.PostMessage("channel1", "user1", timestamp, "message1")
	if testActionsLogger.PostMessageCalled != 1 || testActionsLogger.PostMessageChannelname[0] != "channel1" ||
		testActionsLogger.PostMessageUsername[0] != "user1" || testActionsLogger.PostMessageTimestamp[0] != timestamp ||
		testActionsLogger.PostMessageText[0] != "message1" {
		t.Error("PostMessage didn't correctly log action")
	}
}
