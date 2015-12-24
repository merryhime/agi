package main

import "os"
import "bufio"
import "fmt"

func main() {
	f, err := os.Open("./lexer.go")
	if err != nil {
		panic("Couldn't open main.go")
	}
	l := MakeLexer(bufio.NewReader(f), "main.go")
	for {
		t := l.NextToken()
		fmt.Printf("%v\n", t)
		if t.Type == EndOfFile {
			break
		}
	}
}
