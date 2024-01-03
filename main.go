package main

import (
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mattn/go-tty"
	"log"
	"math/big"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

const showCursorASCII = "\033[?25h"
const hideCursorASCII = "\033[?25l"

type Client struct {
	conn *websocket.Conn
}

var addr = "localhost:8080"

// https://github.com/gorilla/websocket/blob/main/examples/echo/server.go
func main() {
	InitClear()
	CallClear()
	hideCursor()
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	sessionId := os.Args[1]
	u := url.URL{Scheme: "ws", Host: addr, Path: "/connect/" + sessionId}
	//log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	keyboardChannel := initInputChannel()
	//defer onExit(keyboardChannel, c)
	handleSigtermExit(keyboardChannel, c)
	go readProcessor(done, c, keyboardChannel)
	sendProcessor(done, ticker, c, interrupt, keyboardChannel)
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
				log.Println("write:", err)
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
		onExit(keyboardChannel, conn)
	}()
}

// onExit Closes keyboard input stream and makes cursor visible back
func onExit(keyboardChannel *tty.TTY, conn *websocket.Conn) {
	showCursor()
	keyboardChannel.Close()
	conn.Close()
	CallClear()
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
				log.Println("Lost")
				onExit(keyboardChannel, c)
			}
			field, _ := big.NewInt(0).SetString(string(fields[1]), 10)
			speed := fields[2]
			score := fields[3]
			cleanCount := fields[4]
			nextPieceTypeIntRepr, err := strconv.Atoi(fields[5])
			nextPieceType := PieceType(nextPieceTypeIntRepr)
			PrintField(field, speed, score, cleanCount, nextPieceType)
			//log.Printf("recv: %s", message)
		}
	}()
}

//func (gameField *Field) String() string {
//	newField := big.NewInt(0).Set(gameField.Val)
//	newShape := big.NewInt(0).Set(gameField.CurrentPiece.GetVal())
//	newField.Or(newField, newShape)
//	return fmt.Sprintf("%b", newField)
//}
