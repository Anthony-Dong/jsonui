package main

import (
	"bytes"
	"context"
	"fmt"
	"github.com/anthony-dong/jsonui/internal"
	"github.com/atotto/clipboard"
	"golang.org/x/sync/errgroup"
	"io/ioutil"
	"log"
	"math"
	"strings"
	"time"

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
		position{0.3, 1},
		position{0.9, 1},
	},
	textView: {
		position{0.3, 0},
		position{0.0, 0},
		position{1.0, 1},
		position{0.9, 1},
	},
	pathView: {
		position{0.0, 0},
		position{0.89, 0},
		position{1.0, 1},
		position{1.0, 1},
	},
}

var tree treeNode
var treeController internal.ViewBufferController
var textController internal.ViewBufferController
var rootTextController internal.ViewBufferController

var helpMessage = ""

func initController() error {
	helpMessage = initHelpMsg().String()
	wg := errgroup.Group{}
	wg.Go(func() error {
		buf := bytes.NewBuffer(make([]byte, 0, 1024*1024))
		if err := tree.draw(buf, 0); err != nil {
			return err
		}
		treeController.Write(buf.Bytes())
		return nil
	})
	wg.Go(func() error {
		data := tree.String(jsonPadding)
		textController.WriteString(data)
		rootTextController.WriteString(data)
		return nil
	})
	return wg.Wait()
}

func initGUI(g *gocui.Gui) {
	if err := initController(); err != nil {
		return
	}
	g.Cursor = true
	g.SetManagerFunc(layout)

	internal.MultiSetKeybinding(g, "", []interface{}{gocui.KeyCtrlC, 'q'}, internal.Quit)

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, switchView); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyEnter, gocui.ModNone, switchView); err != nil {
		log.Panicln(err)
	}
	if err := internal.BindingDirectionKey(g, textView, &textController); err != nil {
		log.Panicln(err)
	}

	internal.MultiSetKeybinding(g, treeView, []interface{}{gocui.KeyCtrlY, gocui.KeyArrowUp}, cursorMovement(-1))
	internal.MultiSetKeybinding(g, treeView, []interface{}{gocui.KeyCtrlE, gocui.KeyArrowDown}, cursorMovement(1))
	internal.MultiSetKeybinding(g, treeView, []interface{}{gocui.KeyCtrlU, gocui.KeyPgup, gocui.KeyCtrlB}, cursorMovement(-15))
	internal.MultiSetKeybinding(g, treeView, []interface{}{gocui.KeyCtrlD, gocui.KeyPgdn, gocui.KeyCtrlF}, cursorMovement(15))

	if err := g.SetKeybinding(treeView, gocui.KeyArrowRight, gocui.ModNone, toggleExpand); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding(treeView, gocui.KeyArrowLeft, gocui.ModNone, toggleExpand); err != nil {
		log.Panicln(err)
	}
	if err := g.SetKeybinding("", 'c', gocui.ModNone, func(gui *gocui.Gui, view *gocui.View) error {
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
	if err := g.SetKeybinding("", 'f', gocui.ModNone, formatView); err != nil {
		log.Panicln(err)
	}
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
	g.SelFgColor = gocui.ColorBlack
	g.SelBgColor = gocui.ColorGreen
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
				if err := treeController.Draw(v); err != nil {
					return err
				}
			}
			if v.Name() == textView {
				if err := textController.Draw(v); err != nil {
					return err
				}
				v.Title = fmt.Sprintf(` %s [lines=%d]`, views, len(textController.Lines))
			}
			if v.Name() == pathView {
				v.Wrap = true
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
		p = p + " (EnableFormat)"
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
		return err
	}
	path := findTreePosition(g)
	if len(path) == 0 {
		return rootTextController.Draw(dv)
	}
	treeToDraw := tree.find(path)
	if treeToDraw == nil {
		return nil
	}
	var data = ""
	if err := internal.RunWithTimeout(context.Background(), time.Second, func() error {
		data = treeToDraw.String(jsonPadding)
		return nil
	}); err != nil {
		return textController.ReDraw(dv, []byte("Error: 超时"))
	}
	if formatData {
		data = FormatData(data)
	}
	if err := textController.ReDraw(dv, internal.String2Bytes(data)); err != nil {
		return err
	}
	dv.Title = fmt.Sprintf(" text [lines=%d] ", len(textController.Lines))
	return nil
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
	_, yCurrent := v.Cursor()
	y := treeController.Origin + yCurrent
	for cy := y; cy >= 0; cy-- {
		line := string(treeController.Lines[cy])
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

func drawTree(g *gocui.Gui, tree treeNode) error {
	tv, err := g.View(treeView)
	if err != nil {
		log.Fatal("failed to get treeView", err)
	}
	buf := bytes.NewBuffer(make([]byte, 0, 1024*1024))
	if err := tree.draw(buf, 0); err != nil {
		return err
	}
	return treeController.ReDraw(tv, buf.Bytes())
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
		_ = treeController.MoveCursor(v, 0, d)
		_ = drawJSON(g)
		_ = drawPath(g)
		return nil
	}
}
func toggleHelp(g *gocui.Gui, v *gocui.View) error {
	helpWindowToggle = !helpWindowToggle
	return nil
}

func switchView(gui *gocui.Gui, view *gocui.View) error {
	if view.Name() == treeView {
		currentViewName = textView
		return nil
	}
	if view.Name() == textView {
		currentViewName = treeView
	}
	return nil
}

func formatView(gui *gocui.Gui, view *gocui.View) error {
	formatData = !formatData
	drawJSON(gui)
	drawPath(gui)
	return nil
}
