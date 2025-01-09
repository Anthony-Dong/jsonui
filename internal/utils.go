package internal

import (
	"bytes"
	"context"
	"fmt"
	"github.com/jroimartin/gocui"
	"github.com/mattn/go-runewidth"
	"log"
	"reflect"
	"time"
	"unsafe"
)

func TerminalBytes(data []byte) []byte {
	return String2Bytes(TerminalString(Bytes2String(data)))
}

func TerminalString(line string) string {
	lineWithWidth := bytes.NewBuffer(make([]byte, 0, len(line)+1024))
	for _, r := range line {
		w := runewidth.RuneWidth(r)
		if w == 0 {
			w = 1
		}
		lineWithWidth.WriteRune(r)
		for i := 1; i < w; i++ {
			lineWithWidth.WriteString(" ")
		}
	}
	return lineWithWidth.String()
}

func String2Bytes(data string) []byte {
	hdr := *(*reflect.StringHeader)(unsafe.Pointer(&data))
	return *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
		Data: hdr.Data,
		Len:  hdr.Len,
		Cap:  hdr.Len,
	}))
}

func Bytes2String(data []byte) string {
	hdr := *(*reflect.SliceHeader)(unsafe.Pointer(&data))
	return *(*string)(unsafe.Pointer(&reflect.StringHeader{
		Data: hdr.Data,
		Len:  hdr.Len,
	}))
}

func RunWithTimeout(ctx context.Context, timeout time.Duration, fn func() error) (err error) {
	errChan := make(chan error, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				errChan <- fmt.Errorf("RunWithTimeout panic: %v", r)
			}
			close(errChan)
		}()
		if bizErr := fn(); bizErr != nil {
			errChan <- bizErr
		}
	}()
	select {
	case err = <-errChan:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("timeout")
	}
}

func MultiSetKeybinding(g *gocui.Gui, viewName string, keys []interface{}, handler func(*gocui.Gui, *gocui.View) error) {
	for _, key := range keys {
		if err := g.SetKeybinding(viewName, key, gocui.ModNone, handler); err != nil {
			log.Panicln(err)
		}
	}
}

func Quit(*gocui.Gui, *gocui.View) error {
	return gocui.ErrQuit
}
