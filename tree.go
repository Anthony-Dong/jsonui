package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/anthony-dong/jsonui/internal/orderedmap"
)

const (
	treeSignDash     = "─"
	treeSignVertical = "│"
	treeSignUpMiddle = "├"
	treeSignUpEnding = "└"
)

type treePosition []string

func (t treePosition) empty() bool {
	return len(t) == 0
}
func (t treePosition) shift() treePosition {
	newLength := len(t) - 1
	newPosition := make([]string, newLength)
	for i := 0; i < newLength; i++ {
		newPosition[i] = t[i+1]
	}
	return newPosition
}

type query struct {
	q string
}

type treeNode interface {
	String(int) string
	draw(w io.Writer, level int) error
	filter(query query) bool
	find(treePosition) treeNode
	search(query string) (treeNode, error)
	isCollapsable() bool
	toggleExpanded()
	collapseAll()
	expandAll()
	isExpanded() bool
}

type baseTreeNode struct {
	expanded bool
}

func (n *baseTreeNode) isExpanded() bool {
	return n.expanded
}

func (n *baseTreeNode) toggleExpanded() {
	n.expanded = !n.expanded
}

func (n baseTreeNode) expIcon() string {
	if n.expanded {
		return "[+]"
	}
	return "[-]"
}

type complexNode struct {
	baseTreeNode
	data *orderedmap.OrderedMap
	raw  *orderedmap.OrderedMap
}

func (n *complexNode) collapseAll() {
	n.expanded = false
	n.data.Foreach(func(key string, value interface{}) {
		value.(treeNode).collapseAll()
	})
}

func (n *complexNode) expandAll() {
	n.expanded = true
	n.data.Foreach(func(key string, value interface{}) {
		value.(treeNode).expandAll()
	})
}
func (n complexNode) isCollapsable() bool {
	return true
}
func (n complexNode) search(query string) (treeNode, error) {
	filteredNode := &complexNode{
		baseTreeNode: baseTreeNode{true},
		data:         orderedmap.New(),
	}
	n.data.Foreach(func(key string, value interface{}) {
		if key == query {
			filteredNode.data.Set(key, value)
		}
	})
	return filteredNode, nil
}

func (n complexNode) get(key string) (treeNode, bool) {
	v, isOk := n.data.Get(key)
	if !isOk {
		return nil, false
	}
	return v.(treeNode), true
}

func (n complexNode) keys() []string {
	return n.data.Keys()
}

func (n complexNode) find(tp treePosition) treeNode {
	if tp.empty() {
		return &n
	}
	e, ok := n.get(tp[0])
	newTp := tp.shift()
	if !ok {
		// This can't happen in theory
		return nil
	}
	if newTp.empty() {
		return e
	}
	return e.find(newTp)
}

func (n complexNode) String(indent int) string {
	if n.raw != nil {
		return encodeJson(n.raw, indent)
	}
	result := orderedmap.NewWithSize(n.data.Size())
	result.SetUseNumber(true)
	result.SetEscapeHTML(true)
	n.data.Foreach(func(key string, value interface{}) {
		data := value.(treeNode).String(indent)
		result.Set(key, json.RawMessage(data))
	})
	return encodeJson(result, indent)
}

func (n complexNode) draw(writer io.Writer, level int) error {
	if level == 0 {
		fmt.Fprintf(writer, "%s\n", "root")
	}
	keys := n.keys()
	length := len(keys)
	for i, key := range keys {
		value, _ := n.get(key)
		var char string
		if i < length-1 {
			char = treeSignUpMiddle
		} else {
			char = treeSignUpEnding
		}
		char += treeSignDash
		expendedCharacter := ""
		if value.isCollapsable() && !value.isExpanded() {
			expendedCharacter += " (+)"
		}
		fmt.Fprintf(writer,
			"%s%s %s%s\n",
			strings.Repeat("│  ", level),
			char,
			key,
			expendedCharacter,
		)
		if value.isExpanded() {
			value.draw(writer, level+1)
		}
	}
	return nil

}
func (n complexNode) filter(query query) bool {
	return true

}

type listNode struct {
	baseTreeNode
	data []treeNode
	raw  []interface{}
}

func (n *listNode) collapseAll() {
	n.expanded = false
	for _, v := range n.data {
		v.collapseAll()
	}
}
func (n *listNode) expandAll() {
	n.expanded = true
	for _, v := range n.data {
		v.expandAll()
	}
}
func (n listNode) isCollapsable() bool {
	return true
}

func (n listNode) search(query string) (treeNode, error) {
	return nil, nil

}
func (n listNode) find(tp treePosition) treeNode {
	if tp.empty() {
		return &n
	}
	i, err := strconv.Atoi(strings.TrimSuffix(strings.TrimPrefix(tp[0], "["), "]"))
	if err != nil {
		return nil
	}
	newTp := tp.shift()
	if newTp.empty() {
		return n.data[i]
	}
	return n.data[i].find(newTp)
}

func (n listNode) String(indent int) string {
	return encodeJson(n.raw, indent)
}

func (n listNode) draw(writer io.Writer, level int) error {
	if level == 0 {
		fmt.Fprintf(writer, "%s\n", "root")
	}
	length := len(n.data)
	for i, value := range n.data {
		var char string
		if i < length-1 {
			char = treeSignUpMiddle
		} else {
			char = treeSignUpEnding
		}
		char += treeSignDash
		expendedCharacter := ""

		if value.isCollapsable() && !value.isExpanded() {
			expendedCharacter += " (+)"
		}

		fmt.Fprintf(writer,
			"%s%s [%d]%s\n",
			strings.Repeat("│  ", level),
			char,
			i,
			expendedCharacter,
		)
		if value.isExpanded() {
			value.draw(writer, level+1)
		}
	}
	return nil

}
func (n listNode) filter(query query) bool {
	return true

}

type floatNode struct {
	baseTreeNode
	data json.Number
}

func (n *floatNode) collapseAll() {
}
func (n *floatNode) expandAll() {
}
func (n floatNode) isCollapsable() bool {
	return false
}
func (n floatNode) search(query string) (treeNode, error) {
	return nil, nil
}
func (n floatNode) find(tp treePosition) treeNode {
	return nil

}

func (n floatNode) String(_ int) string {
	return string(n.data)
}

func (n floatNode) draw(writer io.Writer, _ int) error {
	return nil

}
func (n floatNode) filter(query query) bool {
	return true
}

type stringNode struct {
	baseTreeNode
	data string
}

func (n *stringNode) collapseAll() {
}
func (n *stringNode) expandAll() {
}

func (n stringNode) isCollapsable() bool {
	return false
}

func (n stringNode) find(tp treePosition) treeNode {
	return nil
}

func (n stringNode) String(_ int) string {
	return encodeJson(n.data, 0)
}

func (n stringNode) search(query string) (treeNode, error) {
	return nil, nil

}
func (n stringNode) draw(writer io.Writer, _ int) error {
	//fmt.Fprintf(writer, "%s%q\n", strings.Repeat(" ", padding+padding*lvl), n.data)
	return nil

}
func (n stringNode) filter(query query) bool {
	return true
}

type boolNode struct {
	baseTreeNode
	data bool
}

func (n *boolNode) collapseAll() {
}
func (n *boolNode) expandAll() {
}

func (n boolNode) isCollapsable() bool {
	return false
}

func (n boolNode) find(tp treePosition) treeNode {
	return nil
}

func (n boolNode) String(_ int) string {
	return fmt.Sprintf("%t", n.data)
}

func (n boolNode) search(query string) (treeNode, error) {
	return nil, nil

}
func (n boolNode) draw(writer io.Writer, _ int) error {
	return nil

}
func (n boolNode) filter(query query) bool {
	return true
}

type nilNode struct {
	baseTreeNode
}

func (n *nilNode) collapseAll() {
}
func (n *nilNode) expandAll() {
}

func (n nilNode) isCollapsable() bool {
	return false
}

func (n nilNode) find(tp treePosition) treeNode {
	return nil
}

func (n nilNode) String(_ int) string {
	return "null"
}

func (n nilNode) search(query string) (treeNode, error) {
	return nil, nil

}
func (n nilNode) draw(writer io.Writer, _ int) error {
	return nil
}

func (n nilNode) filter(query query) bool {
	return true
}

func newTree(y interface{}) (treeNode, error) {
	var err error
	var tree treeNode
	switch v := y.(type) {
	case bool:
		tree = &boolNode{
			baseTreeNode{true},
			v,
		}
	case string:
		tree = &stringNode{
			baseTreeNode{true},
			v,
		}

	case nil:
		tree = &nilNode{baseTreeNode{true}}

	case json.Number:
		tree = &floatNode{
			baseTreeNode{true},
			v,
		}
	case *orderedmap.OrderedMap, orderedmap.OrderedMap, map[string]interface{}:
		tree, err = newComplexNode(v)
		if err != nil {
			return nil, err
		}
	case []interface{}:
		data := make([]treeNode, 0, len(v))
		for _, listItemInterface := range v {
			listItem, err := newTree(listItemInterface)
			if err != nil {
				return nil, err
			}
			data = append(data, listItem)
		}
		tree = &listNode{
			baseTreeNode: baseTreeNode{true},
			data:         data,
			raw:          v,
		}
	default:
		tree = &stringNode{baseTreeNode{true}, "TODO"}
	}
	return tree, nil
}

func newComplexNode(v interface{}) (treeNode, error) {
	orderMap, err := toOrderMap(v)
	if err != nil {
		return nil, err
	}
	data := orderedmap.NewWithSize(orderMap.Size())
	if err := orderMap.ForeachErr(func(key string, value interface{}) error {
		childNode, err := newTree(value)
		if err != nil {
			return err
		}
		data.Set(key, childNode)
		return nil
	}); err != nil {
		return nil, err
	}
	tree = &complexNode{
		baseTreeNode: baseTreeNode{true},
		data:         data,
		raw:          orderMap,
	}
	return tree, nil
}

func fromReader(r io.Reader) (treeNode, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}
	return fromBytes(b)
}

func fromFile(filename string) (treeNode, error) {
	open, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer open.Close()
	return fromReader(open)
}
