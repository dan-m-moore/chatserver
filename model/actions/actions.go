// Package actions provides model changing actions (Actor) that can be persisted (Logger) and
// replayed (Replayer).
package actions

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Actor provides an interface for responding to model actions.
type Actor interface {
	CreateUser(username string)
	DeleteUser(username string)
	BlockUser(username string, usernameToBlock string)
	UnblockUser(username string, usernameToUnblock string)
	CreateChannel(channelname string)
	DeleteChannel(channelname string)
	PostMessage(channelname string, username string, timestamp time.Time, text string)
}

// Action contains information about an action.
type Action struct {
	Name      string
	Timestamp time.Time
}

// CreateUserAction contains information about a CreateUser action.
type CreateUserAction struct {
	Action   Action `json:"Action"`
	Username string
}

// DeleteUserAction contains information about a DeleteUser action.
type DeleteUserAction struct {
	Action   Action `json:"Action"`
	Username string
}

// BlockUserAction contains information about a BlockUser action.
type BlockUserAction struct {
	Action          Action `json:"Action"`
	Username        string
	UsernameToBlock string
}

// UnblockUserAction contains information about a UnblockUser action.
type UnblockUserAction struct {
	Action            Action `json:"Action"`
	Username          string
	UsernameToUnblock string
}

// CreateChannelAction contains information about a CreateChannel action.
type CreateChannelAction struct {
	Action      Action `json:"Action"`
	Channelname string
}

// DeleteChannelAction contains information about a DeleteChannel action.
type DeleteChannelAction struct {
	Action      Action `json:"Action"`
	Channelname string
}

// PostMessageAction contains information about a PostMessage action.
type PostMessageAction struct {
	Action      Action `json:"Action"`
	Channelname string
	Username    string
	Timestamp   time.Time
	Text        string
}

// Logger provides a means to log model actions to a file.  It provides the Actor interface
// and will persist the actions sequentially.
type Logger struct {
	logFilePath string
}

// NewLogger creates/initializes/returns a new Logger.
func NewLogger(logFilePath string) (*Logger, error) {
	// Validate the path
	if logFilePath == "" {
		return nil, errors.New("invalid log file path")
	}

	info, err := os.Stat(logFilePath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, err
		}
	} else {
		if info.IsDir() {
			return nil, errors.New("log file path points to a directory")
		}
	}

	// If the file doesn't exist or is empty, create/initialize it
	info, err = os.Stat(logFilePath)
	if os.IsNotExist(err) || info.Size() == 0 {
		// Create the directory if it doesn't exist
		dir := filepath.Dir(logFilePath)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			return nil, err
		}

		logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_RDWR, 0644)
		if err != nil {
			return nil, err
		}

		// Write the array brackets to the file
		_, err = logFile.WriteString("[\n{}\n]")
		if err != nil {
			return nil, err
		}

		err = logFile.Close()
		if err != nil {
			return nil, err
		}
	}

	// At this point, we have a valid log file
	logger := Logger{
		logFilePath: logFilePath,
	}

	return &logger, nil
}

// CreateUser logs the CreateUser action.
func (l *Logger) CreateUser(username string) {
	action := CreateUserAction{
		Action: Action{
			Name:      "CreateUser",
			Timestamp: time.Now(),
		},
		Username: username,
	}

	l.commitAction(&action)
}

// DeleteUser logs the DeleteUser action.
func (l *Logger) DeleteUser(username string) {
	action := DeleteUserAction{
		Action: Action{
			Name:      "DeleteUser",
			Timestamp: time.Now(),
		},
		Username: username,
	}

	l.commitAction(&action)
}

// BlockUser logs the BlockUser action.
func (l *Logger) BlockUser(username string, usernameToBlock string) {
	action := BlockUserAction{
		Action: Action{
			Name:      "BlockUser",
			Timestamp: time.Now(),
		},
		Username:        username,
		UsernameToBlock: usernameToBlock,
	}

	l.commitAction(&action)
}

// UnblockUser logs the UnblockUser action.
func (l *Logger) UnblockUser(username string, usernameToUnblock string) {
	action := UnblockUserAction{
		Action: Action{
			Name:      "UnblockUser",
			Timestamp: time.Now(),
		},
		Username:          username,
		UsernameToUnblock: usernameToUnblock,
	}

	l.commitAction(&action)
}

// CreateChannel logs the CreateChannel action.
func (l *Logger) CreateChannel(channelname string) {
	action := CreateChannelAction{
		Action: Action{
			Name:      "CreateChannel",
			Timestamp: time.Now(),
		},
		Channelname: channelname,
	}

	l.commitAction(&action)
}

// DeleteChannel logs the DeleteChannel action.
func (l *Logger) DeleteChannel(channelname string) {
	action := DeleteChannelAction{
		Action: Action{
			Name:      "DeleteChannel",
			Timestamp: time.Now(),
		},
		Channelname: channelname,
	}

	l.commitAction(&action)
}

// PostMessage logs the PostMessage action.
func (l *Logger) PostMessage(channelname string, username string, timestamp time.Time, text string) {
	action := PostMessageAction{
		Action: Action{
			Name:      "PostMessage",
			Timestamp: time.Now(),
		},
		Channelname: channelname,
		Username:    username,
		Timestamp:   timestamp,
		Text:        text,
	}

	l.commitAction(&action)
}

func (l *Logger) commitAction(action interface{}) {
	// Marshal the JSON
	jsonAction, err := json.Marshal(action)
	if err != nil {
		log.Fatal(err)
	}

	logFile, err := os.OpenFile(l.logFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}

	// Seek to the end of the file minus 2 bytes (to overwrite the last entry's trailing newline)
	_, err = logFile.Seek(-2, 2)
	if err != nil {
		log.Fatal(err)
	}

	// Write the action to the file
	_, err = logFile.WriteString(",\n" + string(jsonAction) + "\n]")
	if err != nil {
		log.Fatal(err)
	}

	// Close the file
	err = logFile.Close()
	if err != nil {
		log.Fatal(err)
	}
}

// Replayer provides a means to replay model actions sequentially that were written to a log file.
type Replayer struct {
	logFilePath string
	actor       Actor
}

// NewReplayer creates/initializes/returns a new Replayer.
func NewReplayer(logFilePath string) (*Replayer, error) {
	// Validate the path
	if logFilePath == "" {
		return nil, errors.New("invalid log file path")
	}

	// Validate the log file
	info, err := os.Stat(logFilePath)
	if err != nil {
		return nil, err
	}

	if info.IsDir() {
		return nil, errors.New("log file path points to a directory")
	}

	if info.Size() == 0 {
		return nil, errors.New("log file is empty")
	}

	replayer := Replayer{
		logFilePath: logFilePath,
		actor:       nil,
	}

	return &replayer, nil
}

// Replay will replay the model actions sequentially on the Actor.
func (r *Replayer) Replay(actor Actor) error {
	r.actor = actor

	// Read the entire file
	wholeFile, err := ioutil.ReadFile(r.logFilePath)
	if err != nil {
		return err
	}

	// Parse the json string
	var result []map[string]interface{}
	err = json.Unmarshal(wholeFile, &result)
	if err != nil {
		return errors.New("invalid input log file - malformed json")
	}

	// Parse the action entries
	for _, action := range result {
		// Disregard empty entries
		if len(action) == 0 {
			continue
		}

		// Parse the individual action
		err = r.parseAction(&action)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Replayer) parseAction(action *map[string]interface{}) error {
	if _, ok := (*action)["Action"]; !ok {
		return errors.New("invalid input log file - action not found")
	}

	actionStruct := (*action)["Action"].(map[string]interface{})

	if _, ok := actionStruct["Name"]; !ok {
		return errors.New("invalid input log file - name not found")
	}

	actionName, ok := actionStruct["Name"].(string)
	if !ok {
		return errors.New("invalid input log file - name not string")
	}

	switch actionName {
	case "CreateUser":
		err := r.parseCreateUser(action)
		if err != nil {
			return err
		}
	case "DeleteUser":
		err := r.parseDeleteUser(action)
		if err != nil {
			return err
		}
	case "BlockUser":
		err := r.parseBlockUser(action)
		if err != nil {
			return err
		}
	case "UnblockUser":
		err := r.parseUnblockUser(action)
		if err != nil {
			return err
		}
	case "CreateChannel":
		err := r.parseCreateChannel(action)
		if err != nil {
			return err
		}
	case "DeleteChannel":
		err := r.parseDeleteChannel(action)
		if err != nil {
			return err
		}
	case "PostMessage":
		err := r.parsePostMessage(action)
		if err != nil {
			return err
		}
	default:
		return errors.New("invalid input log file - unknown action")
	}

	return nil
}

func (r *Replayer) parseCreateUser(action *map[string]interface{}) error {
	if _, ok := (*action)["Username"]; !ok {
		return errors.New("invalid input log file - CreateUser - missing Username")
	}
	username, ok := (*action)["Username"].(string)
	if !ok {
		return errors.New("invalid input log file - CreateUser - Username not a string")
	}

	r.actor.CreateUser(username)
	return nil
}

func (r *Replayer) parseDeleteUser(action *map[string]interface{}) error {
	if _, ok := (*action)["Username"]; !ok {
		return errors.New("invalid input log file - DeleteUser - missing Username")
	}
	username, ok := (*action)["Username"].(string)
	if !ok {
		return errors.New("invalid input log file - DeleteUser - Username not a string")
	}

	r.actor.DeleteUser(username)
	return nil
}

func (r *Replayer) parseBlockUser(action *map[string]interface{}) error {
	if _, ok := (*action)["Username"]; !ok {
		return errors.New("invalid input log file - BlockUser - missing Username")
	}
	username, ok := (*action)["Username"].(string)
	if !ok {
		return errors.New("invalid input log file - BlockUser - Username not a string")
	}

	if _, ok := (*action)["UsernameToBlock"]; !ok {
		return errors.New("invalid input log file - BlockUser - missing UsernameToBlock")
	}
	usernameToBlock, ok := (*action)["UsernameToBlock"].(string)
	if !ok {
		return errors.New("invalid input log file - BlockUser - UsernameToBlock not a string")
	}

	r.actor.BlockUser(username, usernameToBlock)
	return nil
}

func (r *Replayer) parseUnblockUser(action *map[string]interface{}) error {
	if _, ok := (*action)["Username"]; !ok {
		return errors.New("invalid input log file - UnblockUser - missing Username")
	}
	username, ok := (*action)["Username"].(string)
	if !ok {
		return errors.New("invalid input log file - UnblockUser - Username not a string")
	}

	if _, ok := (*action)["UsernameToUnblock"]; !ok {
		return errors.New("invalid input log file - UnblockUser - missing UsernameToUnblock")
	}
	usernameToUnblock, ok := (*action)["UsernameToUnblock"].(string)
	if !ok {
		return errors.New("invalid input log file - UnblockUser - UsernameToUnblock not a string")
	}

	r.actor.UnblockUser(username, usernameToUnblock)
	return nil
}

func (r *Replayer) parseCreateChannel(action *map[string]interface{}) error {
	if _, ok := (*action)["Channelname"]; !ok {
		return errors.New("invalid input log file - CreateChannel - missing Channelname")
	}
	channelname, ok := (*action)["Channelname"].(string)
	if !ok {
		return errors.New("invalid input log file - CreateChannel - Channelname not a string")
	}

	r.actor.CreateChannel(channelname)
	return nil
}

func (r *Replayer) parseDeleteChannel(action *map[string]interface{}) error {
	if _, ok := (*action)["Channelname"]; !ok {
		return errors.New("invalid input log file - DeleteChannel - missing Channelname")
	}
	channelname, ok := (*action)["Channelname"].(string)
	if !ok {
		return errors.New("invalid input log file - DeleteChannel - Channelname not a string")
	}

	r.actor.DeleteChannel(channelname)
	return nil
}

func (r *Replayer) parsePostMessage(action *map[string]interface{}) error {
	if _, ok := (*action)["Channelname"]; !ok {
		return errors.New("invalid input log file - PostMessage - missing Channelname")
	}
	channelname, ok := (*action)["Channelname"].(string)
	if !ok {
		return errors.New("invalid input log file - PostMessage - Channelname not a string")
	}

	if _, ok := (*action)["Username"]; !ok {
		return errors.New("invalid input log file - PostMessage - missing Username")
	}
	username, ok := (*action)["Username"].(string)
	if !ok {
		return errors.New("invalid input log file - PostMessage - Username not a string")
	}

	if _, ok := (*action)["Timestamp"]; !ok {
		return errors.New("invalid input log file - PostMessage - missing Timestamp")
	}
	timestampString, ok := (*action)["Timestamp"].(string)
	if !ok {
		return errors.New("invalid input log file - PostMessage - Timestamp not a string")
	}
	timestamp, err := time.Parse(time.RFC3339, timestampString)
	if err != nil {
		return err
	}

	if _, ok := (*action)["Text"]; !ok {
		return errors.New("invalid input log file - PostMessage - missing Text")
	}
	text, ok := (*action)["Text"].(string)
	if !ok {
		return errors.New("invalid input log file - PostMessage - Text not a string")
	}

	r.actor.PostMessage(channelname, username, timestamp, text)
	return nil
}
