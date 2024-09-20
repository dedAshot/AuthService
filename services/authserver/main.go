package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"gotestprj/notificator"
	"gotestprj/server"
	"gotestprj/store"
	"io"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

var (
	loggerWriter        *bufio.Writer
	loggerFlushInterval = time.Millisecond * 500
	stopper             = make(chan bool)
	sigChan             = make(chan os.Signal, 1)
)

func setupLogger() {
	dataTime := time.Now().UTC().Format("2006-01-02T15-04-05")
	fileName := "./logs/log" + dataTime + ".log"

	fd, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Can't create log file", err)
		os.Exit(0)
	}

	wr := io.MultiWriter(fd, os.Stdout)

	loggerWriter = bufio.NewWriter(wr)

	go func() {

		for {
			time.Sleep(loggerFlushInterval)
			loggerWriter.Flush()
		}
	}()

	logger := slog.New(slog.NewTextHandler(loggerWriter, nil))
	slog.SetDefault(logger)

	slog.Info("Logger setup complited")
}

func ckeckFlags() {

	devEnv := flag.Bool("dev", false, "Load developing enviroment variables (for testing purposes)")
	flag.Parse()

	if *devEnv {
		err := godotenv.Load()
		if err != nil {
			slog.Error("Error loading .env file")
			os.Exit(0)
		}
	}

	slog.Info("flag ckeck complited")
}

func serviceCloser() {
	signal.Notify(
		sigChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT,
	)

	go func() {
		sig := <-sigChan
		slog.Warn("process kill signal captured", "signal", sig)

		server.Close()
		store.Close()
		loggerWriter.Flush()

		stopper <- true
	}()
}

func init() {
	setupLogger()

}

//filepath abs

func main() {
	defer loggerWriter.Flush()
	slog.Info("App started")

	ckeckFlags()
	serviceCloser()
	setUpConnections()

	if err := server.Start(getServerConfig()); err != nil {
		slog.Error("server start error", "err", err)
	}

	//start close procedure
	sigChan <- os.Kill
	<-stopper

	slog.Info("App process ended")
}

func setUpConnections() {
	slog.Info("establishing connections")

	if err := store.Connect(getDbConfig()); err != nil {
		slog.Error("Could't connect to DB", "err", err)
		panic(1)
	}
	slog.Info("connected to DB")

	notificator.ConnectToSMTP(getSMTPConfig())
	slog.Info("connected to SMTP")

	slog.Info("connections are established")
}

func getDbConfig() *store.StoreConfig {

	dbCredentials := os.Getenv("DB_CREDENTIALS")
	if strings.Compare(dbCredentials, "") == 0 {
		dbCredentialsPath := os.Getenv("DB_CREDENTIALS_FILE")
		if strings.Compare(dbCredentialsPath, "") == 0 {
			slog.Error("no DB_CREDENTIALS or DB_CREDENTIALS_FILE supplied in ENV")
			panic(1)
		}

		var err error
		dbCredentials, err = readSecret(dbCredentialsPath)
		if err != nil {
			slog.Error("an error occured during DB_CREDENTIALS_FILE reading", "err", err.Error())
			panic(1)
		}
	}

	retriesCount := 0
	dbRetriesCountStr := os.Getenv("DB_RETRIES_COUNT")
	if strings.Compare(dbRetriesCountStr, "") == 0 {
		retriesCount = 5
	} else {
		var err error
		retriesCount, err = strconv.Atoi(dbRetriesCountStr)
		if err != nil {
			slog.Warn("ENV DB_RETRIES_COUNT should be int, set count = 5")
			retriesCount = 5
		}
	}
	path := filepath.Join("store", "query_templates", ".")
	return store.NewStoreConfig(dbCredentials, retriesCount, path)
}

func getSMTPConfig() *notificator.SMTPConfig {

	email := os.Getenv("SMTP_EMAIL")
	if strings.Compare(email, "") == 0 {
		emailPath := os.Getenv("SMTP_EMAIL_FILE")
		if strings.Compare(emailPath, "") == 0 {
			slog.Error("no SMTP_EMAIL or SMTP_EMAIL_FILE supplied in ENV")
			panic(1)
		}
		var err error
		email, err = readSecret(emailPath)
		if err != nil {
			slog.Error("an error occured during SMTP_EMAIL_FILE reading", "err", err.Error())
			panic(1)
		}
	}

	password := os.Getenv("SMTP_PASSWORD")
	if strings.Compare(password, "") == 0 {
		passwordPath := os.Getenv("SMTP_PASSWORD_FILE")
		if strings.Compare(passwordPath, "") == 0 {
			slog.Error("no SMTP_PASSWORD or SMTP_PASSWORD_FILE supplied in ENV")
			panic(1)
		}

		var err error
		password, err = readSecret(passwordPath)
		if err != nil {
			slog.Error("an error occured during SMTP_PASSWORD_FILE reading", "err", err.Error())
			panic(1)
		}
	}

	host := os.Getenv("SMTP_HOST")
	if strings.Compare(host, "") == 0 {
		slog.Error("no SMTP_HOST supplied in ENV")
		panic(1)
	}

	port := os.Getenv("SMTP_PORT")
	if strings.Compare(port, "") == 0 {
		slog.Error("no SMTP_PORT supplied in ENV")
		panic(1)
	}

	return notificator.NewSMTPConfig(email, password, host, port)
}

func getServerConfig() *server.ServerConfig {

	port := "8080"
	portEnv := os.Getenv("SERVER_PORT")
	if strings.Compare(portEnv, "") != 0 {
		port = portEnv
	}

	return server.NewServerConfig(port)
}

func readSecret(path string) (string, error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		return "", nil
	}

	scan := bufio.NewScanner(bytes.NewReader(buf))
	txt := ""
	for scan.Scan() {
		txt += scan.Text()
	}

	return txt, nil
}
