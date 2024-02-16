package main

import (
	"bufio"
	"os"
	"strings"
)

type CommandType int
type ArgType string

const (
	C_UNKNOW     CommandType = -1
	C_ARITHMETIC CommandType = iota
	C_PUSH
	C_POP
	C_LABEL
	C_GOTO
	C_IF
	C_FUNCTION
	C_RETURN
	C_CALL
)

type Parser struct {
	lines          []string
	currentCommand string
	currentRow     int
}

var arithmetic = []string{"add", "sub", "neg", "eq", "gt", "lt", "and", "or", "not"}

func checkExist(listElement []string, element string) bool {
	for _, value := range listElement {
		if value == element {
			return true
		}
	}
	return false
}

func NewParser(filePath string) (*Parser, error) {
	file, err := os.Open(filePath)
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

func (p *Parser) hasMoreCommand() bool {
	if p.currentRow < len(p.lines) {
		return true
	}
	return false
}

func (p *Parser) commandType() CommandType {
	if p.currentCommand == "" {
		return C_UNKNOW
	}
	// argCommands := strings.Split(p.currentCommand, " ")
	// if argCommands[0] == "pop" {
	// 	return C_POP
	// } else if argCommands[0] == "push" {
	// 	return C_PUSH
	// } else if checkExist(arithmetic, argCommands[0]) {
	// 	return C_ARITHMETIC
	// }
	return 1
}

func (p *Parser) arg1() ArgType {
	if p.currentCommand == "" {
		return ""
	}
	argCommands := strings.Split(p.currentCommand, " ")
	if len(argCommands) > 0 && argCommands[0] != "return" {
		return ArgType(argCommands[0])
	}
	return ""
}

func (p *Parser) arg2() CommandType {
	return 1
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

func main() {

}
