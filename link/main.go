package main

import (
	"flag"
	"fmt"
	"os"

	link "link/parser"
)

func main() {
	htmlFile := flag.String("i", "./ex1.html", "Input html file to parse")
	flag.Parse()

	file, err := os.Open(*htmlFile)
	if err != nil {
		panic(err)
	}

	res, err := link.Parse(file)
	if err != nil {
		panic(err)
	}

	fmt.Printf("%+v\n", res)
}
