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

type CodeWriter struct {
	parser *Parser
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
	// p.arg2 = p.getArg2()
	p.currentRow += 1
}

func (p *Parser) reset() {
	p.currentRow = 0
	p.currentCommand = ""
}

func (c *CodeWriter) popValueToD() ASMType {
	translator := ""
	translator += "@SP\n"
	translator += "M=M+1\n" // Increments the stack pointer (SP) to indicate that there is one more item on the stack.
	return ASMType(translator)
}

func (c *CodeWriter) setAddressToA() ASMType {
	translator := ""
	translator += "@SP\n"    // Sets the address of the stack pointer (SP) to the A-register.
	translator += "AM=M-1\n" // Decrements the stack pointer (SP) and sets the memory address (M) to point to the value at the top of the stack (M-1)
	translator += "D=M\n"    // Copies the value from the memory address pointed to by the stack pointer into register D, storing the second operand.
	return ASMType(translator)
}

func (c *CodeWriter) add() ASMType {
	translator := ""
	translator += string(c.setAddressToA())
	translator += "@SP\n" // pop first value into D
	translator += "AM=M-1\n"
	translator += "M=D+M\n" // Adds the value in register D (the second operand) to the value at the top of the stack (the first operand), storing the result back at the top of the stack.
	translator += string(c.popValueToD())
	return ASMType(translator)
}

func (c *CodeWriter) sub() ASMType {
	translator := ""
	translator += string(c.setAddressToA())
	translator += "@SP\n" // pop first value into D
	translator += "AM=M-1\n"
	translator += "M=D-M\n" // Adds the value in register D (the second operand) to the value at the top of the stack (the first operand), storing the result back at the top of the stack.
	translator += string(c.popValueToD())
	return ASMType(translator)
}

func (c *CodeWriter) neg() ASMType {
	translator := ""
	translator += "@SP\n"
	translator += "A=M-1\n"
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
	translator += string(c.setAddressToA())
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "M=D&M\n"
	translator += string(c.popValueToD())
	return ASMType(translator)
}

func (c *CodeWriter) or() ASMType {
	translator := ""
	translator += string(c.setAddressToA())
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "M=D|M\n"
	translator += string(c.popValueToD())
	return ASMType(translator)
}

func (c *CodeWriter) eq() ASMType {
	label := c.label
	translator := ""
	translator += string(c.setAddressToA())
	translator += "AM=M-1\n"
	translator += "D=M-D\n"
	translator += "M=-1\n"                               // Assume they are equal and store true (-1) at the top of the stack
	translator += fmt.Sprintf("@EQ_END%d", label) + "\n" // Jump to EQ_END if the top two values are equal
	translator += "D;JEQ\n"
	translator += "@SP\n" // If not equal, store false (0) at the top of the stack
	translator += "A=M\n"
	translator += "M=0\n"
	translator += fmt.Sprintf("(EQ_END%d)", label) + "\n"
	translator += string(c.popValueToD())
	c.label += 1
	return ASMType(translator)
}

func (c *CodeWriter) gt() ASMType {
	label := c.label
	translator := ""
	translator += string(c.setAddressToA())
	translator += "AM=M-1\n"
	translator += "D=M-D\n"
	translator += "M=-1\n" // Assume greater and store true (-1) at the top of the stack
	translator += fmt.Sprintf("@GT_END%d", label) + "\n"
	translator += "D;JGT\n"
	translator += "@SP\n" // If not great, store false (0) at the top of the stack
	translator += "A=M\n"
	translator += "M=0\n"
	translator += fmt.Sprintf("(GT_END%d)", label) + "\n"
	translator += string(c.popValueToD())
	c.label += 1
	return ASMType(translator)
}

func (c *CodeWriter) lt() ASMType {
	label := c.label
	translator := ""
	translator += string(c.setAddressToA())
	translator += "AM=M-1\n"
	translator += "D=M-D\n"
	translator += "M=-1\n" // Assume lesser and store true (-1) at the top of the stack
	translator += fmt.Sprintf("@LT_END%d", label) + "\n"
	translator += "D;JGT\n"
	translator += "@SP\n" // If not less, store false (0) at the top of the stack
	translator += "A=M\n"
	translator += "M=0\n"
	translator += fmt.Sprintf("(LT_END%d)", label) + "\n"
	translator += string(c.popValueToD())
	c.label += 1
	return ASMType(translator)
}

func (c *CodeWriter) pushConstant(index int) ASMType {
	translator := ""
	translator += fmt.Sprintf("@%d", index) + "\n"
	translator += "D=A\n"
	translator += "@SP\n"
	translator += "A=M\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) pushThis(index int) ASMType {
	translator := ""
	translator += "@THIS\n"                        // Address of base of 'this' segment
	translator += "D=M\n"                          // D = M[THIS], base address of 'this' segment
	translator += fmt.Sprintf("@%d", index) + "\n" // Offset to the desired element
	translator += "A=D+A\n"                        // Calculate address: THIS + index
	translator += "D=M\n"
	translator += "@SP\n"
	translator += "A=M\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) popThis(index int) ASMType {
	translator := ""
	translator += "@THIS\n"                        // Address of base of 'this' segment
	translator += "D=M\n"                          // D = M[THIS], base address of 'this' segment
	translator += fmt.Sprintf("@%d", index) + "\n" // Offset to the desired element
	translator += "D=D+A\n"                        // Calculate address: THIS + index
	translator += "@R13\n"                         // Temporarily store the address in R13
	translator += "M=D\n"                          // M[R13] = THIS + 0
	translator += "@SP\n"                          // Decrement SP and set A to point to top of stack
	translator += "AM=M-1\n"                       // Decrement the stack pointer and set A to point to the top of the stack
	translator += "D=M\n"                          // Load the value from the top of the stack to D-register
	translator += "@R13\n"                         // Retrieve the address from R13
	translator += "A=M\n"                          // Set the A-register to the target address
	translator += "M=D\n"                          // Store value from the stack into the target local variable
	return ASMType(translator)
}
func (c *CodeWriter) pushThat(index int) ASMType {
	translator := ""
	translator += "@THAT\n"                        // Address of base of 'this' segment
	translator += "D=M\n"                          // D = M[THAT], base address of 'this' segment
	translator += fmt.Sprintf("@%d", index) + "\n" // Offset to the desired element
	translator += "A=D+A\n"                        // Calculate address: THIS + index
	translator += "D=M\n"
	translator += "@SP\n"
	translator += "A=M\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) popThat(index int) ASMType {
	translator := ""
	translator += "@THAT\n"                        // Address of base of 'this' segment
	translator += "D=M\n"                          // D = M[THAT], base address of 'this' segment
	translator += fmt.Sprintf("@%d", index) + "\n" // Offset to the desired element
	translator += "D=D+A\n"                        // Calculate address: THIS + index
	translator += "@R13\n"                         // Temporarily store the address in R13
	translator += "M=D\n"                          // M[R13] = THIS + 0
	translator += "@SP\n"                          // Decrement SP and set A to point to top of stack
	translator += "AM=M-1\n"                       // Decrement the stack pointer and set A to point to the top of the stack
	translator += "D=M\n"                          // Load the value from the top of the stack to D-register
	translator += "@R13\n"                         // Retrieve the address from R13
	translator += "A=M\n"                          // Set the A-register to the target address
	translator += "M=D\n"                          // Store value from the stack into the target local variable
	return ASMType(translator)
}

func (c *CodeWriter) pushStatic(index int) ASMType {
	translator := ""
	translator += fmt.Sprintf("@STATIC_%d", index) + "\n"
	translator += "D=M\n"
	translator += "@SP\n"
	translator += "A=M\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) popStatic(index int) ASMType {
	translator := ""
	translator += fmt.Sprintf("@STATIC_%d", index) + "\n"
	translator += "D=A\n"
	translator += "@R13\n"
	translator += "@M=D\n" // M[index] = STATIC_index
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	translator += "@R13"
	translator += "A=M\n" // Set the A-register to the target address
	translator += "M=D\n" // Store value from the stack into the target local variable
	return ASMType(translator)
}

func (c *CodeWriter) pushLocal(index int) ASMType {
	translator := ""
	translator += "@LCL\n"                         // Load the base address of the local segment into the A-register
	translator += "D=M\n"                          // Load the value stored at the base address into the D-register
	translator += fmt.Sprintf("@%d", index) + "\n" // Load the index of the desired local variable into the A-register
	translator += "A=D+A\n"                        // Calculate the address of the desired local variable
	translator += "D=M\n"                          // Load the value of the desired local variable into the D-register
	translator += "@SP\n"
	translator += "A=M\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) popLocal(index int) ASMType {
	translator := ""
	translator += "@LCL\n"                         // Load the base address of the local segment into the A-register
	translator += "D=M\n"                          // Load the value stored at the base address into the D-register
	translator += fmt.Sprintf("@%d", index) + "\n" // Load the index of the desired local variable into the A-register
	translator += "D=D+A\n"
	translator += "@R13\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	translator += "@R13\n"
	translator += "A=M\n"
	translator += "M=D\n"
	return ASMType(translator)
}

func (c *CodeWriter) pushArgument(index int) ASMType {
	translator := ""
	translator += "@ARG\n"
	translator += "D=M\n"
	translator += fmt.Sprintf("@%d", index) + "\n"
	translator += "A=D+A\n"
	translator += "D=M\n"
	translator += "@SP\n"
	translator += "A=M\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) popArgument(index int) ASMType {
	translator := ""
	translator += "@ARG\n"
	translator += "D=M\n"
	translator += fmt.Sprintf("@%d", index) + "\n"
	translator += "D=D+A\n"
	translator += "@R13\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	translator += "@R13\n"
	translator += "A=M\n"
	translator += "M=D\n"
	return ASMType(translator)
}

func (c *CodeWriter) pushTemp(index int) ASMType {
	translator := ""
	translator += fmt.Sprintf("@%d", index) + "\n"
	translator += "D=A\n"
	translator += "@R5\n"
	translator += "A=D+A\n"
	translator += "D=M\n"
	translator += "@SP\n"
	translator += "A=M\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) popTemp(index int) ASMType {
	translator := ""
	translator += fmt.Sprintf("@%d", index) + "\n"
	translator += "D=A\n"
	translator += "@R5\n"
	translator += "D=D+A\n"
	translator += "@R13\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	translator += "@R13\n"
	translator += "A=M\n"
	translator += "M=D\n"
	return ASMType(translator)
}

func (c *CodeWriter) pushPointer(index int) ASMType {
	translator := ""

	if index == 0 {
		translator += "@THIS\n"
	} else if index == 1 {
		translator += "@THAT\n"
	} else {
		return ""
	}
	translator += "D=M\n" // D = THAT : load the value stored at the 'that' pointer into the D-register
	translator += "@SP\n"
	translator += "A=M\n"
	translator += "M=D\n"
	translator += "@SP\n"
	translator += "M=M+1\n"
	return ASMType(translator)
}

func (c *CodeWriter) popPointer(index int) ASMType {
	translator := ""
	translator += "@SP\n"
	translator += "AM=M-1\n"
	translator += "D=M\n"
	if index == 0 {
		translator += "@THIS\n"
	} else if index == 1 {
		translator += "@THAT\n"
	} else {
		return ""
	}
	translator += "M=D\n"
	return ASMType(translator)
}

func (c *CodeWriter) writerPushPop(command string, segment string, index int) (ASMType, error) {
	if command == "push" {
		switch segment {
		case "pointer":
			return c.pushPointer(index), nil
		case "this":
			return c.pushThis(index), nil
		case "that":
			return c.pushThat(index), nil
		case "static":
			return c.pushStatic(index), nil
		case "local":
			return c.pushLocal(index), nil
		case "argument":
			return c.pushArgument(index), nil
		case "temp":
			return c.pushTemp(index), nil
		case "constant":
			return c.pushConstant(index), nil
		}
	} else if command == "pop" {
		switch segment {
		case "pointer":
			return c.popPointer(index), nil
		case "this":
			return c.popThis(index), nil
		case "that":
			return c.popThat(index), nil
		case "static":
			return c.popStatic(index), nil
		case "local":
			return c.popLocal(index), nil
		case "argument":
			return c.popArgument(index), nil
		case "temp":
			return c.popTemp(index), nil
		}
	}
	return "", fmt.Errorf("The command not implemented yet")

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

func (c *CodeWriter) writeLabel(label string) ASMType {
	translator := ""
	translator += fmt.Sprintf("%s\n", label)
	return ASMType(translator)
}

func (c *CodeWriter) writeInit() ASMType {
	return ASMType("")
}

func (c *CodeWriter) writeGoto() ASMType {
	return ASMType("")
}

func (c *CodeWriter) writeIf() ASMType {
	return ASMType("")
}

func (c *CodeWriter) writeCall() ASMType {
	return ASMType("")
}

func (c *CodeWriter) writeReturn() ASMType {
	return ASMType("")
}

func (c *CodeWriter) writeFunction() ASMType {
	return ASMType("")
}

func main() {

}
