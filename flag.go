package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type helpMsg struct {
	Header string
	flags  []helpMsgFlags
}

func NewHelpMsg(header string) *helpMsg {
	return &helpMsg{Header: header}
}

func (h *helpMsg) addFlag(flag string, desc string) {
	h.flags = append(h.flags, helpMsgFlags{flag, desc})
}

type helpMsgFlags struct {
	Flag string
	Desc string
}

func (h *helpMsg) String() string {
	flagLenSize := 0
	descLenSize := 0
	for _, flag := range h.flags {
		if len(flag.Flag) > flagLenSize {
			flagLenSize = len(flag.Flag)
		}
		if len(flag.Desc) > descLenSize {
			descLenSize = len(flag.Desc)
		}
	}
	out := bytes.NewBuffer(nil)
	out.WriteString(h.Header)
	out.WriteString("\n")
	out.WriteString(strings.Repeat("-", flagLenSize))
	out.WriteString(strings.Repeat("-", descLenSize))
	out.WriteString("---")
	out.WriteString("\n")

	padding := func(k string, size int) string {
		padding := size - len(k)
		if padding > 0 {
			return k + strings.Repeat(" ", padding)
		}
		return k
	}
	for _, elem := range h.flags {
		out.WriteString(padding(elem.Flag, flagLenSize))
		out.WriteString(" = ")
		out.WriteString(padding(elem.Desc, descLenSize))
		out.WriteString("\n")
	}
	return out.String()
}

type flagArgs struct {
	File string `json:"file"`
}

func initFlag() *flagArgs {
	result := &flagArgs{}
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "JSONUI Usage: %s [-r file]\n", filepath.Base(os.Args[0]))
	}
	flag.StringVar(&result.File, "r", "", "File to read from")
	flag.Parse()
	return result
}

func initHelpMsg() *helpMsg {
	msg := NewHelpMsg("JSONUI - Help")
	msg.addFlag("ctrl+e/ArrowDown", "Move a line down")
	msg.addFlag("ctrl+y/ArrowUp", "Move a line up")
	msg.addFlag("ctrl+d", "Move 15 line down")
	msg.addFlag("ctrl+u", "Move 15 line up")
	msg.addFlag("ctrl+f", "PageDown")
	msg.addFlag("ctrl+b", "PageUp")
	msg.addFlag("c", "Copy node value")
	msg.addFlag("f", "Format node data")
	msg.addFlag("q/ctrl+c", "Exit")
	msg.addFlag("h/?", "Toggle help message")
	return msg
}
