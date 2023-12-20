package main

import (
	"flag"
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

type Client struct {
	conn *websocket.Conn
}

var addr = flag.String("addr", "localhost:8080", "http service address")

// https://github.com/gorilla/websocket/blob/main/examples/echo/server.go
func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/echo"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go readProcessor(done, c)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	sendProcessor(done, ticker, c, interrupt)
}

func sendProcessor(
	done chan struct{},
	ticker *time.Ticker,
	c *websocket.Conn,
	interrupt chan os.Signal,
) {
	keyboardSendChannel := make(chan rune)
	keyboardChannel := initInputChannel()
	defer onExit(keyboardChannel)
	handleSigtermExit(keyboardChannel)
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

func handleSigtermExit(keyboardChannel *tty.TTY) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		onExit(keyboardChannel)
		os.Exit(1)
	}()
}

// onExit Closes keyboard input stream and makes cursor visible back
func onExit(keyboardChannel *tty.TTY) {
	keyboardChannel.Close()
}

func initInputChannel() *tty.TTY {
	keyPressedChannel, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	return keyPressedChannel
}

func readProcessor(done chan struct{}, c *websocket.Conn) {
	func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			field, _ := big.NewInt(0).SetString(string(message), 10)
			PrintField(field)
			//log.Printf("recv: %s", message)
		}
	}()
}

var builder = strings.Builder{}

const moveToTopASCII = "\033[22A"

func PrintField(field *big.Int) {
	builder.Reset()
	builder.WriteString(moveToTopASCII)
	fieldStr := fmt.Sprintf("%b", field)
	for i := 20; i >= 0; i-- {
		line := fieldStr[i*12 : i*12+12]
		line = strings.ReplaceAll(line, "1", " Ж ")
		line = strings.ReplaceAll(line, "0", "   ")
		builder.WriteString(line)
		builder.WriteString("\n")
	}
	builder.WriteString("Score: ")
	builder.WriteString(strconv.Itoa(0))
	builder.WriteString(" | Speed: ")
	builder.WriteString(strconv.Itoa(0))
	builder.WriteString(" | Cleaned: ")
	builder.WriteString(strconv.Itoa(0))
	fmt.Println(builder.String())
	//printNextPiece(field.NextPiece)
}

//func (gameField *Field) String() string {
//	newField := big.NewInt(0).Set(gameField.Val)
//	newShape := big.NewInt(0).Set(gameField.CurrentPiece.GetVal())
//	newField.Or(newField, newShape)
//	return fmt.Sprintf("%b", newField)
//}
