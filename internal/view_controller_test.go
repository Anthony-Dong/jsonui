package internal

import (
	"strconv"
	"testing"
)

func Test_incCursor(t *testing.T) {
	t.Run("test1", func(t *testing.T) {
		sy := 30
		oy := 0
		cy := 0
		Max := 5
		for x := 0; x < 10; x++ {
			cy, oy = incCursor(cy, oy, sy, 1, Max)
			t.Log(cy, oy)
		}
		t.Log("=============")
		for x := 0; x < 10; x++ {
			cy, oy = incCursor(cy, oy, sy, -1, Max)
			t.Log(cy, oy)
		}
	})

	t.Run("test2", func(t *testing.T) {
		sy := 10
		oy := 0
		cy := 0
		Max := 15
		for x := 0; x < 20; x++ {
			cy, oy = incCursor(cy, oy, sy, 1, Max)
			t.Log(cy, oy)
		}
	})

	t.Run("test3", func(t *testing.T) {
		sy := 10
		oy := 0
		cy := 0
		Max := 15
		for x := 0; x < 20; x++ {
			cy, oy = incCursor(cy, oy, sy, 10, Max)
			t.Log(cy, oy)
		}
		if cy != 9 && oy != 5 {
			t.Fatal("error")
		}
		t.Log("==")
		for x := 0; x < 20; x++ {
			cy, oy = incCursor(cy, oy, sy, -10, Max)
			t.Log(cy, oy)
		}
		if cy != 0 && oy != 0 {
			t.Fatal("error")
		}
	})

}

func TestViewBufferController_GetCurView(t *testing.T) {
	c := ViewBufferController{}
	for x := 0; x < 2; x++ {
		c.WriteString(strconv.Itoa(x))
	}
	t.Log(c.GetCurView(10))
}
