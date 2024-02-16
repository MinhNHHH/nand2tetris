package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type CommandType int
type ArgType string
type ASMType string

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
	arg1           ArgType
	arg2           ArgType
	currentCommand string
	currentRow     int
}

type Writer struct {
	writer *os.File
}

type CodeWriter struct {
	parser *Parser
	writer *Writer
	label  int
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

func (p *Parser) hasMoreCommands() bool {
	if p.currentRow < len(p.lines) {
		return true
	}
	return false
}

func (p *Parser) getArg1() ArgType {
	if p.currentCommand == "" {
		return ""
	}
	splitText := strings.Split(p.currentCommand, " ")
	if p.commandType(splitText[0]) == C_RETURN {
		return ""
	}
	return ArgType(splitText[0])
}

func (p *Parser) getArg2() ArgType {
	if p.currentCommand == "" {
		return ""
	}
	splitText := strings.Split(p.currentCommand, " ")
	if len(splitText) < 2 {
		return ""
	}
	return ArgType(splitText[1])
}

func (p *Parser) commandType(arg string) CommandType {
	if checkExist(arithmetic, arg) {
		return C_ARITHMETIC
	} else if arg == "pop" {
		return C_POP
	} else if arg == "push" {
		return C_PUSH
	} else if arg == "goto" {
		return C_GOTO
	} else if arg == "if" {
		return C_IF
	} else if arg == "function" {
		return C_FUNCTION
	} else if arg == "label" {
		return C_LABEL
	} else if arg == "call" {
		return C_CALL
	} else if arg == "return" {
		return C_RETURN
	}
	return C_UNKNOW
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
	p.arg1 = p.getArg1()
	p.arg2 = p.getArg2()
	p.currentRow += 1
}

func (p *Parser) reset() {
	p.currentRow = 0
	p.currentCommand = ""
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

func NewCodeWriter(parse *Parser, writer *Writer) *CodeWriter {
	return &CodeWriter{
		parser: parse,
		writer: writer,
	}
}

func (c *CodeWriter) add() ASMType {
	translator := ""
	translator += "@SP\n"    // Sets the address of the stack pointer (SP) to the A-register.
	translator += "AM=M-1\n" // Decrements the stack pointer (SP) and sets the memory address (M) to point to the value at the top of the stack (M-1)
	translator += "D=M\n"    // Copies the value from the memory address pointed to by the stack pointer into register D, storing the second operand.
	translator += "@SP\n"    // pop first value into D
	translator += "AM=M-1\n"
	translator += "M=D+M\n" // Adds the value in register D (the second operand) to the value at the top of the stack (the first operand), storing the result back at the top of the stack.
	translator += "@SP\n"
	translator += "M=M+1\n" // Increments the stack pointer (SP) to indicate that there is one more item on the stack.
	return ASMType(translator)
}

func (c *CodeWriter) sub() ASMType {
	translator := ""
	translator += "@SP\n"
	translator += "@AM=M-1\n"
	translator += "D=M\n"
	translator += "@SP\n" // pop first value into D
	translator += "AM=M-1\n"
	translator += "M=M-D\n"
	translator += "@SP\n"
	translator += "M=M+1\n" // Increments the stack pointer (SP) to indicate that there is one more item on the stack.
	return ASMType(translator)
}

func (c *CodeWriter) neg() ASMType {
	translator := ""
	translator += "@SP\n"
	translator += "@A=M-1\n"
	translator += "M=-M\n"
	return ASMType(translator)
}

func (c *CodeWriter) not() ASMType {
	translator := ""
	translator += "@SP\n"
	translator += "@A=M-1\n"
	translator += "@M=!M\n"
	return ASMType(translator)
}

func (c *CodeWriter) and() ASMType {
	translator := ""
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "M=D&M\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) or() ASMType {
	translator := ""
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "M=D|M\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) eq() ASMType {
	label := c.label
	translator := ""
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	translator += "AM=M-1\n"
	translator += "D=M-D\n"
	translator += "M=-1\n"                               // Assume they are equal and store true (-1) at the top of the stack
	translator += fmt.Sprintf("@EQ_END%d", label) + "\n" // Jump to EQ_END if the top two values are equal
	translator += "D;JEQ\n"
	translator += "@SP\n" // If not equal, store false (0) at the top of the stack
	translator += "A=M\n"
	translator += "M=0\n"
	translator += fmt.Sprintf("(EQ_END%d)", label) + "\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	c.label += 1
	return ASMType(translator)
}

func (c *CodeWriter) gt() ASMType {
	label := c.label
	translator := ""
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	translator += "AM=M-1\n"
	translator += "D=M-D\n"
	translator += "M=-1\n" // Assume greater and store true (-1) at the top of the stack
	translator += fmt.Sprintf("@GT_END%d", label) + "\n"
	translator += "D;JGT\n"
	translator += "@SP\n" // If not great, store false (0) at the top of the stack
	translator += "A=M\n"
	translator += "M=0\n"
	translator += fmt.Sprintf("(GT_END%d)", label) + "\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	c.label += 1
	return ASMType(translator)
}

func (c *CodeWriter) lt() ASMType {
	label := c.label
	translator := ""
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	translator += "AM=M-1\n"
	translator += "D=M-D\n"
	translator += "M=-1\n" // Assume lesser and store true (-1) at the top of the stack
	translator += fmt.Sprintf("@LT_END%d", label) + "\n"
	translator += "D;JGT\n"
	translator += "@SP\n" // If not less, store false (0) at the top of the stack
	translator += "A=M\n"
	translator += "M=0\n"
	translator += fmt.Sprintf("(LT_END%d)", label) + "\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	c.label += 1
	return ASMType(translator)
}

func (c *CodeWriter) writeArithmetic(command string) (ASMType, error) {
	switch command {
	case "add":
		return c.add(), nil
	case "sub":
		return c.sub(), nil
	case "neg":
		return c.neg(), nil
	case "lt":
		return c.lt(), nil
	case "gt":
		return c.gt(), nil
	case "eq":
		return c.eq(), nil
	case "and":
		return c.and(), nil
	case "not":
		return c.not(), nil
	case "or":
		return c.or(), nil
	}
	return "", fmt.Errorf("The command not implemented yet")
}

func (c *CodeWriter) genCode() {
	c.parser.reset()
	for c.parser.hasMoreCommands() {
		c.parser.advance()
		fmt.Println(c.writeArithmetic(string(c.parser.arg1)))
	}
}

func main() {
	if len(os.Args) != 2 {
		return
	}

	// Frist pass: Clear comment
	parser, err := NewParser(os.Args[1])
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	outputFile := strings.Replace(os.Args[1], ".vm", ".asm", 1)
	writer, err2 := createFile(outputFile)
	if err2 != nil {
		fmt.Println("Error:", err)
		return
	}

	codeWriter := NewCodeWriter(parser, writer)
	codeWriter.genCode()
}
