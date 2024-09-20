package server

import (
	"encoding/json"
	"gotestprj/auth"
	"log/slog"
	"net/http"
	"strings"
)

type KeyPair struct {
	Acces   string `json:"access"`
	Refresh string `json:"refresh"`
}

func newKeyPair(acc, ref string) *KeyPair {
	return &KeyPair{Acces: acc, Refresh: ref}
}

func GetToken(w http.ResponseWriter, r *http.Request) {

	guid := r.URL.Query().Get("GUID")
	if strings.Compare(guid, "") == 0 {
		logError(w, "empty GUID", nil)
		return
	}

	ip := r.RemoteAddr

	acc, ref, err := auth.CreateTokens(ip, guid)
	if err != nil {
		logError(w, "tokens creation error", err)
		return
	}

	pair := newKeyPair(acc, ref)
	response, err := json.Marshal(pair)

	if err != nil {
		logError(w, "tokens marshaling error", err)
		return
	}

	w.Write(response)
	slog.Info("sended key pair to", "ip", ip)
}

func Refresh(w http.ResponseWriter, r *http.Request) {

	body := make([]byte, 0, 1024)
	buf := make([]byte, 1024)
	ip := r.RemoteAddr

	for n, err := r.Body.Read(buf); n > 0 || err == nil; n, err = r.Body.Read(buf) {
		body = append(body, buf[:n]...)
	}

	OldKeyPair := newKeyPair("", "")
	if err := json.Unmarshal(body, OldKeyPair); err != nil {
		logError(w, "tokens unmarshaling error", err)
		return
	}

	acc, ref, err := auth.RefreshToken(OldKeyPair.Acces, OldKeyPair.Refresh, ip)
	if err != nil {
		logError(w, "tokens creation error", err)
		return
	}

	pair := newKeyPair(acc, ref)
	response, err := json.Marshal(pair)

	if err != nil {
		logError(w, "tokens marshaling error", err)
		return
	}

	w.Write(response)
	slog.Info("refreshed key pair to", "ip", ip)
}

func logError(w http.ResponseWriter, msg string, err error) {
	slog.Error(msg, "err", err)
	w.Write([]byte(msg))
}
