package internal

import (
	"bytes"
	"github.com/jroimartin/gocui"
)

type ViewBufferController struct {
	Lines  [][]byte
	Origin int
	Size   int
}

func (c *ViewBufferController) Clear() *ViewBufferController {
	c.Lines = c.Lines[:0]
	c.Origin = 0
	return c
}

func (c *ViewBufferController) ClearCursor(v *gocui.View) error {
	if err := v.SetCursor(0, 0); err != nil {
		return err
	}
	if err := v.SetOrigin(0, 0); err != nil {
		return err
	}
	return nil
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

func (c *ViewBufferController) addCursor(v *gocui.View, inc int) error {
	cx, cy := v.Cursor()
	_, sy := v.Size()
	ncy, noy := incCursor(cy, c.Origin, sy, inc, len(c.Lines))
	if c.Origin != noy {
		c.Origin = noy
		_ = v.SetCursor(cx, ncy)
		if err := c.Draw(v); err != nil {
			return err
		}
		return nil
	}
	_ = v.SetCursor(cx, ncy)
	return nil
}

func (c *ViewBufferController) getCurrentLine(cy int) []byte {
	return c.Lines[cy+c.Origin]
}

func (c *ViewBufferController) MoveCursor(v *gocui.View, ix int, iy int) error {
	if ix != 0 {
		cx, cy := v.Cursor()
		_ = v.SetCursor(moveCursor(cx, ix), cy)
	}
	if iy != 0 {
		return c.addCursor(v, iy)
	}
	return nil
}

func moveCursor(cx, inc int) (_ int) {
	if cx+inc <= 0 {
		return 0
	}
	return cx + inc
}

func incCursor(cy int, oy, sy, inc, max int) (int, int) {
	// 索引大于 max
	if oy+cy+inc >= max {
		add := max - (oy + cy + 1)
		if cy == sy-1 {
			return cy, oy + add
		}
		return cy + add, oy
	}
	// 1. 如果新增后 cy >= sy，那么就cy不变位置
	if cy+inc >= sy {
		return sy - 1, oy + (inc - (sy - cy - 1))
	}
	if cy+inc <= 0 {
		oy = oy + (inc - cy)
		if oy <= 0 {
			oy = 0
		}
		return 0, oy
	}
	return cy + inc, oy
}

func (c *ViewBufferController) GetCurView(viewSize int) [][]byte {
	data := c.Lines[c.Origin:]
	if len(data) <= viewSize {
		return data
	}
	return data[:viewSize]
}

func (c *ViewBufferController) getCurViewData(v *gocui.View) []byte {
	_, y := v.Size()
	return bytes.Join(c.GetCurView(y), []byte{})
}

func (c *ViewBufferController) Draw(v *gocui.View) error {
	v.Clear()
	data := c.getCurViewData(v)
	if _, err := v.Write(TerminalBytes(data)); err != nil {
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
