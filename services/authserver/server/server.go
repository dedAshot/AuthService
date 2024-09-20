package server

import (
	"crypto/tls"
	"net/http"

	"log/slog"
)

var server *http.Server

type ServerConfig struct {
	port string
}

func NewServerConfig(port string) *ServerConfig {
	return &ServerConfig{port: port}
}

func init() {
	setupServer()
}

func setupServer() {

	server = &http.Server{
		Handler: http.DefaultServeMux,
		TLSConfig: &tls.Config{
			MinVersion: tls.VersionTLS13,
		},
	}

	http.HandleFunc("GET /gettoken/{$}", GetToken)
	http.HandleFunc("POST /refreshtoken/{$}", Refresh)
}

func Start(cfg *ServerConfig) error {
	slog.Info("Server stared")

	server.Addr = ":" + cfg.port

	return server.ListenAndServe()
}

func Close() {
	slog.Info("server closed")
	server.Close()
}
