package main

import (
	"bufio"
	"bytes"
	"log"
	"os"

	"fmt"
	"led/utils"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

// Editor is the initializing class
type Editor struct {
}

// Buffer is the temporary view presented to the user
type Buffer struct {
	Lines []string
}

// Screen is the rendering of the entire viewport
type Screen struct {
}

// Modeline is the message line at the bottom of the viewport
type Modeline struct {
	message string
}

// Statusline is the file path line at the bottom of the viewport
type Statusline struct {
	message    string
	fileSize   string
	filePath   string
	location   string
	fileFormat string
	format     string
	unformat   string
}

// Cursor is the object containing cursor point position
type Cursor struct {
	Row int
	Col int
}

type messageColor struct {
	foreground int
	background int
}

var arrow = map[string][]byte{
	"up":    []byte{0x1b, 0x5b, 0x41},
	"down":  []byte{0x1b, 0x5b, 0x42},
	"right": []byte{0x1b, 0x5b, 0x43},
	"left":  []byte{0x1b, 0x5b, 0x44},
}
var ctrl = map[string][]byte{
	"a": []byte{0x1}, // C-a
	"b": []byte{0x2},
	"c": []byte{0x3},
	"d": []byte{0x4},
	"e": []byte{0x5},
	"f": []byte{0x6},
	"g": []byte{0x7},
	"h": []byte{0x8},
	"i": []byte{0x9},
	"j": []byte{0xa},
	"k": []byte{0xb},
	"l": []byte{0xc},
	"m": []byte{0xd},
	"n": []byte{0xe},
	"o": []byte{0xf},
	"p": []byte{0x10},
	"q": []byte{0x11},
	"r": []byte{0x12},
	"s": []byte{0x13},
	"t": []byte{0x14},
	"u": []byte{0x15},
	"v": []byte{0x16},
	"w": []byte{0x17},
	"x": []byte{0x18},
	"y": []byte{0x19},
	"z": []byte{0x1a},
}
var meta = map[string][]byte{
	"x": []byte{0x1b, 0x78},
}

var filePath string
var editor Editor
var buffer Buffer
var screen Screen
var modeline Modeline
var statusline Statusline
var cursor Cursor
var multiplier int
var prefix []byte
var fileBytes int

func main() {
	if len(os.Args) == 1 {
		fmt.Println("You MUST supply a file to edit")
		return
	}
	filePath = os.Args[1]
	editor.initialize()
	editor.run()
}

func (e *Editor) initialize() {
	// Load a file
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}

	multiplier = 1
	statusline.setFormat()
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
	screen.render()
	moveCursor(cursor.Row, cursor.Col) // restore cursor
}

func (e *Editor) handleInput() {
	c := utils.Getch()
	// log.Printf("%#v\t%s\n", c, string(c))
	modeline.setMessage("")
	switch {
	case bytes.Equal(c, ctrl["x"]):
		reset()
		prefix = ctrl["x"]
		modeline.setMessage("C-x-")
	case bytes.Equal(c, ctrl["c"]):
		if bytes.Equal(prefix, ctrl["x"]) {
			clearScreen()
			moveCursor(1, 1)
			os.Exit(0)
			return
		}
		modeline.setMessage("Invalid command!")
		reset()
	case bytes.Equal(c, ctrl["s"]):
		if bytes.Equal(prefix, ctrl["x"]) {
			buffer.save()
			statusline.setFormat()
			modeline.setMessage(fmt.Sprintf("Saved \"%s\"", filePath))
			return
		}
		modeline.setMessage("Invalid command, did you mean to save -> C-x C-s")
		reset()
	case bytes.Equal(c, ctrl["g"]):
		modeline.setMessage("Quit")
		reset()
	case bytes.Equal(c, ctrl["e"]):
		cursorEOL()
		reset()
	case bytes.Equal(c, ctrl["a"]):
		cursorBOL()
		reset()
	case bytes.Equal(c, ctrl["u"]):
		setMultiplier()
	case bytes.Equal(c, arrow["up"]), bytes.Equal(c, ctrl["p"]): // UP, C-p
		cursorUp(multiplier)
		reset()
	case bytes.Equal(c, []byte{0x1b, 0x5b, 0x42}), bytes.Equal(c, []byte{0xe}): // DOWN, C-n
		cursorDown(multiplier)
		reset()
	case bytes.Equal(c, []byte{0x1b, 0x5b, 0x43}), bytes.Equal(c, []byte{0x6}): // RIGHT, C-f
		cursorForward(multiplier)
		reset()
	case bytes.Equal(c, []byte{0x1b, 0x5b, 0x44}), bytes.Equal(c, []byte{0x2}): // LEFT, C-b
		cursorBackward(multiplier)
		reset()
	case bytes.Equal(c, []byte{0x1b, 0x62}): // M-b
		cursorBackwardWord()
		reset()
	case bytes.Equal(c, []byte{0x1b, 0x66}): // M-f
		cursorForwardWord()
		reset()
	case bytes.Equal(c, []byte{0x7f}): // backspace
		buffer.deleteChar()
		reset()
	case bytes.Equal(c, []byte{0xb}): // C-k
		buffer.deleteForward()
		statusline.setFilePathColor(messageColor{1, 8})
		reset()
	default:
		buffer.insertChar(string(c))
		statusline.setFilePathColor(messageColor{1, 8})
		reset()
	}
}

func reset() {
	prefix = nil
	multiplier = 1
}

func setMultiplier() {
	if multiplier == 1 {
		multiplier = 4
		modeline.setMessage("C-u-")
	} else {
		multiplier = multiplier + multiplier
		msg := []string{}
		count := 0
		start := multiplier
		for start > 2 {
			start = start / 2
			count++
		}
		for i := 0; i < count; i++ {
			msg = append(msg, "C-u-")
		}

		modeline.setMessage(strings.Join(msg, ""))
	}
}

/**
 * BUFFER
 */
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
	fileBytes, err = file.WriteString(out)
	if err != nil {
		log.Fatal(err)
	}
}

func (b *Buffer) size() string {
	total := 0
	for _, v := range b.Lines {
		total += len(v)
	}
	return humanReadable(total)
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

/**
 * SCREEN
 */
func getTermSize() []int {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
	size := string(out)
	size = strings.TrimSpace(size)
	slice := strings.Split(size, " ")
	asRowInt, _ := strconv.ParseInt(slice[0], 0, 32)
	rows := int(asRowInt)
	asColInt, _ := strconv.ParseInt(slice[1], 0, 32)
	cols := int(asColInt)
	// fmt.Printf("%#v, %#v", size, []int{rows, cols})
	return []int{rows, cols}
}

func humanReadable(inp int) string {
	switch {
	case inp > 1000000:
		return fmt.Sprintf("%dM", inp/1000000)
	case inp > 1000:
		return fmt.Sprintf("%dk", inp/1000)
	default:
		return fmt.Sprintf("%d", inp)
	}
}

func (s *Screen) render() {
	termSize := getTermSize()

	buffer.render()

	statusline.setSize(buffer.size())
	statusline.setFilePath(filePath)
	statusline.setLocation(cursor.Row, cursor.Col)
	statusline.render(termSize)

	modeline.render(termSize)
}

/**
 * MODELINE
 */
func (m *Modeline) setMessage(msg string) {
	m.message = msg
}

func (m *Modeline) render(size []int) {
	if m.message != "" {
		moveCursor(size[0], 0)
		fmt.Printf("%s", m.message)
	}
}

/**
 * STATUSLINE
 */
func (s *Statusline) setFormat() {
	s.fileFormat = fmt.Sprintf("[48;5;8m[38;5;2m")
	s.format = fmt.Sprintf("[48;5;8m[38;5;7m")
	s.unformat = fmt.Sprintf("[48;5;0m[38;5;15m")
}
func (s *Statusline) setSize(fileSize string) {
	s.fileSize = fileSize
}
func (s *Statusline) setFilePath(filePath string) {
	array := []string{s.fileFormat, filePath, s.format}
	s.filePath = strings.Join(array, " ")
}
func (s *Statusline) setFilePathColor(color messageColor) {
	s.fileFormat = fmt.Sprintf("[48;5;%dm[38;5;%dm", color.background, color.foreground)
	// statusline.setFilePath(s.filePath)
}
func (s *Statusline) setLocation(row, col int) {
	s.location = fmt.Sprintf("%d:%d", row, col)
}
func (s *Statusline) render(size []int) {
	moveCursor(size[0]-1, 0)
	statuslineParts := []string{s.fileSize, s.fileFormat, s.filePath, s.format, s.location}
	line := strings.Join(statuslineParts, " ")
	addLength := len(s.fileFormat) + len(s.format) + len(s.unformat)
	fmt.Printf("%s%-*s%s", s.format, size[1]+addLength-1, line, s.unformat)
}

/**
 * CURSOR
 */
func (c *Cursor) new() Cursor {
	return Cursor{
		1,
		1,
	}
}

// func clamp(max, cur int) int {
//  if cur <= max {
//      return cur
//  }
//  return max
// }

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
