package xhandler

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

type AuthFn func(string, string) (string, <-chan []byte, error)

const (
	MethodAuth         = "auth"
	MethodAuthExpiring = "authExpiring"
	MethodExecutions   = "executions"
)

var (
	AuthTimeoutError = errors.New("Auth timeout")
	AuthFailedError  = errors.New("Auth failed")
)

func HandleEvent(auth AuthFn, data []byte) error {
	var msg map[string]interface{}
	err := json.Unmarshal(data, &msg)
	if err != nil {
		return err
	}

	method, ok := msg["method"]
	if ok {
		switch method {
		case MethodAuthExpiring:
			logrus.Info("Auth Expiring")
			handleAuth(auth)
		case MethodExecutions:
			handleMethodExecutions(msg)
		}
	} else {
		logrus.WithField("payload", msg).Warn("dont know how to handle msg")
	}

	return nil
}

func handleAuth(auth AuthFn) error {
	reqId, read, err := auth("foo", "bar")
	if err != nil {
		return err
	}

	for {
		select {
		case <-time.After(1 * time.Second):
			logrus.Warn("Auth timeout")
			return AuthTimeoutError

		case data := <-read:
			var msg map[string]interface{}

			err := json.Unmarshal(data, &msg)
			if err != nil {
				return err
			}

			matchOnId := msg["req_id"] == reqId
			isAuthSuccess := msg["status"].(bool)

			if matchOnId {
				if isAuthSuccess {
					logrus.Info("Auth success")
					return nil
				} else {
					logrus.Info("Auth fail")
					return AuthFailedError
				}
			}
		}
	}
}

func handleMethodExecutions(msg map[string]interface{}) {
	data := msg["data"].(map[string]interface{})
	s := fmt.Sprintf("symbol: %s price: %s amount: %s",
		data["symbol"], data["price"], data["amount"])

	logrus.WithField("payload", s).Info("got new message")
}
