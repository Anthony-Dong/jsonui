package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/jroimartin/gocui"
	"github.com/mattn/go-runewidth"
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

func FormatData(input string) string {
	if input == "" {
		return input
	}
	input = strings.TrimSpace(input)
	if input[0] == '"' {
		unquote, err := strconv.Unquote(input)
		if err != nil {
			return input
		}
		jsonData, err := prettyJson(unquote)
		if err != nil {
			return unquote
		}
		return jsonData
	}
	return input
}

func prettyJson(src string) (string, error) {
	out := bytes.Buffer{}
	if err := json.Indent(&out, []byte(src), "", "  "); err != nil {
		return "", err
	}
	return out.String(), nil
}

func Printf(w io.Writer, format string, args ...interface{}) {
	if len(args) == 0 {
		w.Write([]byte(TerminalString(format)))
		return
	}
	w.Write([]byte(TerminalString(fmt.Sprintf(format, args...))))
}

func Println(w io.Writer, format string, args ...interface{}) {
	if len(args) == 0 {
		w.Write([]byte(TerminalString(format)))
		w.Write([]byte{'\n'})
		return
	}
	w.Write([]byte(TerminalString(fmt.Sprintf(format, args...))))
	w.Write([]byte{'\n'})
}
