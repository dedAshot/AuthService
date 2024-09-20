package store_test

//tests doesnt work becouse store init() function contains paths

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"gotestprj/auth"
	"gotestprj/store"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/lib/pq"
)

var (
	pgCred = "postgres://postgres:12345678@127.0.0.1:5431/testauthusers?sslmode=disable"
)

// call defer server.Close() on end
func setUpConnections() error {
	dbcfg := store.NewStoreConfig(
		pgCred,
		3,
		filepath.Join("query_templates"),
	)

	if err := store.Connect(dbcfg); err != nil {
		return err
	}

	return nil
}

func createTestUser() (string, *store.User, error) {
	_, refHash, _ := auth.GenerateRefreshToken()

	user := store.NewUser(nil, "testmail", "password", refHash)

	guid, err := store.CreateUser(user)
	return guid, user, err
}

func TestCreateUser(t *testing.T) {
	setUpConnections()

	_, refHash, _ := auth.GenerateRefreshToken()

	user := store.NewUser(nil, "testmail", "password", refHash)

	_, err := store.CreateUser(user)
	if err != nil {
		t.Error("error creating user: " + err.Error())
		t.FailNow()
	}
}

func TestGetUserByGuid(t *testing.T) {
	setUpConnections()

	_, refHash, _ := auth.GenerateRefreshToken()

	user := store.NewUser(nil, "testmail", "password", refHash)

	guid, err := store.CreateUser(user)
	if err != nil {
		t.Error("error creating user: " + err.Error())
		t.FailNow()
	}

	u, err := store.GetUserByGuid(guid)
	if err != nil {
		t.Error("error get user: " + err.Error())
		t.FailNow()
	}

	if !(strings.Compare(u.Email, user.Email) == 0 &&
		strings.Compare(u.Password, user.Password) == 0 &&
		strings.Compare(string(u.ReftokenHash), string(u.ReftokenHash)) == 0) {
		t.Error("inserted and getted user mismatched")
	}
}

func TestSetUserRefreshTokenHash(t *testing.T) {
	setUpConnections()

	guid, _, err := createTestUser()
	if err != nil {
		t.Fatal(err)
	}

	_, hash, _ := auth.GenerateRefreshToken()

	err = store.SetUserRefreshTokenHash(hash, guid)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetUserRefreshHashByGuid(t *testing.T) {
	setUpConnections()

	guid, u, err := createTestUser()
	if err != nil {
		t.Fatal(err)
	}

	hash, err := store.GetUserRefreshHashByGuid(guid)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(hash, u.ReftokenHash) {
		t.Fatal("hashes arent equal")
	}
}

func TestGetUserEmailByGuid(t *testing.T) {
	setUpConnections()

	guid, u, err := createRandomUser()
	if err != nil {
		t.Fatal(err)
	}

	email, err := store.GetUserEmailByGuid(guid)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.EqualFold(email, u.Email) {
		t.Fatal("different emails")
	}
}

func createRandomUser() (string, *store.User, error) {
	_, refHash, _ := auth.GenerateRefreshToken()

	randMail := make([]byte, 20)
	randPassword := make([]byte, 20)
	rand.Read(randMail)
	rand.Read(randPassword)
	mail := hex.EncodeToString(randMail)
	password := hex.EncodeToString(randPassword)

	user := store.NewUser(nil, mail, password, refHash)

	guid, err := store.CreateUser(user)
	return guid, user, err
}
