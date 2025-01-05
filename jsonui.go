package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"strconv"
	"strings"

	"github.com/atotto/clipboard"

	"github.com/jroimartin/gocui"
)

const VERSION = "1.0.1"

const (
	treeView = "tree"
	textView = "text"
	pathView = "path"
	helpView = "help"
)

const jsonPadding = 2

type position struct {
	prc    float32
	margin int
}

func (p position) getCoordinate(max int) int {
	// value = prc * MAX + abs
	return int(p.prc*float32(max)) - p.margin
}

type viewPosition struct {
	x0, y0, x1, y1 position
}

func logFile(s string) error {
	d1 := []byte(s + "\n")
	return ioutil.WriteFile("log.txt", d1, 0644)
}

func (vp viewPosition) getCoordinates(maxX, maxY int) (int, int, int, int) {
	var x0 = vp.x0.getCoordinate(maxX)
	var y0 = vp.y0.getCoordinate(maxY)
	var x1 = vp.x1.getCoordinate(maxX)
	var y1 = vp.y1.getCoordinate(maxY)
	return x0, y0, x1, y1
}

var helpWindowToggle = false
var currentViewName = ""
var expandAllStatus = true
var formatData = false

var viewPositions = map[string]viewPosition{
	treeView: {
		position{0.0, 0},
		position{0.0, 0},
		position{0.3, 2},
		position{0.9, 2},
	},
	textView: {
		position{0.3, 0},
		position{0.0, 0},
		position{1.0, 2},
		position{0.9, 2},
	},
	pathView: {
		position{0.0, 0},
		position{0.89, 0},
		position{1.0, 2},
		position{1.0, 2},
	},
}

var tree treeNode

func bindingDirectionKey(g *gocui.Gui, viewName string) error {
	if err := g.SetKeybinding(viewName, gocui.KeyArrowLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		vx, vy := v.Cursor()
		ox, oy := v.Origin()
		if vx+ox == 0 {
			return nil
		}
		if err := v.SetCursor(vx-1, vy); err != nil {
			return v.SetOrigin(ox-1, oy)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding(viewName, gocui.KeyArrowRight, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		vx, vy := v.Cursor()
		ox, oy := v.Origin()
		if err := v.SetCursor(vx+1, vy); err != nil {
			return v.SetOrigin(ox+1, oy)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding(viewName, gocui.KeyArrowUp, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		vx, vy := v.Cursor()
		ox, oy := v.Origin()
		if vy+oy == 0 {
			return nil
		}
		if err := v.SetCursor(vx, vy-1); err != nil {
			return v.SetOrigin(ox, oy-1)
		}
		return nil
	}); err != nil {
		return err
	}

	if err := g.SetKeybinding(viewName, gocui.KeyArrowDown, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		vx, vy := v.Cursor()
		ox, oy := v.Origin()
		if err := v.SetCursor(vx, vy+1); err != nil {
			return v.SetOrigin(ox, oy+1)
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func initGUI(g *gocui.Gui) {
	g.Cursor = true
	g.SetManagerFunc(layout)

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'q', gocui.ModNone, quit); err != nil {
		log.Panicln(err)
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if view.Name() == treeView {
			currentViewName = textView
			drawJSON(gui)
			return nil
		}
		if view.Name() == textView {
			currentViewName = treeView
			drawJSON(gui)
		}
		return nil
	}); err != nil {
		log.Panicln(err)
	}
	if err := bindingDirectionKey(g, textView); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyCtrlY, gocui.ModNone, cursorMovement(-1)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyCtrlE, gocui.ModNone, cursorMovement(1)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyArrowUp, gocui.ModNone, cursorMovement(-1)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyArrowDown, gocui.ModNone, cursorMovement(1)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyCtrlU, gocui.ModNone, cursorMovement(-15)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyCtrlD, gocui.ModNone, cursorMovement(15)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyCtrlB, gocui.ModNone, cursorMovement(-15)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyCtrlF, gocui.ModNone, cursorMovement(15)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyPgup, gocui.ModNone, cursorMovement(-15)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyPgdn, gocui.ModNone, cursorMovement(15)); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyArrowRight, gocui.ModNone, toggleExpand); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyArrowLeft, gocui.ModNone, toggleExpand); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, 'c', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		p := findTreePosition(g)
		subTree := tree.find(p)
		data := subTree.String(2)
		if formatData {
			data = FormatData(data)
		}
		_ = clipboard.WriteAll(data)
		return nil
	}); err != nil {
		log.Panicln(err)
	}
	g.SetKeybinding(treeView, 'f', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		formatData = !formatData
		drawJSON(g)
		drawPath(g)
		return nil
	})
	if err := g.SetKeybinding(treeView, 'e', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
		if expandAllStatus {
			expandAllStatus = false
			return collapseAll(gui)
		}
		expandAllStatus = true
		return expandAll(gui)
	}); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'h', gocui.ModNone, toggleHelp); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", '?', gocui.ModNone, toggleHelp); err != nil {
		log.Panicln(err)
	}
	if _, err := g.SetCurrentView(treeView); err != nil {
		log.Println(err)
	}
	g.SelFgColor = gocui.ColorBlack
	g.SelBgColor = gocui.ColorGreen
}

var helpMessage = ""

func init() {
	helpMessage = initHelpMsg().String()
}

func layout(g *gocui.Gui) error {
	var views = []string{treeView, textView, pathView}
	maxX, maxY := g.Size()
	for _, view := range views {
		x0, y0, x1, y1 := viewPositions[view].getCoordinates(maxX, maxY)
		if v, err := g.SetView(view, x0, y0, x1, y1); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			v.SelFgColor = gocui.ColorBlack
			v.SelBgColor = gocui.ColorGreen

			v.Title = " " + view + " "
			if v.Name() == treeView {
				v.Highlight = true
				drawTree(g, tree)
				// v.Autoscroll = true
			}
			if v.Name() == textView {
				drawJSON(g)
			}
		}
	}
	if helpWindowToggle {
		height := strings.Count(helpMessage, "\n") + 1
		width := -1
		for _, line := range strings.Split(helpMessage, "\n") {
			width = int(math.Max(float64(width), float64(len(line)+2)))
		}
		if v, err := g.SetView(helpView, maxX/2-width/2, maxY/2-height/2, maxX/2+width/2, maxY/2+height/2); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			Println(v, helpMessage)
		}
	} else {
		g.DeleteView(helpView)
	}
	if currentViewName == "" {
		currentViewName = treeView
	}
	_, err := g.SetCurrentView(currentViewName)
	if err != nil {
		log.Fatal("failed to set current view: ", err)
	}
	return nil

}
func getPath(g *gocui.Gui) string {
	p := findTreePosition(g)
	for i, s := range p {
		transformed := s
		if !strings.HasPrefix(s, "[") && !strings.HasSuffix(s, "]") {
			transformed = fmt.Sprintf("[%q]", s)
		}
		p[i] = transformed
	}
	return strings.Join(p, "")
}

func drawPath(g *gocui.Gui) error {
	pv, err := g.View(pathView)
	if err != nil {
		log.Fatal("failed to get pathView", err)
	}
	p := getPath(g)
	if formatData {
		p = p + " (format)"
	}
	pv.Clear()
	pv.SetOrigin(0, 0)
	pv.SetCursor(0, 0)
	Printf(pv, p)
	return nil
}

func drawJSON(g *gocui.Gui) error {
	dv, err := g.View(textView)
	if err != nil {
		log.Fatal("failed to get textView", err)
	}
	if currentViewName == treeView {
		return drawSimpleJSON(g, dv)
	}
	treeTodraw := tree.find(findTreePosition(g))
	if treeTodraw == nil {
		return nil
	}
	data := treeTodraw.String(jsonPadding)
	dv.Clear()
	dv.SetOrigin(0, 0)
	dv.SetCursor(0, 0)
	if formatData {
		data = FormatData(data)
	}
	Printf(dv, data)
	return nil
}

var jsonCache = newCacheData(1024 * 10)

func drawSimpleJSON(g *gocui.Gui, dv *gocui.View) error {
	_, y := dv.Size()
	p := findTreePosition(g)
	key := strings.Join(p, "|") + "#" + strconv.Itoa(y) + "#" + strconv.FormatBool(formatData)
	value, _ := jsonCache.getOrStore(key, func() interface{} {
		treeTodraw := tree.find(p)
		if treeTodraw == nil {
			return ""
		}
		data := treeTodraw.String(jsonPadding)
		if formatData {
			data = FormatData(data)
		}
		return limitLineData(data, y)
	})
	data := value.(string)
	if value != "" {
		dv.Clear()
		dv.SetOrigin(0, 0)
		dv.SetCursor(0, 0)
		Printf(dv, data)
	}
	return nil
}

func limitLineData(data string, size int) string {
	lines := strings.SplitN(data, "\n", size+1)
	if len(lines) == size+1 {
		return strings.Join(lines[:size], "\n")
	}
	return data
}

func lineBelow(v *gocui.View, d int) bool {
	_, y := v.Cursor()
	line, err := v.Line(y + d)
	return err == nil && line != ""
}

func countIndex(s string) int {
	count := 0
	for _, c := range s {
		if c == ' ' {
			count++
		}
	}
	return count
}

func getLine(s string, y int) string {
	lines := strings.Split(s, "\n")
	return lines[y]
}

var cleanPatterns = []string{
	treeSignUpEnding,
	treeSignDash,
	treeSignUpMiddle,
	treeSignVertical,
	" (+)",
}

func findTreePosition(g *gocui.Gui) treePosition {
	v, err := g.View(treeView)
	if err != nil {
		log.Fatal("failed to get treeview", err)
	}
	path := treePosition{}
	ci := -1
	_, yOffset := v.Origin()
	_, yCurrent := v.Cursor()
	y := yOffset + yCurrent
	s := v.Buffer()
	for cy := y; cy >= 0; cy-- {
		line := getLine(s, cy)
		for _, pattern := range cleanPatterns {
			line = strings.Replace(line, pattern, "", -1)
		}

		if count := countIndex(line); count < ci || ci == -1 {
			path = append(path, strings.TrimSpace(line))
			ci = count
		}
	}
	for i := len(path)/2 - 1; i >= 0; i-- {
		opp := len(path) - 1 - i
		path[i], path[opp] = path[opp], path[i]
	}

	return path[1:]
}

// This is a workaround for not having a Buffer
// function in gocui
func bufferLen(v *gocui.View) int {
	s := v.Buffer()
	return len(strings.Split(s, "\n")) - 1
}

func drawTree(g *gocui.Gui, tree treeNode) error {
	tv, err := g.View(treeView)
	if err != nil {
		log.Fatal("failed to get treeView", err)
	}
	tv.Clear()
	tree.draw(tv, 0)
	maxY := bufferLen(tv)
	cx, cy := tv.Cursor()
	lastLine := maxY - 2
	if cy > lastLine {
		tv.SetCursor(cx, lastLine)
		tv.SetOrigin(0, 0)
	}
	return nil
}

func expandAll(g *gocui.Gui) error {
	tree.expandAll()
	return drawTree(g, tree)
}

func collapseAll(g *gocui.Gui) error {
	tree.collapseAll()
	return drawTree(g, tree)
}

func toggleExpand(g *gocui.Gui, v *gocui.View) error {
	p := findTreePosition(g)
	subTree := tree.find(p)
	subTree.toggleExpanded()
	return drawTree(g, tree)
}

func cursorMovement(d int) func(g *gocui.Gui, v *gocui.View) error {
	return func(g *gocui.Gui, v *gocui.View) error {
		dir := 1
		if d < 0 {
			dir = -1
		}
		distance := int(math.Abs(float64(d)))
		for ; distance > 0; distance-- {
			if lineBelow(v, distance*dir) {
				v.MoveCursor(0, distance*dir, false)
				drawJSON(g)
				drawPath(g)
				return nil
			}
		}
		return nil
	}
}
func toggleHelp(g *gocui.Gui, v *gocui.View) error {
	helpWindowToggle = !helpWindowToggle
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}
