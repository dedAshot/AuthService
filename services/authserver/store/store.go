package store

import (
	"bufio"
	"database/sql"
	"errors"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type StoreConfig struct {
	DbCredentials               string
	ConnectionRetriesCount      int
	PathToQuriesTemplatesFolder string
}

func NewStoreConfig(DbCredentials string,
	ConnectionRetriesCount int, PathToQuriesTemplatesFolder string) *StoreConfig {
	return &StoreConfig{
		DbCredentials:               DbCredentials,
		ConnectionRetriesCount:      ConnectionRetriesCount,
		PathToQuriesTemplatesFolder: PathToQuriesTemplatesFolder,
	}
}

var queryStorage = make(map[string]string)

var ErrEmptySQLQuery = errors.New("auth: token ip and client ip mismatch")

// "gotestprj/store/query_templates/"

// Store queries in local queryStorage[fileName]
func chaceQueries(PathToQuriesTemplatesFolder string) error {
	queryDirs, err := os.ReadDir(PathToQuriesTemplatesFolder)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	for _, entry := range queryDirs {
		name := entry.Name()
		templ, err := getQueryTemplFromFile(filepath.Join(PathToQuriesTemplatesFolder, name))
		if err != nil {
			slog.Error("queryStorage file "+name+" parse err:", "err", err.Error())
			return err
		}
		queryStorage[name] = templ
	}

	return nil
}

func getQueryTemplFromFile(fileAddr string) (string, error) {

	fd, err := os.Open(fileAddr)
	if err != nil {
		return "", err
	}
	scan := bufio.NewScanner(fd)
	query := ""

	for scan.Scan() {
		query += scan.Text() + "\n"
	}

	if strings.Compare(query, "") == 0 {
		return "", ErrEmptySQLQuery
	}

	return query, nil

}

var db *sql.DB

func Connect(config *StoreConfig) error {
	if db != nil {
		return nil
	}

	if err := chaceQueries(config.PathToQuriesTemplatesFolder); err != nil {
		return err
	}

	var err error = connectToDb(config)
	for retries := config.ConnectionRetriesCount; err != nil && retries > 0; retries-- {
		err = connectToDb(config)
		time.Sleep(time.Second)
	}

	if err != nil {
		return err
	}

	return nil
}

func connectToDb(config *StoreConfig) error {

	var err error
	if db, err = sql.Open("postgres", config.DbCredentials); err != nil {
		return err
	}

	if err = db.Ping(); err != nil {
		return err
	}

	return nil
}

func Close() error {
	err := db.Close()
	return err
}

type User struct {
	Guid         []byte
	Email        string
	Password     string
	ReftokenHash []byte
}

func NewUser(Guid []byte, Email string, Password string, ReftokenHash []byte) *User {
	return &User{Guid: Guid, Email: Email, Password: Password, ReftokenHash: ReftokenHash}
}

func CreateUser(u *User) (string, error) {
	query := queryStorage["insert_users.sql"]

	guid := ""
	row := db.QueryRow(query, u.Email, u.Password, u.ReftokenHash)
	err := row.Scan(&guid)
	if err != nil {
		return "", err
	}

	return guid, nil
}

func GetUserByGuid(guid string) (*User, error) {
	query := queryStorage["select_users_user.sql"]
	user := &User{}

	row := db.QueryRow(query, guid)

	err := row.Scan(&user.Guid, &user.Email, &user.Password, &user.ReftokenHash)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func SetUserRefreshTokenHash(reftokenHash []byte, guid string) error {
	query := queryStorage["update_users_refreshtoken.sql"]
	_, err := db.Exec(query, reftokenHash, guid)

	return err
}

func GetUserRefreshHashByGuid(guid string) ([]byte, error) {
	query := queryStorage["select_users_refreshtoken.sql"]

	row := db.QueryRow(query, guid)

	var hash []byte
	err := row.Scan(&hash)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

func GetUserEmailByGuid(guid string) (string, error) {
	query := queryStorage["select_users_email.sql"]

	row := db.QueryRow(query, guid)

	var email string
	err := row.Scan(&email)
	if err != nil {
		return "", err
	}

	return email, nil
}

// for testing purposes
// func Exec(query string) error {
// 	_, err := db.Exec(query)
// 	return err
// }
