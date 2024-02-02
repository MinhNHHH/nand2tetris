package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
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
	"null": "000",
	"JGT":  "001",
	"JEQ":  "010",
	"JGE":  "011",
	"JLT":  "100",
	"JNE":  "101",
	"JLE":  "110",
	"JMP":  "111",
}

var destCode = map[string]string{
	"null": "000",
	"M":    "001",
	"D":    "010",
	"MD":   "011",
	"A":    "100",
	"AM":   "101",
	"AD":   "110",
	"AMD":  "111",
}

var compCode = map[string]string{
	"0":   "0101010",
	"1":   "0111111",
	"-1":  "0111010",
	"D":   "0001100",
	"A":   "0110000",
	"M":   "1110000",
	"!D":  "0001101",
	"!A":  "0110001",
	"!M":  "1110001",
	"-D":  "0001111",
	"-A":  "0110011",
	"-M":  "1110011",
	"D+1": "0011111",
	"A+1": "0110111",
	"M+1": "1110111",
	"D-1": "0001110",
	"A-1": "0110010",
	"M-1": "1110010",
	"D+A": "0000010",
	"D+M": "1000010",
	"D-A": "0010011",
	"D-M": "1010011",
	"A-D": "0000111",
	"M-D": "1000111",
	"D&A": "0000000",
	"D&M": "1000000",
	"D|A": "0010101",
	"D|M": "1010101",
}

type Parser struct {
	lines          []string
	currentCommand string
	currentRow     int
}

type Writer struct {
	writer *os.File
}

func createFile(fileName string) (*Writer, error) {
	f, err := os.Create(fileName)
	if err != nil {
		return nil, err
	}

	return &Writer{
		writer: f,
	}, nil

}

func NewParser(fileName string) (*Parser, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(file)
	lines := []string{}
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return &Parser{
		lines: lines,
	}, nil

}

func (p *Parser) hasMoreCommands() bool {
	if p.currentRow < len(p.lines) {
		return true
	}
	return false
}

func (p *Parser) advance() {
	trimmedLine := strings.TrimSpace(string(p.lines[p.currentRow]))

	if strings.HasPrefix(trimmedLine, "//") || trimmedLine == "" {
		p.currentRow += 1
		return
	}

	if index := strings.Index(trimmedLine, "//"); index != -1 {
		p.currentCommand = trimmedLine[:index]
	} else {
		p.currentCommand = trimmedLine
	}
	p.currentRow += 1
}

func (p *Parser) reset() {
	p.currentRow = 0
	p.currentCommand = ""
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
	if p.currentCommand == "" {
		return -1
	}
	if strings.HasPrefix(p.currentCommand, "@") {
		return ACommand
	} else if strings.HasPrefix(p.currentCommand, "(") {
		return LCommand
	}
	return CCommand
}

func (p *Parser) dest(code string) string {
	return destCode[code]
}

func (p *Parser) comp(code string) string {
	return compCode[code]
}

func (p *Parser) jump(code string) string {
	return jumpCode[code]
}

type TranslateInstruction struct {
	writer         *Writer
	parser         *Parser
	currentAddress int
	currentROM     int
	symbolTable    map[string]int
}

func NewTranslateInstruction(parser *Parser, writer *Writer) *TranslateInstruction {
	return &TranslateInstruction{
		writer:         writer,
		parser:         parser,
		currentAddress: 16,
		symbolTable:    symbolTable,
	}
}

func (t *TranslateInstruction) buildSymbolTable() {
	for t.parser.hasMoreCommands() {
		t.parser.advance()
		if t.parser.commandType() == LCommand {
			symbol := t.parser.symbol()
			t.symbolTable[symbol] = t.currentROM
		} else if t.parser.commandType() == CCommand || t.parser.commandType() == ACommand {
			t.currentROM++
		}
	}
}

func convertNumberToBinary(num int) string {
	binaryString := strconv.FormatInt(int64(num), 2)
	paddedBinaryString := fmt.Sprintf("%016s", binaryString)
	return paddedBinaryString
}

func (t *TranslateInstruction) genCode() {
	t.parser.reset()
	for t.parser.hasMoreCommands() {
		t.parser.advance()
		if t.parser.commandType() == ACommand {
			res := ""
			if value, err := strconv.Atoi(t.parser.symbol()); err == nil {
				res = convertNumberToBinary(value)
			} else {
				if value, exist := t.symbolTable[t.parser.symbol()]; exist {
					res = convertNumberToBinary(value)
				} else {
					t.symbolTable[t.parser.symbol()] = t.currentAddress
					res = convertNumberToBinary(t.symbolTable[t.parser.symbol()])
					t.currentAddress++
				}
			}
			t.writer.writer.WriteString(res + "\n")
		} else if t.parser.commandType() == CCommand {
			CCode := "111"
			code := ""
			if strings.Contains(t.parser.currentCommand, "=") {
				splitCode := strings.Split(t.parser.currentCommand, "=")
				code += t.parser.comp(strings.TrimSpace(splitCode[1])) + t.parser.dest(splitCode[0]) + t.parser.jump("null")
			} else if strings.Contains(t.parser.currentCommand, ";") {
				splitCode := strings.Split(t.parser.currentCommand, ";")
				code += t.parser.comp(splitCode[0]) + t.parser.dest("null") + t.parser.jump(strings.TrimSpace(splitCode[1]))
			}
			CCode += code
			t.writer.writer.WriteString(CCode + "\n")
		}
	}
}

func main() {
	if len(os.Args) != 2 {
		return
	}

	// Frist pass: Clear comment and build a symbolTable
	parser, err := NewParser(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	outputFile := strings.Split(os.Args[1], ".")[0] + ".hack"
	writer, err2 := createFile(outputFile)
	if err2 != nil {
		fmt.Println("Error:", err)
		return
	}

	translateInstruction := NewTranslateInstruction(parser, writer)
	translateInstruction.buildSymbolTable()

	// Second pass: Translate instruction to binary
	translateInstruction.genCode()

	translateInstruction.writer.writer.Close()
}
