package main_test

import (
	"bytes"
	"encoding/json"
	"gotestprj/notificator"
	"gotestprj/server"
	"gotestprj/store"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

var (
	pgCred   = "postgres://postgres:12345678@127.0.0.1:5431/testauthusers?sslmode=disable"
	smtpMail = "empty"
	smtpPass = "empty"
)

// call defer server.Close() on end
func setUpConnections() error {
	dbcfg := store.NewStoreConfig(
		pgCred,
		3,
		filepath.Join("store", "query_templates", "."),
	)

	if err := store.Connect(dbcfg); err != nil {
		return err
	}

	smtpCfg := notificator.NewSMTPConfig(
		smtpMail,
		smtpPass,
		"smtp.gmail.com",
		"587",
	)

	notificator.ConnectToSMTP(smtpCfg)

	go func() {
		srvCfg := server.NewServerConfig("5050")
		//we'll see an error if the server dont starts
		server.Start(srvCfg)
	}()
	time.Sleep(time.Second * 2)

	return nil
}

func TestGetToken(t *testing.T) {

	if err := setUpConnections(); err != nil {
		t.Fatal(err)
	}

	cl := &http.Client{}

	u := store.NewUser(nil, "testuser", "testpass", nil)
	guid, err := store.CreateUser(u)
	if err != nil {
		t.Error("create user err ", err)
	}
	t.Log(guid)

	resp, err := cl.Get("http://127.0.0.1:5050/gettoken/?GUID=" + guid)
	if err != nil {
		t.Fatal("get keys err ", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("read response err ", err)
	}
	t.Log(string(body))

	keyPair := &server.KeyPair{}
	err = json.Unmarshal(body, keyPair)
	if err != nil {
		t.Fatal("unmarshal response err ", err)
	}
	t.Log("keys:", keyPair)

	if !(strings.Compare(keyPair.Acces, "") != 0 &&
		strings.Compare(keyPair.Refresh, "") != 0) {
		t.Fatal("empty one or both keys")
	}
}

func TestRefreshToken(t *testing.T) {

	if err := setUpConnections(); err != nil {
		t.Fatal(err)
	}

	cl := &http.Client{}

	u := store.NewUser(nil, "testuser", "testpass", nil)
	guid, err := store.CreateUser(u)
	if err != nil {
		t.Error("create user err ", err)
	}
	t.Log(guid)

	resp, err := cl.Get("http://127.0.0.1:5050/gettoken/?GUID=" + guid)
	if err != nil {
		t.Fatal("get keys err ", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("read response err ", err)
	}
	t.Log(string(body))

	keyPair := &server.KeyPair{}
	err = json.Unmarshal(body, keyPair)
	if err != nil {
		t.Fatal("unmarshal response err ", err)
	}
	t.Log("keys:", keyPair)

	resp, err = cl.Post(
		"http://localhost:5050/refreshtoken/",
		"application/json",
		bytes.NewReader(body),
	)
	if err != nil {
		t.Fatal("post send err ", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("read response err ", err)
	}

	newKeyPair := &server.KeyPair{}
	err = json.Unmarshal(body, newKeyPair)
	if err != nil {
		t.Fatal("create user err ", err)
	}
	t.Log("new keys:", newKeyPair)

	if strings.Compare(keyPair.Acces, newKeyPair.Acces) == 0 ||
		strings.Compare(keyPair.Refresh, newKeyPair.Refresh) == 0 {
		t.Fatal("new access or refresh key matches the previos one")
	}
}

func TestWrongRefreshToken(t *testing.T) {

	if err := setUpConnections(); err != nil {
		t.Fatal(err)
	}

	cl := &http.Client{}

	u := store.NewUser(nil, "testuser", "testpass", nil)
	guid, err := store.CreateUser(u)
	if err != nil {
		t.Error("create user err ", err)
	}
	t.Log(guid)

	resp, err := cl.Get("http://127.0.0.1:5050/gettoken/?GUID=" + guid)
	if err != nil {
		t.Fatal("get keys err ", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("read response err ", err)
	}
	t.Log(string(body))

	keyPair := &server.KeyPair{}
	err = json.Unmarshal(body, keyPair)
	if err != nil {
		t.Fatal("unmarshal response err ", err)
	}
	t.Log("keys:", keyPair)

	lenr := len(keyPair.Refresh)
	keyPair.Refresh = keyPair.Refresh[:lenr/2]

	t.Log("changed refresh key:", keyPair)

	changedPiarString, _ := json.Marshal(keyPair)

	resp, err = cl.Post(
		"http://localhost:5050/refreshtoken/",
		"application/json",
		bytes.NewReader(changedPiarString),
	)
	if err != nil {
		t.Fatal("post send err ", err)
	}

	body, err = io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal("read response err ", err)
	}
	t.Log("body:", string(body))

	newKeyPair := &server.KeyPair{}
	err = json.Unmarshal(body, newKeyPair)
	if err == nil {

		t.Fatal("server sended some keys:", newKeyPair)
	}
}
