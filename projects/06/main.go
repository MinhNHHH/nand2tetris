package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CommandType int

const (
	ACommand CommandType = iota
	CCommand
	LCommand
)

// symbolTable contain key is label and value is RAM address.
var symbolTable = map[string]int{
	"SP":     0,
	"LCL":    1,
	"ARG":    2,
	"THIS":   3,
	"THAT":   4,
	"SCREEN": 16384,
	"KBD":    24576,
	"R0":     0,
	"R1":     1,
	"R2":     2,
	"R3":     3,
	"R4":     4,
	"R5":     5,
	"R6":     6,
	"R7":     7,
	"R8":     8,
	"R9":     9,
	"R10":    10,
	"R11":    11,
	"R12":    12,
	"R13":    13,
	"R14":    14,
	"R15":    15,
}

var jumpCode = map[string]string{
	"JGT": "001",
	"JEQ": "010",
	"JLT": "100",
	"JNE": "101",
	"JLE": "110",
	"JMP": "111",
}

var destCode = map[string]string{
	"M":   "001",
	"D":   "010",
	"MD":  "011",
	"A":   "100",
	"AM":  "101",
	"AD":  "110",
	"AMD": "111",
}

var compCode = map[string]string{
	"0":   "101010",
	"1":   "111111",
	"-1":  "111110",
	"D":   "001100",
	"A":   "110000",
	"M":   "110000",
	"!D":  "001101",
	"!A":  "110001",
	"!M":  "110001",
	"-D":  "001111",
	"-A":  "110011",
	"-M":  "110011",
	"D+1": "011111",
	"A+1": "110111",
	"M+1": "110111",
	"D-1": "001110",
	"A-1": "110010",
	"M-1": "110010",
	"D+A": "000010",
	"D+M": "000010",
	"D-A": "010011",
	"D-M": "010011",
	"A-D": "000111",
	"M-D": "000111",
	"D&A": "000000",
	"D&M": "000000",
	"D|A": "010101",
	"D|M": "010101",
}

type Parser struct {
	scanner        *bufio.Scanner
	currentRow     int
	currentCommand string
}

func NewParser(fileName string) (*Parser, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	return &Parser{
		scanner: bufio.NewScanner(file),
	}, nil
}

func (p *Parser) hasMoreCommands() bool {
	return p.scanner.Scan()
}

func (p *Parser) advance() {
	if !strings.HasPrefix(p.scanner.Text(), "//") {
		trimmedLine := strings.TrimSpace(p.scanner.Text())
		if index := strings.Index(trimmedLine, "//"); index != -1 {
			p.currentCommand = trimmedLine[:index]
		} else {
			p.currentCommand = trimmedLine
		}
		p.currentRow += 1
	}
}

func (p *Parser) symbol() string {
	switch p.commandType() {
	case ACommand:
		return p.currentCommand[1:]
	case LCommand:
		return p.currentCommand[1 : len(p.currentCommand)-1]
	}
	return ""
}

func (p *Parser) commandType() CommandType {
	if strings.HasPrefix(p.currentCommand, "@") {
		return ACommand
	} else if strings.HasPrefix(p.currentCommand, "(") {
		return LCommand
	}
	return CCommand
}

func (p *Parser) dest() string {
	if p.commandType() == CCommand && strings.Contains(p.currentCommand, "=") {
		return strings.Split(p.currentCommand, "=")[0]
	}
	return ""
}

func (p *Parser) comp() string {
	if p.commandType() == CCommand && strings.Contains(p.currentCommand, "=") {
		splitCommand := strings.Split(p.currentCommand, "=")
		return strings.Split(splitCommand[1], ";")[0]
	}
	return ""
}

func (p *Parser) jump() string {
	if p.commandType() == CCommand && strings.Contains(p.currentCommand, "=") {
		return strings.Split(p.currentCommand, ";")[1]
	}
	return ""
}

type TranslateInstruction struct {
	parser         *Parser
	currentAddress int // 16
	symbolTable    map[string]int
}

func NewTranslateInstruction(parser *Parser) *TranslateInstruction {
	return &TranslateInstruction{
		parser:         parser,
		currentAddress: 16,
		symbolTable:    symbolTable,
	}
}

func (t *TranslateInstruction) buildSymbolTable() {
	for t.parser.hasMoreCommands() {
		t.parser.advance()
		fmt.Println(t.parser.currentCommand, t.parser.currentRow)
		if t.parser.commandType() == LCommand {
			symbol := t.parser.symbol()
			t.symbolTable[symbol] = t.parser.currentRow
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		return
	}

	parser, err := NewParser(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	translateInstruction := NewTranslateInstruction(parser)

	translateInstruction.buildSymbolTable()
	fmt.Println(translateInstruction.symbolTable)
	if err := parser.scanner.Err(); err != nil {
		fmt.Println(err)
		return
	}
}
