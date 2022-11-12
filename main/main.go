package main

import (
	"djson"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

var (
	file         string
	input        string
	outputFormat string
	indent       string
	bufSize      uint
)

func main() {
	flag.StringVar(&file, "f", "", "input pathfile")
	flag.StringVar(&input, "i", "", "input bytes")
	flag.StringVar(&outputFormat, "o", "json", "output format, current support: json, default is json")
	flag.UintVar(&bufSize, "b", 512, "buffer size, default is 512")
	flag.StringVar(&indent, "indent", "  ", "buffer size, default is \"  \"")
	flag.Parse()
	var r io.Reader
	var f *os.File
	var err error
	if input != "" {
		r = strings.NewReader(input)
	} else if file != "" {
		f, err = os.Open(file)
		if err != nil {
			fmt.Printf("can't open file: %s: %s", file, err.Error())
			os.Exit(1)
		}
		r = f
	} else {
		fmt.Printf("plz specific a file thru -f or a byte string with -i")
		os.Exit(1)
		return
	}
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	var encoder djson.Encoder
	switch outputFormat {
	case "json":
		encoder = djson.NewJsonEncoder(indent)
	default:
		fmt.Printf("decoder [%s] not support", outputFormat)
		os.Exit(1)
	}
	trans := djson.NewTranslator(encoder, djson.BuffSize(bufSize))
	if _, err := trans.Translate(r, os.Stdout); err != nil {
		fmt.Printf("translate failed: %s", err.Error())
		os.Exit(1)
	}
	os.Exit(0)
}
