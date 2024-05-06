package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/term"
)

const (
	EscapeChar = 27
	Backspace  = 127
	Enter      = 13
	Tabulator  = 9
)

type Cursor struct {
	x int
	y int
}

func (e *Editor) move_cursor_x(amount int) {
	newPos := e.cursor.x + amount
	if newPos < 1 {
		e.move_cursor_abs(1, e.cursor.y)
	} else if newPos > len(e.data[e.cursor.y-1])+1 {
		return
	} else {
		e.move_cursor_abs(newPos, e.cursor.y)
	}
}

func (e *Editor) move_cursor_y(amount int) {
	newPos := e.cursor.y + amount
	if newPos < 1 {
		e.move_cursor_abs(e.cursor.x, 1)
	} else if len(e.data[newPos-1]) < e.cursor.x-1 {
		xPos := len(e.data[newPos-1]) + 1
		e.move_cursor_abs(xPos, newPos)
	} else {
		e.move_cursor_abs(e.cursor.x, newPos)
	}
}

func (e *Editor) move_cursor_abs(x, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
	e.cursor.x = x
	e.cursor.y = y
}

func (e *Editor) move_cursor_relative(x, y int) {
	e.move_cursor_abs(e.cursor.x+x, e.cursor.y+y)
}

func is_ctrl(buf rune) int {
	return int(buf) & 31
}

type Editor struct {
	data     [][]string
	cursor   Cursor
	filename string
}

func NewEditor(filename string) *Editor {
	cursor := Cursor{1, 1}
	editor := Editor{
		cursor:   cursor,
		filename: filename,
	}

	data, err := os.ReadFile(filename)

	if err != nil {
		fmt.Println(err.Error())
	}

	editor.flush()
	editor.initData(data)
	editor.drawEditor()
	editor.move_cursor_abs(1, 1)

	return &editor
}

func (e Editor) flush() {
	fmt.Print("\033[H\033[2J")
}

func (e *Editor) initData(data []byte) {
	for _, line := range strings.Split(string(data), "\n") {
		e.data = append(e.data, strings.Split(line, ""))
	}
}

func (e *Editor) drawEditor() {
	e.flush()
	for _, line := range e.data {
		for _, c := range line {
			fmt.Printf("%s", c)
		}
		fmt.Printf("\r\n")
	}
}

func (e *Editor) writeChar(char []byte) {
	line := e.data[e.cursor.y-1]
	e.data[e.cursor.y-1] = slices.Insert(line, e.cursor.x-1, string(char))
	e.drawEditor()
	e.move_cursor_x(1)
}

func (e *Editor) backspace() {
	cursorX := e.cursor.x
	if cursorX <= 1 {
		if e.cursor.y <= 1 {
			return
		} else {
			currentLine := e.data[e.cursor.y-1]
			length := len(e.data[e.cursor.y-2]) + 1
			e.data[e.cursor.y-2] = append(e.data[e.cursor.y-2], currentLine...)
			e.data = slices.Delete(e.data, e.cursor.y-1, e.cursor.y)
			e.drawEditor()
			e.move_cursor_abs(length, e.cursor.y-1)
		}
	} else {
		line := e.data[e.cursor.y-1]
		e.data[e.cursor.y-1] = slices.Delete(line, cursorX-2, cursorX-1)
		e.drawEditor()
		e.move_cursor_x(-1)
	}
}

func (e *Editor) enter() {
	cursorY := e.cursor.y
	cursorX := e.cursor.x

	currentLine := e.data[cursorY-1]

	newLine := make([]string, len(currentLine[cursorX-1:]))
	copy(newLine, currentLine[cursorX-1:])
	currentLine = currentLine[:cursorX-1]

	// Not sure about this
	e.data = append(e.data[:cursorY], append([][]string{newLine}, e.data[cursorY:]...)...)

	e.data[cursorY-1] = currentLine

	e.drawEditor()

	e.move_cursor_abs(1, cursorY+1)
}

func (e *Editor) save() {
	file, err := os.Create(e.filename)

	if err != nil {
		panic(err)
	}
	defer file.Close()

	for _, line := range e.data {
		file.WriteString(strings.Join(line, ""))
		file.WriteString("\n")
	}
}

func (e *Editor) run() {
	for {
		var b []byte = make([]byte, 1)
		os.Stdin.Read(b)

		switch b[0] {
		case EscapeChar:
			var seq []byte = make([]byte, 2)
			os.Stdin.Read(seq)

			if seq[0] == 91 {
				switch seq[1] {
				case 'A':
					e.move_cursor_y(-1)
				case 'B':
					e.move_cursor_y(1)
				case 'C':
					e.move_cursor_x(1)
				case 'D':
					e.move_cursor_x(-1)
				}
			}
		case Backspace:
			e.backspace()
		case Enter:
			e.enter()
		case byte(is_ctrl('q')):
			return
		case byte(is_ctrl('s')):
			e.save()
		default:
			e.writeChar(b)
		}
	}
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)

	filename := os.Args[1]
	editor := NewEditor(filename)
	defer editor.flush()

	editor.run()
}
