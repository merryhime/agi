package main

import "os"
import "bufio"
import "fmt"
import "github.com/MerryMage/agi/lexer"

func main() {
	f, err := os.Open("./lexer.go")
	if err != nil {
		panic("Couldn't open main.go")
	}
	l := lexer.MakeLexer(bufio.NewReader(f), "main.go")
	for {
		t := l.NextToken()
		fmt.Printf("%v\n", t)
		if t.Type == lexer.EndOfFile {
			break
		}
	}
}
