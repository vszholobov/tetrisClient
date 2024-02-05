package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mattn/go-tty"
)

const showCursorASCII = "\033[?25h"
const hideCursorASCII = "\033[?25l"

type CreateSessionResponse struct {
	SessionId int64 `json:"sessionId"`
}

type SessionDto struct {
	SessionId int64 `json:"sessionId"`
	Started   bool  `json:"started"`
}

var addr = "84.201.177.35:8080"

// https://github.com/gorilla/websocket/blob/main/examples/echo/server.go
func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	operation := os.Args[1]

	var sessionId string
	if operation == "connect" {
		sessionId = os.Args[2]
	} else if operation == "create" {
		response, createSessionError := http.Get("http://" + addr + "/session/create")
		if createSessionError != nil {
			panic(createSessionError.Error())
		}
		body, readResponseError := ioutil.ReadAll(response.Body)

		if readResponseError != nil {
			panic(readResponseError.Error())
		}

		var createSessionResponse CreateSessionResponse
		json.Unmarshal(body, &createSessionResponse)
		sessionId = strconv.FormatInt(createSessionResponse.SessionId, 10)
	} else if operation == "list" {
		response, getSessionsListError := http.Get("http://" + addr + "/session")
		if getSessionsListError != nil {
			panic(getSessionsListError.Error())
		}
		body, readResponseError := ioutil.ReadAll(response.Body)
		if readResponseError != nil {
			panic(readResponseError.Error())
		}

		listSessions := make([]SessionDto, 0)
		json.Unmarshal(body, &listSessions)
		fmt.Println("Sessions:")
		for _, session := range listSessions {
			fmt.Printf("Id: %d Started: %t", session.SessionId, session.Started)
			fmt.Println()
		}
		return
	}
	u := url.URL{Scheme: "ws", Host: addr, Path: "/session/connect/" + sessionId}

	connect, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer connect.Close()

	InitClear()
	CallClear()
	hideCursor()

	fmt.Println("SessionId: " + sessionId)

	done := make(chan struct{})

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	keyboardChannel := initInputChannel()
	handleSigtermExit(keyboardChannel, connect)
	go readProcessor(done, connect, keyboardChannel)
	sendProcessor(done, ticker, connect, interrupt, keyboardChannel)
}

func sendProcessor(
	done chan struct{},
	ticker *time.Ticker,
	c *websocket.Conn,
	interrupt chan os.Signal,
	keyboardChannel *tty.TTY,
) {
	keyboardSendChannel := make(chan rune)
	// input
	go func(keyboardChannel *tty.TTY, keyboardSendChannel chan<- rune) {
		for {
			r, err := keyboardChannel.ReadRune()
			if err != nil {
				log.Fatal(err)
			}
			//fmt.Println("Key press => " + string(r))
			keyboardSendChannel <- r
		}
	}(keyboardChannel, keyboardSendChannel)
	for {
		select {
		case <-done:
			return
		case messageFromKeyboard := <-keyboardSendChannel:
			err := c.WriteMessage(websocket.TextMessage, []byte(string(messageFromKeyboard)))
			if err != nil {
				// log.Println("write:", err)
				return
			}
		//case ticker := <-ticker.C:
		//	err := c.WriteMessage(websocket.TextMessage, []byte("CHURKA "+ticker.String()))
		//	if err != nil {
		//		log.Println("write:", err)
		//		return
		//	}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func hideCursor() {
	fmt.Print(hideCursorASCII)
}

func showCursor() {
	fmt.Print(showCursorASCII)
}

func handleSigtermExit(keyboardChannel *tty.TTY, conn *websocket.Conn) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		onExit(keyboardChannel, conn, "")
	}()
}

// onExit Closes keyboard input stream and makes cursor visible back
func onExit(keyboardChannel *tty.TTY, conn *websocket.Conn, exitMessage string) {
	showCursor()
	keyboardChannel.Close()
	conn.Close()
	CallClear()
	fmt.Println(exitMessage)
	os.Exit(1)
}

func initInputChannel() *tty.TTY {
	keyPressedChannel, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	return keyPressedChannel
}

func readProcessor(done chan struct{}, c *websocket.Conn, keyboardChannel *tty.TTY) {
	func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				// TODO: если закрыли сокет, то попадем сюда. Нужно обработать и завершить без ошибки
				log.Println("read:", err)
				return
			}
			fields := strings.Fields(string(message))
			if fields[0] == "0" {
				// self field
				if fields[1] == "0" {
					onExit(keyboardChannel, c, fields[2])
				}
				field, _ := big.NewInt(0).SetString(string(fields[2]), 10)
				speed := fields[3]
				score := fields[4]
				cleanCount := fields[5]
				nextPieceTypeIntRepr, _ := strconv.Atoi(fields[6])
				nextPieceType := PieceType(nextPieceTypeIntRepr)
				PrintSelfField(field, speed, score, cleanCount, nextPieceType)
			} else {
				// enemy field
				if fields[1] == "0" {

				}
				field, _ := big.NewInt(0).SetString(string(fields[2]), 10)
				speed := fields[3]
				score := fields[4]
				cleanCount := fields[5]
				nextPieceTypeIntRepr, _ := strconv.Atoi(fields[6])
				nextPieceType := PieceType(nextPieceTypeIntRepr)
				PrintEnemyField(field, speed, score, cleanCount, nextPieceType)
			}
		}
	}()
}
