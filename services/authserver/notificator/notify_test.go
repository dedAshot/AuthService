package notificator_test

import (
	"gotestprj/notificator"
	"testing"
)

func TestSendMsg(t *testing.T) {

	cfg := notificator.NewSMTPConfig(
		"",
		"",
		"smtp.gmail.com",
		"587",
	)

	notificator.ConnectToSMTP(cfg)

	resiverMail := ""
	err := notificator.SendMsg(resiverMail, "Hi, "+resiverMail, "Test msg")
	if err != nil {
		t.Fatal(err)
	}
}
