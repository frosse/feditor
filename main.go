package main

import (
	"fmt"
	"os"
	"slices"
	"strings"

	"golang.org/x/term"
)

func move_cursor(y int, x int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

type Cursor struct {
	x int
	y int
}

func (cursor *Cursor) move_cursor(x int, y int) {
	if y < 1 {
		fmt.Printf("\033[%d;%dH", 1, x)
	} else if x < 1 {
		fmt.Printf("\033[%d;%dH", y, 1)
	} else {
		fmt.Printf("\033[%d;%dH", y, x)
		cursor.x = x
		cursor.y = y
	}
}

func (cursor *Cursor) move_cursor_relative(x, y int) {
	cursor.move_cursor(cursor.x+x, cursor.y+y)
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
	editor.flush()
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
	cursorX := e.cursor.x
	cursorY := e.cursor.y
	line := e.data[e.cursor.y-1]
	e.data[e.cursor.y-1] = slices.Insert(line, e.cursor.x-1, string(char))
	e.drawEditor()
	// How to handle array out of bounds
	e.cursor.move_cursor(cursorX+1, cursorY)
}

func (e *Editor) backspace() {
	cursorX := e.cursor.x
	cursorY := e.cursor.y
	line := e.data[e.cursor.y-1]
	e.data[e.cursor.y-1] = slices.Delete(line, cursorX-2, cursorX-1)
	e.drawEditor()
	// How to handle array out of bounds
	e.cursor.move_cursor(cursorX-1, cursorY)
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

	e.cursor.move_cursor(1, cursorY+1)
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

func main() {

	filename := os.Args[1]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
	}

	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState)

	editor := NewEditor(filename)
	defer editor.flush()

	editor.initData(data)

	editor.drawEditor()

	editor.cursor.move_cursor(1, 1)

	for {
		var b []byte = make([]byte, 1)
		os.Stdin.Read(b)

		// handle escape char
		if b[0] == 27 {
			var seq []byte = make([]byte, 2)
			os.Stdin.Read(seq)

			if seq[0] == 91 {
				switch seq[1] {
				case 'A':
					editor.cursor.move_cursor_relative(0, -1)
				case 'B':
					editor.cursor.move_cursor_relative(0, 1)
				case 'C':
					editor.cursor.move_cursor_relative(1, 0)
				case 'D':
					editor.cursor.move_cursor_relative(-1, 0)
				}
			}
		} else if b[0] == 127 {
			editor.backspace()
		} else if b[0] == 13 {
			editor.enter()
		} else if int(b[0]) == is_ctrl('q') {
			break
		} else if int(b[0]) == is_ctrl('s') {
			editor.save()
		} else {
			editor.writeChar(b)
		}
	}

}
