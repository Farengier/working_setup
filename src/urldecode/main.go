package main

import (
	"flag"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
)

var version = "dev"

func main() {
	versionFlag := flag.Bool("v", false, "вывести версию программы")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	// Читаем всё из stdin
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения stdin: %v", err)
		os.Exit(1)
	}

	strInput := strings.TrimSpace(string(input))
	decoded, err := url.QueryUnescape(strInput)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка urldecode: %v", err)
		os.Exit(1)
	}
	decoded = strings.ReplaceAll(decoded, " ", "+")
	os.Stdout.WriteString(decoded)
}
