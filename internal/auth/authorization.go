package auth

import (
	"errors"
	"filthy/internal/constants"
	"github.com/gorilla/websocket"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func Authorize() (*websocket.Conn, string, error) {
	key := constants.SETTINGS.AuthKey
	u := url.URL{Scheme: "wss", Host: "filthyaio.online", Path: "/key/" + key}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, "", err
	}

	_, message, err := conn.ReadMessage()
	if err != nil {
		return nil, "", err
	}

	//fmt.Printf("Received message: %s\n", message)

	decrypted, err := GetAESDecrypted(string(message))

	status, name := checkDeadline(string(decrypted))
	if status != true {
		return nil, "", errors.New("time.Now > deadline")
	}

	return conn, name, nil

}

func checkDeadline(str string) (bool, string) {
	parts := strings.Split(str, ":")
	if len(parts) != 4 {
		return false, ""
	}

	deadline, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false, ""
	}

	return time.Now().Unix() < deadline, parts[2]
}
