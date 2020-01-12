package actions_test

import (
	"chatserver/model/actions"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

type CreateUserAction struct {
	Username string
}

type DeleteUserAction struct {
	Username string
}

type BlockUserAction struct {
	Username        string
	UsernameToBlock string
}

type UnblockUserAction struct {
	Username          string
	UsernameToUnblock string
}

type CreateChannelAction struct {
	Channelname string
}

type DeleteChannelAction struct {
	Channelname string
}

type PostMessageAction struct {
	Channelname string
	Username    string
	Timestamp   time.Time
	Text        string
}

type TestActor struct {
	Actions []interface{}
}

func NewTestActor() *TestActor {
	testActor := TestActor{}
	testActor.Reset()
	return &testActor
}

func (t *TestActor) Reset() {
	t.Actions = make([]interface{}, 0)
}

func (t *TestActor) CreateUser(username string) {
	action := CreateUserAction{
		Username: username,
	}

	t.Actions = append(t.Actions, action)
}

func (t *TestActor) DeleteUser(username string) {
	action := DeleteUserAction{
		Username: username,
	}

	t.Actions = append(t.Actions, action)
}

func (t *TestActor) BlockUser(username string, usernameToBlock string) {
	action := BlockUserAction{
		Username:        username,
		UsernameToBlock: usernameToBlock,
	}

	t.Actions = append(t.Actions, action)
}

func (t *TestActor) UnblockUser(username string, usernameToUnblock string) {
	action := UnblockUserAction{
		Username:          username,
		UsernameToUnblock: usernameToUnblock,
	}

	t.Actions = append(t.Actions, action)
}

func (t *TestActor) CreateChannel(channelname string) {
	action := CreateChannelAction{
		Channelname: channelname,
	}

	t.Actions = append(t.Actions, action)
}

func (t *TestActor) DeleteChannel(channelname string) {
	action := DeleteChannelAction{
		Channelname: channelname,
	}

	t.Actions = append(t.Actions, action)
}

func (t *TestActor) PostMessage(channelname string, username string, timestamp time.Time, text string) {
	action := PostMessageAction{
		Channelname: channelname,
		Username:    username,
		Timestamp:   timestamp,
		Text:        text,
	}

	t.Actions = append(t.Actions, action)
}

func TestLoggerReplayerIntegrationTest(t *testing.T) {
	// NOTE: we shouldn't be doing file I/O in the unit test
	tempFile, err := ioutil.TempFile("", "test.*.txt")
	if err != nil {
		t.Error("Couldn't create temp file")
	}

	defer os.Remove(tempFile.Name())

	logFilePath := tempFile.Name()

	// Create the logger
	logger, err := actions.NewLogger(logFilePath)
	if err != nil {
		t.Error("Failed to create Logger")
	}

	// Log some actions
	logger.BlockUser("user1", "Anonymous")
	logger.CreateUser("user1")
	logger.CreateUser("user2")
	logger.CreateChannel("channel1")
	logger.DeleteChannel("channel1")
	logger.DeleteUser("user1")
	timestamp := time.Now()
	logger.PostMessage("General", "Anonymous", timestamp, "message1")
	logger.UnblockUser("user1", "Anonymous")
	logger.CreateUser("user3")

	// Create the replayer
	replayer, err := actions.NewReplayer(logFilePath)
	if err != nil {
		t.Error("Failed to create Replayer")
	}

	testActor := NewTestActor()

	// Replay the log
	err = replayer.Replay(testActor)
	if err != nil {
		t.Error(err)
	}

	action0 := testActor.Actions[0].(BlockUserAction)
	if action0.Username != "user1" || action0.UsernameToBlock != "Anonymous" {
		t.Error("Failed to replay BlockUser action")
	}

	action1 := testActor.Actions[1].(CreateUserAction)
	if action1.Username != "user1" {
		t.Error("Failed to replay CreateUser action")
	}

	action2 := testActor.Actions[2].(CreateUserAction)
	if action2.Username != "user2" {
		t.Error("Failed to replay CreateUser action")
	}

	action3 := testActor.Actions[3].(CreateChannelAction)
	if action3.Channelname != "channel1" {
		t.Error("Failed to replay CreateChannel action")
	}

	action4 := testActor.Actions[4].(DeleteChannelAction)
	if action4.Channelname != "channel1" {
		t.Error("Failed to replay DeleteChannel action")
	}

	action5 := testActor.Actions[5].(DeleteUserAction)
	if action5.Username != "user1" {
		t.Error("Failed to replay DeleteUser action")
	}

	action6 := testActor.Actions[6].(PostMessageAction)
	expectedTimestamp := timestamp.Format(time.RFC3339)
	action6Timestamp := action6.Timestamp.Format(time.RFC3339)
	if action6.Channelname != "General" || action6.Username != "Anonymous" || action6Timestamp != expectedTimestamp || action6.Text != "message1" {
		t.Error("Failed to replay PostMessage action")
	}

	action7 := testActor.Actions[7].(UnblockUserAction)
	if action7.Username != "user1" || action7.UsernameToUnblock != "Anonymous" {
		t.Error("Failed to replay UnblockUser action")
	}

	action8 := testActor.Actions[8].(CreateUserAction)
	if action8.Username != "user3" {
		t.Error("Failed to replay CreateUser action")
	}
}
