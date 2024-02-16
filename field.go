package main

import (
	"fmt"
	"math/big"
	"strings"
)

const FieldWidth = 12
const FieldHeight = 21
const moveToTopASCII = "\033[22A"
const moveRightASCII = "\r\033[36C"
const moveRightEnemyFieldASCII = "\r\033[52C"
const moveDownOneLineASCII = "\r\033[1B"
const moveDownAllLinesASCII = "\r\033[17B"

type PieceType int

const (
	TShape      PieceType = 0
	ZigZagLeft  PieceType = 1
	ZigZagRight PieceType = 2
	RightLShape PieceType = 3
	LeftLShape  PieceType = 4
	IShape      PieceType = 5
	SquareShape PieceType = 6
)

var RepresentationByType = map[PieceType][]string{
	TShape:      {"      Ж     ", "     ЖЖЖ    "},
	ZigZagRight: {"      ЖЖ    ", "     ЖЖ     "},
	ZigZagLeft:  {"     ЖЖ     ", "      ЖЖ    "},
	IShape:      {"    ЖЖЖЖ    "},
	RightLShape: {"    ЖЖЖ     ", "    Ж       "},
	LeftLShape:  {"    ЖЖЖ     ", "      Ж     "},
	SquareShape: {"     ЖЖ     ", "     ЖЖ     "},
}

func PrintEnemyField(field *big.Int, speed string, score string, cleanCount string, nextPieceType PieceType) {
	fieldStr := fmt.Sprintf("%b", field)
	fmt.Print(moveToTopASCII)
	fmt.Print(moveRightEnemyFieldASCII)
	for i := FieldHeight - 1; i >= 0; i-- {
		line := fieldStr[i*FieldWidth : i*FieldWidth+FieldWidth]
		line = strings.ReplaceAll(line, "1", " Ж ")
		line = strings.ReplaceAll(line, "0", "   ")
		fmt.Print(line)
		fmt.Print(moveDownOneLineASCII)
		fmt.Print(moveRightEnemyFieldASCII)
	}
	builder.Reset()
	builder.WriteString("Score: ")
	builder.WriteString(score)
	builder.WriteString(" | Speed: ")
	builder.WriteString(speed)
	builder.WriteString(" | Cleaned: ")
	builder.WriteString(cleanCount)
	fmt.Print(builder.String())
	fmt.Print(moveDownOneLineASCII)
}

var builder = strings.Builder{}

func PrintSelfField(
	field *big.Int,
	speed string,
	score string,
	cleanCount string,
	nextPieceType PieceType,
	pingMs string,
) {
	fieldStr := fmt.Sprintf("%b", field)
	fmt.Print(moveToTopASCII)
	for i := FieldHeight - 1; i >= 0; i-- {
		line := fieldStr[i*FieldWidth : i*FieldWidth+FieldWidth]
		line = strings.ReplaceAll(line, "1", " Ж ")
		line = strings.ReplaceAll(line, "0", "   ")
		fmt.Print(line)
		fmt.Print(moveDownOneLineASCII)
	}
	builder.Reset()
	builder.WriteString("Score: ")
	builder.WriteString(score)
	builder.WriteString(" | Speed: ")
	builder.WriteString(speed)
	builder.WriteString(" | Cleaned: ")
	builder.WriteString(cleanCount)
	builder.WriteString(" | Ping: ")
	builder.WriteString(pingMs)
	builder.WriteString("    ")
	fmt.Print(builder.String())
	fmt.Print(moveDownOneLineASCII)
	printNextPiece(nextPieceType)
}

func printNextPiece(nextPieceType PieceType) {
	fmt.Print(moveToTopASCII + moveRightASCII + " ##############")
	fmt.Printf(moveDownOneLineASCII + moveRightASCII + " #            #")
	pieceLines := RepresentationByType[nextPieceType]
	for i := 0; i < 2; i++ {
		curLine := "            "
		if i < len(pieceLines) {
			curLine = pieceLines[i]
		}
		fmt.Printf(moveDownOneLineASCII+moveRightASCII+" #%s#", curLine)
	}
	fmt.Printf(moveDownOneLineASCII + moveRightASCII + " #            #")
	fmt.Print(moveDownOneLineASCII + moveRightASCII + " ##############")
	fmt.Print(moveDownAllLinesASCII)
}
