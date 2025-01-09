package internal

import (
	"bytes"
	"fmt"
	"github.com/jroimartin/gocui"
)

type Location struct {
	TotalY int
	CurY   int
	TotalX int
	CurX   int
}

func (l *Location) CurrentIndex() string {
	return fmt.Sprintf("%d:%d", l.CurY, l.CurX)
}

type ViewBufferController struct {
	Lines  [][]byte
	Origin int
	Size   int
}

func (c *ViewBufferController) Location(v *gocui.View) *Location {
	cx, cy := v.Cursor()
	line := c.getCursorLine(cy)
	return &Location{
		TotalY: len(c.Lines),
		CurY:   cy + c.Origin + 1,
		TotalX: len(line),
		CurX:   cx + 1,
	}
}

func (c *ViewBufferController) Clear() *ViewBufferController {
	c.Lines = c.Lines[:0]
	c.Origin = 0
	return c
}

func (c *ViewBufferController) WriteString(s string) *ViewBufferController {
	c.Write(String2Bytes(s))
	return c
}

func (c *ViewBufferController) Write(data []byte) *ViewBufferController {
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}
	split := bytes.Split(data, []byte{'\n'})
	for _, elem := range split {
		c.Lines = append(c.Lines, append(elem, '\n'))
	}
	c.Size = len(data)
	return c
}

func (c *ViewBufferController) MoveCursor(v *gocui.View, x, y int) error {
	cx, cy := v.Cursor()
	_, sy := v.Size()

	ncy, noy := cursorY(cy, c.Origin, sy, y, len(c.Lines))
	ncx := cursorX(cx, len(c.getCursorLine(cy)), x)
	if c.Origin != noy {
		c.Origin = noy
		if err := c.Draw(v); err != nil {
			return err
		}
	}
	_ = v.SetCursor(ncx, ncy)
	return nil
}

func (c *ViewBufferController) getCursorLine(cy int) []byte {
	index := cy + c.Origin
	if index >= len(c.Lines) {
		return nil
	}
	return c.Lines[index]
}

func addCursorX(cx, max int) int {
	if cx+1 > max {
		return cx
	}
	return cx + 1
}

func subCursorX(cx int) int {
	if cx-1 < 0 {
		return 0
	}
	return cx - 1
}

func addCursorY(cy, oy, sy, max int) (int, int) {
	if cy+oy+1 >= max {
		return cy, oy
	}
	if cy+1 >= sy {
		return cy, oy + 1
	}
	return cy + 1, oy
}

func subCursorY(cy, oy int) (int, int) {
	if cy == 0 {
		if oy-1 <= 0 {
			return 0, 0
		}
		return 0, oy - 1
	}
	return cy - 1, oy
}

func abs(a int) (int, bool) {
	if a < 0 {
		return -a, false
	}
	return a, true
}

func cursorX(cx, max int, inc int) int {
	inc, isInc := abs(inc)
	if isInc {
		for x := 0; x < inc; x++ {
			cx = addCursorX(cx, max)
		}
		return cx
	}
	for x := 0; x < inc; x++ {
		cx = subCursorX(cx)
	}
	return cx
}

func cursorY(cy int, oy, sy, inc, max int) (int, int) {
	inc, isInc := abs(inc)
	if isInc {
		for x := 0; x < inc; x++ {
			cy, oy = addCursorY(cy, oy, sy, max)
		}
		return cy, oy
	}
	for x := 0; x < inc; x++ {
		cy, oy = subCursorY(cy, oy)
	}
	return cy, oy
}

func (c *ViewBufferController) getCurView(viewSize int) [][]byte {
	data := c.Lines[c.Origin:]
	if len(data) <= viewSize {
		return data
	}
	return data[:viewSize]
}

func (c *ViewBufferController) getCurViewData(v *gocui.View) []byte {
	_, y := v.Size()
	return bytes.Join(c.getCurView(y), []byte{})
}

func (c *ViewBufferController) Draw(v *gocui.View) error {
	v.Clear()
	if _, err := v.Write(TerminalBytes(c.getCurViewData(v))); err != nil {
		return err
	}
	return nil
}

func (c *ViewBufferController) ReDraw(v *gocui.View, data []byte) error {
	c.Clear()
	c.Write(data)
	if err := c.Draw(v); err != nil {
		return err
	}
	if err := v.SetOrigin(0, 0); err != nil {
		return err
	}
	if err := v.SetCursor(0, 0); err != nil {
		return err
	}
	return nil
}

func BindingDirectionKey(g *gocui.Gui, viewName string, controller *ViewBufferController) error {
	if err := g.SetKeybinding(viewName, gocui.KeyArrowLeft, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return controller.MoveCursor(v, -1, 0)
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyArrowRight, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		return controller.MoveCursor(v, 1, 0)
	}); err != nil {
		return err
	}

	down := func(add int) func(g *gocui.Gui, v *gocui.View) error {
		return func(g *gocui.Gui, v *gocui.View) error {
			return controller.MoveCursor(v, 0, add)
		}
	}
	up := func(add int) func(g *gocui.Gui, v *gocui.View) error {
		return func(g *gocui.Gui, v *gocui.View) error {
			return controller.MoveCursor(v, 0, -add)
		}
	}

	if err := g.SetKeybinding(viewName, gocui.KeyArrowUp, gocui.ModNone, up(1)); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyArrowDown, gocui.ModNone, down(1)); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyCtrlY, gocui.ModNone, up(1)); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyCtrlE, gocui.ModNone, down(1)); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyCtrlF, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, sy := v.Size()
		return down(sy)(g, v)
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyCtrlB, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, sy := v.Size()
		return up(sy)(g, v)
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyCtrlD, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, sy := v.Size()
		return down(sy/2)(g, v)
	}); err != nil {
		return err
	}
	if err := g.SetKeybinding(viewName, gocui.KeyCtrlU, gocui.ModNone, func(g *gocui.Gui, v *gocui.View) error {
		_, sy := v.Size()
		return up(sy/2)(g, v)
	}); err != nil {
		return err
	}
	return nil
}
