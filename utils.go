package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"

	"github.com/gobuild/log"
)

type Protocol struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func Go(f func() error) chan error {
	ch := make(chan error)
	go func() {
		ch <- f()
	}()
	return ch
}

func NewJsonStream(r io.Reader) chan *Protocol {
	ch := make(chan *Protocol, 1)
	go func() {
		decoder := json.NewDecoder(r)
		for {
			v := new(Protocol)
			if err := decoder.Decode(v); err != nil {
				break
			}
			ch <- v
		}
		close(ch)
	}()
	return ch
}

func NewJsonSender(w io.Writer) func(name string, data interface{}) error {
	encoder := json.NewEncoder(w)
	lock := &sync.Mutex{}
	return func(name string, data interface{}) error {
		lock.Lock()
		defer lock.Unlock()
		log.Debug("send", name, fmt.Sprintf("%#v", data))
		err := encoder.Encode(&Protocol{Name: name, Data: fmt.Sprintf("%v", data)})
		if err != nil {
			log.Warn("send error", err)
		}
		return err
	}
}

func ReadExitcode(err error) int {
	if err == nil {
		return 0
	}
	if exiterr, ok := err.(*exec.ExitError); ok {
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return 126
}

func TrapSignal(f func(sig os.Signal), sigs ...os.Signal) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, sigs...)
	go func() {
		for sig := range sigCh {
			f(sig)
		}
	}()
}

// ---------- func writer --------------
type FuncWriter struct {
	fwr func(data string) error
}

func (fw *FuncWriter) Write(data []byte) (n int, err error) {
	err = fw.fwr(string(data))
	return len(data), err
}

func NewFuncWriter(f func(data string) error) *FuncWriter {
	return &FuncWriter{f}
}
