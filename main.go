package main

import (
	"bufio"
	"bytes"
	"log"
	"os"

	"editor/utils"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"regexp"
	"strings"
)

// Editor is the initializing class
type Editor struct {
}

// Buffer is the temporary view presented to the user
type Buffer struct {
	Lines []string
}

// Cursor is the object containing cursor point position
type Cursor struct {
	Row int
	Col int
}

func clamp(max, cur int) int {
	if cur <= max {
		return cur
	}
	return max
}

var filePath string
var editor Editor
var buffer Buffer
var cursor Cursor
var multiplier int

func main() {
	if len(os.Args) == 1 {
		log.Println("a file is required")
		return
	}
	filePath = os.Args[1]
	editor.initialize()
	editor.run()
}

func (e *Editor) initialize() {
	// Load a file
	file, err := os.Open(filePath)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	// RAW terminal
	oldState, err := terminal.MakeRaw(0)
	if err != nil {
		log.Fatal(err)
	}
	defer terminal.Restore(0, oldState)

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	multiplier = 1
	buffer = buffer.new(lines)
	cursor = cursor.new()
}

func (e *Editor) run() {
	for {
		e.render()
		e.handleInput()
	}
}

func (e *Editor) render() {
	clearScreen()
	buffer.render()

	moveCursor(cursor.Row, cursor.Col) // restore cursor
}

func (e *Editor) handleInput() {
	c := utils.Getch()
	// log.Printf("%#v\t%s\n", c, string(c))
	switch {
	case bytes.Equal(c, []byte{0x3}), bytes.Equal(c, []byte{0x11}), bytes.Equal(c, []byte{0x18}): // C-c, C-q, C-x
		// quit
		clearScreen()
		moveCursor(1, 1)
		os.Exit(0)
		return
	case bytes.Equal(c, []byte{0x13}): // C-s
		// save
		buffer.save()
	case bytes.Equal(c, []byte{0x5}): // C-e
		// end of line
		cursorEOL()
	case bytes.Equal(c, []byte{0x1}): // C-a
		// beginning of line
		cursorBOL()
	case bytes.Equal(c, []byte{0x15}): // C-u
		if multiplier == 1 {
			multiplier = 4
		} else {
			multiplier = multiplier + multiplier
		}
	case bytes.Equal(c, []byte{0x1b, 0x5b, 0x41}), bytes.Equal(c, []byte{0x10}): // UP, C-p
		cursorUp(multiplier)
		multiplier = 1
	case bytes.Equal(c, []byte{0x1b, 0x5b, 0x42}), bytes.Equal(c, []byte{0xe}): // DOWN, C-n
		cursorDown(multiplier)
		multiplier = 1
	case bytes.Equal(c, []byte{0x1b, 0x5b, 0x43}), bytes.Equal(c, []byte{0x6}): // RIGHT, C-f
		cursorForward(multiplier)
		multiplier = 1
	case bytes.Equal(c, []byte{0x1b, 0x5b, 0x44}), bytes.Equal(c, []byte{0x2}): // LEFT, C-b
		cursorBackward(multiplier)
		multiplier = 1
	case bytes.Equal(c, []byte{0x1b, 0x62}): // M-b
		cursorBackwardWord()
		multiplier = 1
	case bytes.Equal(c, []byte{0x1b, 0x66}): // M-f
		cursorForwardWord()
		multiplier = 1
	case bytes.Equal(c, []byte{0x7f}): // backspace
		buffer.deleteChar()
		multiplier = 1
	case bytes.Equal(c, []byte{0xb}): // C-k
		buffer.deleteForward()
		multiplier = 1
	default:
		buffer.insertChar(string(c))
		multiplier = 1
	}
}

func (b *Buffer) new(lines []string) Buffer {
	b.Lines = lines

	return *b
}

func (b *Buffer) fetch(line int) string {
	// fix the index 1 issue
	if line > len(b.Lines) {
		return ""
	}
	return b.Lines[line-1]
}

func (b *Buffer) render() {
	moveCursor(1, 1) // reset cursor for printing buffer
	for _, str := range buffer.Lines {
		// fmt.Printf("%d| %s\r\n", num, str) // with line nums
		fmt.Printf("%s\r\n", str)
	}
}

func (b *Buffer) save() {
	// save file
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()
	out := strings.Join(b.Lines, "\n")
	_, err = file.WriteString(out)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Buffer) insertChar(inp string) {
	if len(b.Lines) < cursor.Row {
		b.Lines = append(b.Lines, inp)
	} else {
		b.Lines[cursor.Row-1] = strings.Join(
			[]string{
				b.Lines[cursor.Row-1][:cursor.Col-1],
				inp,
				b.Lines[cursor.Row-1][cursor.Col-1:],
			}, "")
	}
	cursor.Col++
	moveCursor(cursor.Row, cursor.Col)
}

func (b *Buffer) deleteChar() {
	if cursor.Col == 1 {
		return
	}
	b.Lines[cursor.Row-1] = strings.Join(
		[]string{
			b.Lines[cursor.Row-1][:cursor.Col-2],
			b.Lines[cursor.Row-1][cursor.Col-1:],
		}, "")
	cursor.Col--
	moveCursor(cursor.Row, cursor.Col)
}

func (b *Buffer) deleteForward() {
	b.Lines[cursor.Row-1] = b.Lines[cursor.Row-1][:cursor.Col-1]
}

func (c *Cursor) new() Cursor {
	return Cursor{
		1,
		1,
	}
}

func (c *Cursor) clampCol() {
	line := buffer.fetch(c.Row)
	if c.Col > len(line) {
		c.Col = len(line) + 1
	} else if c.Col < 1 {
		c.Col = 1
	}
}

func (c *Cursor) clampRow() {
	rows := len(buffer.Lines)
	if c.Row > rows+1 {
		c.Row = rows + 1
	} else if c.Row < 1 {
		c.Row = 1
	}
}

//
//
//

func clearScreen() {
	fmt.Print("[2J")
}
func moveCursor(row, col int) {
	fmt.Printf("[%d;%dH", row, col)
}

func cursorEOL() {
	line := buffer.fetch(cursor.Row)
	cursor.Col = len(line) + 1
	moveCursor(cursor.Row, cursor.Col)
}

func cursorBOL() {
	cursor.Col = 1
	moveCursor(cursor.Row, cursor.Col)
}

func cursorUp(multiplier int) {
	for k := 0; k < multiplier; k++ {
		cursor.Row = cursor.Row - 1
	}
	cursor.clampRow()
	cursor.clampCol()
	moveCursor(cursor.Row, cursor.Col)
}

func cursorDown(multiplier int) {
	for k := 0; k < multiplier; k++ {
		cursor.Row = cursor.Row + 1
	}
	cursor.clampRow()
	cursor.clampCol()
	moveCursor(cursor.Row, cursor.Col)
}

func cursorForward(multiplier int) {
	for k := 0; k < multiplier; k++ {
		cursor.Col = cursor.Col + 1
	}
	cursor.clampCol()
	cursor.clampRow()
	moveCursor(cursor.Row, cursor.Col)
}

func cursorForwardWord() {
	// split line into array
	line := buffer.fetch(cursor.Row)
	line = line[cursor.Col-1:]

	content := []byte(line)
	pattern := regexp.MustCompile(`\W\w`)
	loc := pattern.FindIndex(content)
	if loc != nil {
		cursor.Col = cursor.Col + loc[0] + 1

		cursor.clampCol()
		cursor.clampRow()
		moveCursor(cursor.Row, cursor.Col)
	}
}

func cursorBackward(multiplier int) {
	for k := 0; k < multiplier; k++ {
		cursor.Col = cursor.Col - 1
	}
	cursor.clampCol()
	cursor.clampRow()
	moveCursor(cursor.Row, cursor.Col)
}

func cursorBackwardWord() {
	line := buffer.fetch(cursor.Row)
	line = strings.TrimRight(line[:cursor.Col-1], " ")

	content := []byte(line)
	pattern := regexp.MustCompile(`\w+.?$`)
	loc := pattern.FindIndex(content)
	if loc != nil {
		cursor.Col = loc[0] + 1

		cursor.clampCol()
		cursor.clampRow()
		moveCursor(cursor.Row, cursor.Col)
	}
}
