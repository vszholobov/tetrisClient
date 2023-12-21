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

var builder = strings.Builder{}

func PrintField(field *big.Int, speed string, score string, cleanCount string, nextPieceType PieceType) {
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
	builder.WriteString(score)
	builder.WriteString(" | Speed: ")
	builder.WriteString(speed)
	builder.WriteString(" | Cleaned: ")
	builder.WriteString(cleanCount)
	fmt.Println(builder.String())
	printNextPiece(nextPieceType)
}

func printNextPiece(nextPieceType PieceType) {
	//pieceVal := nextPiece.GetVal()

	fmt.Print(moveToTopASCII + moveRightASCII + " ##############")
	fmt.Printf(moveDownOneLineASCII + moveRightASCII + " #            #")
	pieceLines := RepresentationByType[nextPieceType]
	for i := 0; i < 2; i++ {
		curLine := "            "
		if i < len(pieceLines) {
			curLine = pieceLines[i]
		}
		fmt.Printf(moveDownOneLineASCII+moveRightASCII+" #%s#", curLine)
		//curLine := big.NewInt(0).Lsh(fullLine, uint(i)*FieldWidth)
		//checkCurrLine := big.NewInt(0).And(curLine, nextPiece.GetVal())
		//line := fmt.Sprintf("%10b", checkCurrLine)
		//line = strings.ReplaceAll(line, "1", "Ж")
		//line = strings.ReplaceAll(line, "0", "")
		//fmt.Print(moveDownOneLineASCII + moveRightASCII + " #          #")
	}
	fmt.Printf(moveDownOneLineASCII + moveRightASCII + " #            #")
	fmt.Print(moveDownOneLineASCII + moveRightASCII + " ##############")
	fmt.Print(moveDownAllLinesASCII)

	//fmt.Print(moveDownOneLineASCII + moveRightASCII + " #          #")
	//fmt.Print(moveDownOneLineASCII + moveRightASCII + " #          #")
	//fmt.Print(moveDownOneLineASCII + moveRightASCII + " #          #")
}
