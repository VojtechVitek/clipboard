package main

import (
	"fmt"
	"log"
	"time"

	"github.com/VojtechVitek/clipboard"
	"github.com/jroimartin/gocui"
	"github.com/pkg/errors"
)

var (
	history            = clipboard.NewHistory()
	gui                *gocui.Gui
	sideView, mainView *gocui.View

	lastClickedCx, lastClickedCy = -1, -1
	doubleClicked                bool
)

func copyToClipboard(value string) error {
	if history.Save(value) {
		gui.Update(func(g *gocui.Gui) error {
			sideView.Clear()
			history.WriteShortValues(sideView)

			if err := sideView.SetCursor(0, 0); err != nil {
				return errors.Wrap(err, "failed to set origin")
			}
			if err := sideView.SetOrigin(0, 0); err != nil {
				return errors.Wrap(err, "failed to set origin")
			}

			mainView.Clear()
			fmt.Fprintf(mainView, value)

			return nil
		})
	}

	return nil
}

func collectClipboard() error {
	for {
		value, err := clipboard.Get()
		if err != nil {
			return errors.Wrap(err, "failed to collect clipboard")
		}

		if err := copyToClipboard(value); err != nil {
			return errors.Wrap(err, "failed to copy to clipboard")
		}

		time.Sleep(1 * time.Second)
	}
}

func main() {
	go collectClipboard()

	var err error
	gui, err = gocui.NewGui(gocui.Output256)
	if err != nil {
		log.Fatal(err)
	}
	defer gui.Close()

	gui.Cursor = true
	gui.Mouse = true

	gui.SetManagerFunc(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		var err error

		sideView, err = g.SetView("side", -1, -1, 50, maxY-2)
		if err != nil {
			sideView.Highlight = true
			sideView.SelBgColor = gocui.ColorGreen
			sideView.SelFgColor = gocui.ColorBlack
			if _, err := g.SetCurrentView("side"); err != nil {
				return errors.Wrap(err, "failed to set current view")
			}
		}
		mainView, err = g.SetView("main", 50, -1, maxX, maxY-2)
		if err != nil {
			mainView.Editable = true
			mainView.Wrap = true
		}
		if helpView, err := g.SetView("help", -1, maxY-2, maxX, maxY); err != nil {
			fmt.Fprintln(helpView, "Copy: double-click/Enter | Edit: click/→ | Delete: d/Del/⌫ | Exit: Ctrl-C")
			helpView.Editable = false
			helpView.Wrap = false
		}
		return nil
	})

	if err := keybindings(gui); err != nil {
		log.Fatal(err)
	}

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Fatalf("%+v", err)
	}
}

func toSideView(g *gocui.Gui, v *gocui.View) error {
	lastClickedCx, lastClickedCy = -1, -1
	doubleClicked = false

	if _, err := g.SetCurrentView("side"); err != nil {
		return errors.Wrap(err, "failed to set current view")
	}

	cx, cy := v.Cursor()
	if cx > 0 {
		return errors.Wrap(v.SetCursor(cx-1, cy), "failed to set cursor")
	}

	if err := clipboard.Set(v.ViewBuffer()); err != nil {
		return errors.Wrap(err, "failed to save clipboard value from editor")
	}

	return nil
}

func toMainView(g *gocui.Gui, v *gocui.View) error {
	lastClickedCx, lastClickedCy = -1, -1
	doubleClicked = false

	if _, err := g.SetCurrentView("main"); err != nil {
		return errors.Wrap(err, "failed to set current view")
	}

	cx, cy := v.Cursor()
	if cx > 0 || cy > 0 {
		return v.SetCursor(cx-1, cy)
	}

	return nil
}

func sideViewArrowDown(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	if cy+1 >= history.Len() {
		return nil
	}
	g.Update(func(g *gocui.Gui) error {
		mainView.Clear()
		fmt.Fprintln(mainView, history.Value(cy+1))
		return nil
	})
	if err := v.SetCursor(cx, cy+1); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy+1); err != nil {
			return errors.Wrap(err, "failed to set origin")
		}
	}
	return nil
}

func sideViewArrowUp(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	if cy == 0 {
		return nil
	}
	g.Update(func(g *gocui.Gui) error {
		mainView.Clear()
		fmt.Fprintln(mainView, history.Value(cy-1))
		return nil
	})
	if err := v.SetCursor(cx, cy-1); err != nil {
		ox, oy := v.Origin()
		if oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return errors.Wrap(err, "failed to set origin")
			}
		}
	}
	return nil
}

func sideViewClick(g *gocui.Gui, v *gocui.View) error {
	if _, err := g.SetCurrentView(v.Name()); err != nil {
		return errors.Wrap(err, "failed to set current view")
	}

	cx, cy := v.Cursor()
	if cy >= history.Len() {
		cy = history.Len() - 1
		if err := v.SetCursor(cx, cy); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy); err != nil {
				return errors.Wrap(err, "failed to set origin")
			}
		}
	}

	if cx == lastClickedCx && cy == lastClickedCy {
		lastClickedCx, lastClickedCy = -1, -1
		doubleClicked = true
		return copySelectedValueToClipboard(g, v)
	}

	lastClickedCx = cx
	lastClickedCy = cy

	g.Update(func(g *gocui.Gui) error {
		mainView.Clear()
		fmt.Fprintln(mainView, history.Value(cy))
		return nil
	})

	return nil
}

func sideViewReleaseClick(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()
	if cy >= history.Len() {
		cy = history.Len() - 1
		if err := v.SetCursor(cx, cy); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy); err != nil {
				return errors.Wrap(err, "failed to set origin")
			}
		}
	}

	if !doubleClicked {
		return nil
	}
	defer func() {
		lastClickedCx, lastClickedCy = -1, -1
		doubleClicked = false
	}()

	// Set focus on the first line.
	if err := v.SetCursor(0, 0); err != nil {
		ox, oy := v.Origin()
		if err := v.SetOrigin(ox, oy); err != nil {
			return errors.Wrap(err, "failed to set origin")
		}
	}

	return nil
}

func copySelectedValueToClipboard(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()

	value := history.Value(cy)
	clipboard.Set(value)
	if err := copyToClipboard(value); err != nil {
		return errors.Wrap(err, "failed to copy to clipboard")
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	bindings := map[string][]struct {
		Key interface{}

		Handler func(*gocui.Gui, *gocui.View) error
	}{
		"side": {
			{gocui.KeyTab, toMainView},
			{gocui.KeyArrowRight, toMainView},
			{gocui.KeyArrowDown, sideViewArrowDown},
			{gocui.KeyArrowUp, sideViewArrowUp},
			{gocui.KeyEnter, copySelectedValueToClipboard},
			{gocui.MouseLeft, sideViewClick},
			{gocui.MouseRelease, sideViewReleaseClick},
		},
		"main": {
			{gocui.KeyArrowLeft, toSideView},
			{gocui.MouseLeft, toMainView},
			{gocui.KeyCtrlS, toSideView},
		},
		"": {
			{gocui.KeyCtrlC, quit},
		},
	}

	for view, bindings := range bindings {
		for _, binding := range bindings {
			if err := g.SetKeybinding(view, binding.Key, gocui.ModNone, binding.Handler); err != nil {
				return errors.Wrap(err, "failed to set keybinding")
			}
		}
	}

	return nil
}
