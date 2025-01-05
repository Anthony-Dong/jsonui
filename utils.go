package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/mattn/go-runewidth"
)

func Printf(w io.Writer, format string, args ...interface{}) {
	if len(args) == 0 {
		w.Write([]byte(terminalLine(format)))
		return
	}
	w.Write([]byte(terminalLine(fmt.Sprintf(format, args...))))
}

func Println(w io.Writer, format string, args ...interface{}) {
	if len(args) == 0 {
		w.Write([]byte(terminalLine(format)))
		w.Write([]byte{'\n'})
		return
	}
	w.Write([]byte(terminalLine(fmt.Sprintf(format, args...))))
	w.Write([]byte{'\n'})
}

func terminalLine(line string) string {
	var lineWithWidth = bytes.NewBuffer(make([]byte, 0, len(line)))
	for _, r := range line {
		w := runewidth.RuneWidth(r)
		if w == 0 {
			w = 1
		}
		lineWithWidth.WriteString(string(r))
		for i := 1; i < w; i++ {
			lineWithWidth.WriteString(" ")
		}
	}
	return lineWithWidth.String()
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

// 检测标准输入是否来自管道
func checkStdInFromPiped() bool {
	if stat, _ := os.Stdin.Stat(); (stat.Mode() & os.ModeCharDevice) == 0 {
		return true
	} else {
		return false
	}
}

type cacheData struct {
	cache     map[string]interface{}
	cacheList []string
	maxSize   int
}

func newCacheData(size int) *cacheData {
	return &cacheData{
		cache:     make(map[string]interface{}, size),
		cacheList: make([]string, 0, size),
		maxSize:   size,
	}
}

func (c *cacheData) get(key string) interface{} {
	return c.cache[key]
}

func (c *cacheData) getOrStore(key string, load func() interface{}) (interface{}, bool) {
	if c.maxSize == 0 {
		return load(), true
	}
	v, isOk := c.cache[key]
	if isOk {
		return v, false
	}
	value := load()
	c.set(key, value)
	return value, true
}

func (c *cacheData) set(key string, value interface{}) {
	if c.maxSize > 1 && len(c.cacheList) == c.maxSize {
		delete(c.cache, c.cacheList[0])
		c.cacheList = c.cacheList[1:]
	}
	c.cacheList = append(c.cacheList, key)
	c.cache[key] = value
}
