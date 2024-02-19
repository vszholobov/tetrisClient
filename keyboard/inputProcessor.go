package keyboard

import (
	"github.com/mattn/go-tty"
)

type InputProcessor struct {
	keyboardTransferChannel chan rune
	keyboardInputChannel    *tty.TTY
}

func MakeInputProcessor() *InputProcessor {
	keyboardInputChannel, _ := tty.Open()
	keyboardTransferChannel := make(chan rune)
	return &InputProcessor{
		keyboardTransferChannel: keyboardTransferChannel,
		keyboardInputChannel:    keyboardInputChannel,
	}
}

func (inputProcessor *InputProcessor) ProcessKeyboardInput() {
	for {
		r, _ := inputProcessor.keyboardInputChannel.ReadRune()
		inputProcessor.keyboardTransferChannel <- r
	}
}

func (inputProcessor *InputProcessor) GetKeyboardInputTransferChannel() chan rune {
	return inputProcessor.keyboardTransferChannel
}

func (inputProcessor *InputProcessor) Close() {
	inputProcessor.keyboardInputChannel.Close()
}
