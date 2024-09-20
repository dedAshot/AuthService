package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"gotestprj/notificator"
	"gotestprj/store"
	"log/slog"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type claims struct {
	Ip string `json:"ip"`
	jwt.RegisteredClaims
}

var ErrIpMismatch = errors.New("auth: token ip and client ip mismatch")
var ErrClaimsMismatch = errors.New("auth: token claims mismatch")
var ErrHashMismatch = errors.New("auth: stored hash and provided hash mismatch")

const bcryptLevel = 15

func CreateTokens(ip, guid string) (acessStr string, refreshStr string, err error) {

	refresh, hash, err := generateAndStoreRefreshToken(guid)
	if err != nil {
		return "", "", err
	}

	refreshStr = base64.RawStdEncoding.EncodeToString(refresh)

	//refresh token hash is used as a signature of access token
	acessStr, err = generateAcessToken(ip, guid, hash)
	if err != nil {
		return "", "", err
	}

	return acessStr, refreshStr, nil
}

func generateAndStoreRefreshToken(guid string) (refRaw, hash []byte, err error) {

	refresh, hash, err := GenerateRefreshToken()
	if err != nil {
		return nil, nil, err
	}

	if err = store.SetUserRefreshTokenHash(hash, guid); err != nil {
		return nil, nil, err
	}

	return refresh, hash, nil
}

func GenerateRefreshToken() (refRaw, hash []byte, err error) {

	refRaw = make([]byte, 72)

	rand.Read(refRaw)

	hash, err = bcrypt.GenerateFromPassword(refRaw, bcryptLevel)
	if err != nil {
		return nil, nil, err
	}

	return refRaw, hash, nil
}

func generateAcessToken(ip, guid string, signature []byte) (string, error) {

	claimsAcc := claims{
		ip,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			Subject:   guid,
		},
	}

	return jwt.NewWithClaims(jwt.SigningMethodHS512, claimsAcc).SignedString(signature)
}

func RefreshToken(acStr, refStr, newIp string) (acessStr string, refreshStr string, err error) {

	token, err := jwt.ParseWithClaims(acStr, &claims{}, checkAcessTokenSignature)
	if err != nil {
		return "", "", err
	}

	if err := verifyAccessToken(token, newIp); err != nil {
		return "", "", err
	}

	guid, err := token.Claims.GetSubject()
	if err != nil {
		return "", "", nil
	}

	refRaw, err := base64.RawStdEncoding.DecodeString(refStr)
	if err != nil {
		return "", "", err
	}

	err = verifyRefreshToken(refRaw, guid)
	if err != nil {
		return "", "", err
	}

	return CreateTokens(newIp, guid)
}

func verifyRefreshToken(refRaw []byte, guid string) error {

	storedHash, err := store.GetUserRefreshHashByGuid(guid)
	if err != nil {
		return err
	}

	if err = bcrypt.CompareHashAndPassword(storedHash, refRaw); err != nil {
		return ErrHashMismatch
	}

	return nil
}

// jwt.Keyfunc implementation, returns key for access token
func checkAcessTokenSignature(token *jwt.Token) (interface{}, error) {

	guid, err := token.Claims.GetSubject()
	if err != nil {
		return nil, err
	}

	return store.GetUserRefreshHashByGuid(guid)
}

func verifyAccessToken(token *jwt.Token, ip string) error {

	claims, ok := token.Claims.(*claims)
	if !ok {
		return ErrClaimsMismatch
	}
	if err := checkAccessTokenIp(claims.Ip, ip); err != nil {
		guid := claims.Subject
		slog.Warn("Client ip addres mismatchs acces token ip")
		notifyAccountOwner(guid, ip)
	}

	return nil
}

func notifyAccountOwner(guid string, ip string) error {

	email, err := store.GetUserEmailByGuid(guid)
	if err != nil {
		return err
	}

	msg := fmt.Sprintf("Warning: You have a new ip address [%s], don't you?", ip)

	err = notificator.SendMsg(email, msg)
	if err != nil {
		return err
	}

	return nil
}

func checkAccessTokenIp(tokenIp, ip string) error {
	if strings.Compare(tokenIp, ip) != 0 {
		return ErrIpMismatch
	}
	return nil
}
