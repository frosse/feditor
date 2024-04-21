package main

import (
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

func flush() {
	fmt.Print("\033[H\033[2J")

}

func move_cursor(y int, x int) {
	fmt.Printf("\033[%d;%dH", y, x)
}

type Cursor struct {
	x int
	y int
}

func (cursor *Cursor) move_cursor(x int, y int) {
	fmt.Printf("\033[%d;%dH", y, x)
	cursor.x = x
	cursor.y = y
}

func (cursor *Cursor) move_cursor_relative(x int, y int) {
	cursor.move_cursor(cursor.x+x, cursor.y+y)
}

func main() {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		panic(err)
	}

	defer term.Restore(int(os.Stdin.Fd()), oldState)

	flush()

	filename := os.Args[1]
	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Println(err.Error())
	}
	cursor := Cursor{0, 0}

	for _, line := range strings.Split(string(data), "\n") {
		fmt.Printf("%s\r\n", line)
	}
	cursor.move_cursor(0, 0)

	var b []byte = make([]byte, 1)
	for {
		os.Stdin.Read(b)
		if b[0] == 113 {
			break
		} else if b[0] == 27 && b[1] == 91 {
			switch b[2] {
			case 'A':
				cursor.move_cursor_relative(0, -1)
			case 'B':
				cursor.move_cursor_relative(0, 1)
			case 'C':
				cursor.move_cursor_relative(1, 0)
			case 'D':
				cursor.move_cursor_relative(-1, 0)
			}

		} else {
			print(string(b[0]))
		}
	}

}
