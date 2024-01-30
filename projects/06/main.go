package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	A_COMMANDS = 1
	C_COMMANDS = 2
	L_COMMANDS = 3
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

func commandType(line string) int {
	if line != "" && strings.HasPrefix(line, "@") {
		return A_COMMANDS
	} else if line != "" && strings.HasPrefix(line, "(") {
		return L_COMMANDS
	}
	return C_COMMANDS
}

// func symbol(line string) string {

// }

func buildSymbol(lines []string) (map[string]int, []string) {
	cleanContent := []string{}
	for index, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" && strings.HasPrefix(trimmedLine, "(") {
			symbolTable[trimmedLine[1:len(trimmedLine)-1]] = index
		} else {
			cleanContent = append(cleanContent, line)
		}
	}
	return symbolTable, cleanContent
}

func readAsm(path string) ([]string, error) {
	data, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	lines := []string{}
	scanner := bufio.NewScanner(data)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return lines, nil
}

func removeComment(lines []string) []string {
	cleanContent := []string{}
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" && !strings.HasPrefix(trimmedLine, "//") {
			if index := strings.Index(trimmedLine, "//"); index != -1 {
				trimmedLine = trimmedLine[:index]
			}
			cleanContent = append(cleanContent, trimmedLine)
		}
	}
	return cleanContent
}

func advance(line string) {

}

func hasMoreCommands(string) bool {
	return true
}

func convertNumberToBinary(num int64) string {
	binaryString := strconv.FormatInt(int64(num), 2)
	paddedBinaryString := fmt.Sprintf("%016s", binaryString)
	return paddedBinaryString
}

func main() {
	if len(os.Args) != 2 {
		return
	}
	// res := []string{}
	lines, err := readAsm(os.Args[1])
	if err != nil {
		return
	}
	// Remove comments and spaces
	lines = removeComment(lines)
	// Build symbol Table
	symbolTable, lines = buildSymbol(lines)

	// Translate instructions
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" && strings.HasPrefix(trimmedLine, "@") {

		} else {

		}
	}

}
