package main

import (
	"encoding/json"
	"fmt"
)

type Protocol struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func encode(v *Protocol) []byte {
	bytes, _ := json.Marshal(v)
	return bytes
}

func decode(bytes []byte) (v *Protocol, err error) {
	v = new(Protocol)
	err = json.Unmarshal(bytes, v)
	return
}

func Go(f func() error) chan error {
	ch := make(chan error)
	go func() {
		ch <- f()
	}()
	return ch
}

// ConsoleWriter
type ConsoleWriter struct {
	wfunc func(data string)
}

func (w *ConsoleWriter) Write(data []byte) (n int, err error) {
	fmt.Println(string(data))
	return len(data), nil
}

func NewConsoleWriter(f func(data string)) *ConsoleWriter {
	return &ConsoleWriter{wfunc: f}
}
