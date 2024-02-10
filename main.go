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

type Session struct {
	conn            *websocket.Conn
	keyboardChannel *tty.TTY
}

var addr = "84.201.177.35:8080"
var session *Session

// https://github.com/gorilla/websocket/blob/main/examples/echo/server.go
func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// keyboard
	InitClear()
	CallClear()
	hideCursor()
	keyboardInputChannel := make(chan rune)
	keyboardChannel := initInputChannel()
	session = &Session{keyboardChannel: keyboardChannel}
	handleSigtermExit(session)
	go inputProcessor(keyboardChannel, keyboardInputChannel)

	var sessionId string
	if len(os.Args) < 2 {
		menu := MakeMenu()
		menu.showMenu()
		menu.handleMenu(keyboardInputChannel)
		if menu.isExit {
			onExit("")
		}
		if menu.isCreateSession {
			sessionId = createSession()
		} else {
			sessionId = strconv.FormatInt(menu.sessionsList[menu.currentSessionIndex].SessionId, 10)
		}
	} else if operation := os.Args[1]; operation == "connect" {
		sessionId = os.Args[2]
	} else if operation == "create" {
		sessionId = createSession()
	} else if operation == "list" {
		listSessions := getSessionsList()
		fmt.Println("Sessions:")
		for _, session := range listSessions {
			fmt.Printf("Id: %d Started: %t", session.SessionId, session.Started)
			fmt.Println()
		}
		return
	} else {
		fmt.Println("Error")
		return
	}
	sessionConnectUrl := url.URL{Scheme: "ws", Host: addr, Path: "/session/connect/" + sessionId}

	connect, _, err := websocket.DefaultDialer.Dial(sessionConnectUrl.String(), nil)
	session.conn = connect
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer connect.Close()

	fmt.Println("SessionId: " + sessionId)

	// server
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	go readProcessor(connect, keyboardChannel)
	sendProcessor(ticker, connect, interrupt, keyboardInputChannel)
}

func createSession() string {
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
	return strconv.FormatInt(createSessionResponse.SessionId, 10)
}

func getSessionsList() []SessionDto {
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
	return listSessions
}

func inputProcessor(keyboardChannel *tty.TTY, keyboardSendChannel chan<- rune) {
	for {
		r, err := keyboardChannel.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		keyboardSendChannel <- r
	}
}

func sendProcessor(
	ticker *time.Ticker,
	c *websocket.Conn,
	interrupt chan os.Signal,
	keyboardSendChannel chan rune,
) {
	for {
		select {
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

func handleSigtermExit(session *Session) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		onExit("")
	}()
}

func initInputChannel() *tty.TTY {
	keyPressedChannel, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	return keyPressedChannel
}

// readProcessor server handler
func readProcessor(c *websocket.Conn, keyboardChannel *tty.TTY) {
	func() {
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
					onExit(fields[2])
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

// onExit Closes keyboard input stream and makes cursor visible back
func onExit(exitMessage string) {
	showCursor()
	CallClear()
	fmt.Println(exitMessage)
	if session.keyboardChannel != nil {
		session.keyboardChannel.Close()
	}
	if session.conn != nil {
		session.conn.Close()
	}
	os.Exit(0)
}

type Menu struct {
	currentSessionIndex int
	sessionsList        []SessionDto
	isEnded             bool
	isCreateSession     bool
	isExit              bool
}

func MakeMenu() Menu {
	sessionsList := getSessionsList()
	return Menu{
		currentSessionIndex: 0,
		sessionsList:        sessionsList,
		isEnded:             false,
		isCreateSession:     false,
		isExit:              false,
	}
}

func (menu *Menu) showMenu() {
	CallClear()
	fmt.Println(" Tetris🕹️")
	fmt.Println("----------")
	for index, session := range menu.sessionsList {
		currentItem := ""
		if index == menu.currentSessionIndex {
			currentItem += "\033[30;5;107m"
		}
		currentItem += strconv.FormatInt(session.SessionId, 10)
		currentItem += " "
		currentItem += strconv.FormatBool(session.Started)
		if index == menu.currentSessionIndex {
			currentItem += "\033[0m"
		}
		fmt.Println(currentItem)
	}
}

func (menu *Menu) handleMenu(keyboardInputChannel chan rune) {
	for !menu.isEnded {
		input := <-keyboardInputChannel
		switch input {
		case 115:
			// s
			if len(menu.sessionsList) == 0 {
				continue
			}
			menu.currentSessionIndex++
			menu.currentSessionIndex = menu.currentSessionIndex % len(menu.sessionsList)
		case 119:
			// w
			if len(menu.sessionsList) == 0 {
				continue
			}
			menu.currentSessionIndex--
			if menu.currentSessionIndex < 0 {
				menu.currentSessionIndex = len(menu.sessionsList) - 1
			}
		case 114:
			// r
			menu.sessionsList = getSessionsList()
		case 99:
			// c
			menu.isEnded = true
			menu.isCreateSession = true
			continue
		case 13:
			// enter
			if len(menu.sessionsList) == 0 {
				continue
			}
			menu.isEnded = true
			continue
		case 27:
			// esc
			menu.isEnded = true
			menu.isExit = true
		default:
			// skip unknown input
			continue
		}
		menu.showMenu()
	}
}
