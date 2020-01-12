package subs_test

import (
	"chatserver/model/subs"
	"errors"
	"testing"
	"time"
)

type TestClient struct {
	OnUsersChangedChan          chan int
	OnUserChangedChan           chan string
	OnUserChangedUsername       []string
	OnChannelsChangedChan       chan int
	OnChannelChangedChan        chan string
	OnChannelChangedChannelname []string
}

func NewTestClient() *TestClient {
	testClient := TestClient{}
	testClient.Reset()

	return &testClient
}

func (t *TestClient) Reset() {
	t.OnUsersChangedChan = make(chan int, 1)
	t.OnUserChangedChan = make(chan string, 1)
	t.OnUserChangedUsername = make([]string, 0)
	t.OnChannelsChangedChan = make(chan int, 1)
	t.OnChannelChangedChan = make(chan string, 1)
	t.OnChannelChangedChannelname = make([]string, 0)
}

func (t *TestClient) WaitForOnUsersChanged() error {
	select {
	case <-t.OnUsersChangedChan:
		return nil
	case <-time.After(25 * time.Millisecond):
		return errors.New("Timed out waiting for OnUsersChanged")
	}
}

func (t *TestClient) WaitForOnUserChanged() error {
	select {
	case username := <-t.OnUserChangedChan:
		t.OnUserChangedUsername = append(t.OnUserChangedUsername, username)
		return nil
	case <-time.After(25 * time.Millisecond):
		return errors.New("Timed out waiting for OnUserChanged")
	}
}

func (t *TestClient) WaitForOnChannelsChanged() error {
	select {
	case <-t.OnChannelsChangedChan:
		return nil
	case <-time.After(25 * time.Millisecond):
		return errors.New("Timed out waiting for OnChannelsChanged")
	}
}

func (t *TestClient) WaitForOnChannelChanged() error {
	select {
	case channelname := <-t.OnChannelChangedChan:
		t.OnChannelChangedChannelname = append(t.OnChannelChangedChannelname, channelname)
		return nil
	case <-time.After(25 * time.Millisecond):
		return errors.New("Timed out waiting for OnChannelChanged")
	}
}

func (t *TestClient) OnUsersChanged() {
	t.OnUsersChangedChan <- 0
}

func (t *TestClient) OnUserChanged(username string) {
	t.OnUserChangedChan <- username
}

func (t *TestClient) OnChannelsChanged() {
	t.OnChannelsChangedChan <- 0
}

func (t *TestClient) OnChannelChanged(channelname string) {
	t.OnChannelChangedChan <- channelname
}

func TestConnectAndDisconnect(t *testing.T) {
	testClient := NewTestClient()
	engine := subs.NewEngine()
	err := engine.Connect(testClient)
	if err != nil {
		t.Error("Connect failed")
	}

	err = engine.Connect(testClient)
	if err == nil {
		t.Error("Double connect didn't fail")
	}

	err = engine.Disconnect(testClient)
	if err != nil {
		t.Error("Disconnect failed")
	}

	err = engine.Disconnect(testClient)
	if err == nil {
		t.Error("Double disconntect didn't fail")
	}
}

func TestMultiClient(t *testing.T) {
	testClient1 := NewTestClient()
	testClient2 := NewTestClient()

	engine := subs.NewEngine()

	engine.Connect(testClient1)
	engine.Connect(testClient2)

	engine.UsersChanged()
	err := testClient1.WaitForOnUsersChanged()
	if err != nil {
		t.Error(err)
	}

	err = testClient2.WaitForOnUsersChanged()
	if err != nil {
		t.Error(err)
	}

	engine.UserChanged("user1")
	err = testClient1.WaitForOnUserChanged()
	if err != nil {
		t.Error(err)
	}
	if len(testClient1.OnUserChangedUsername) != 1 || testClient1.OnUserChangedUsername[0] != "user1" {
		t.Error("Incorrect username provided to OnUserChanged")
	}

	err = testClient2.WaitForOnUserChanged()
	if err != nil {
		t.Error(err)
	}
	if len(testClient2.OnUserChangedUsername) != 1 || testClient2.OnUserChangedUsername[0] != "user1" {
		t.Error("Incorrect username provided to OnUserChanged")
	}

	engine.ChannelsChanged()
	err = testClient1.WaitForOnChannelsChanged()
	if err != nil {
		t.Error(err)
	}

	err = testClient2.WaitForOnChannelsChanged()
	if err != nil {
		t.Error(err)
	}

	engine.ChannelChanged("channel1")
	err = testClient1.WaitForOnChannelChanged()
	if err != nil {
		t.Error(err)
	}
	if len(testClient1.OnChannelChangedChannelname) != 1 || testClient1.OnChannelChangedChannelname[0] != "channel1" {
		t.Error("Incorrect channelname provided to OnChannelChanged")
	}

	err = testClient2.WaitForOnChannelChanged()
	if err != nil {
		t.Error(err)
	}
	if len(testClient2.OnChannelChangedChannelname) != 1 || testClient2.OnChannelChangedChannelname[0] != "channel1" {
		t.Error("Incorrect channelname provided to OnChannelChanged")
	}

	engine.Disconnect(testClient2)

	engine.UsersChanged()
	err = testClient1.WaitForOnUsersChanged()
	if err != nil {
		t.Error(err)
	}

	err = testClient2.WaitForOnUsersChanged()
	if err == nil {
		t.Error("Got UsersChanged call after disconnecting")
	}

	engine.UserChanged("user1")
	err = testClient1.WaitForOnUserChanged()
	if err != nil {
		t.Error(err)
	}

	err = testClient2.WaitForOnUserChanged()
	if err == nil {
		t.Error("Got UserChanged call after disconnecting")
	}

	engine.ChannelsChanged()
	err = testClient1.WaitForOnChannelsChanged()
	if err != nil {
		t.Error(err)
	}

	err = testClient2.WaitForOnChannelsChanged()
	if err == nil {
		t.Error("Got ChannelsChanged call after disconnecting")
	}

	engine.ChannelChanged("channel1")
	err = testClient1.WaitForOnChannelChanged()
	if err != nil {
		t.Error(err)
	}

	err = testClient2.WaitForOnChannelChanged()
	if err == nil {
		t.Error("Got ChannelChanged call after disconnecting")
	}
}
