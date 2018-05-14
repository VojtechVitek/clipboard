package main

import (
	"fmt"
	"log"
	"time"

	"github.com/VojtechVitek/clipboard"
	"github.com/jroimartin/gocui"
)

var (
	history            = clipboard.NewHistory()
	gui                *gocui.Gui
	sideView, mainView *gocui.View
)

func historySave(value string) {
	if history.Save(value) {
		gui.Update(func(g *gocui.Gui) error {
			sideView.Clear()
			history.WriteShortValues(sideView)

			_ = sideView.SetCursor(0, 0)
			_ = sideView.SetOrigin(0, 0)

			mainView.Clear()
			fmt.Fprintf(mainView, value)

			return nil
		})
	}
}

func collectClipboard() {
	for {
		value, err := clipboard.Get()
		if err != nil {
			log.Println(err)
			time.Sleep(1 * time.Second)
			continue
		}

		historySave(value)

		time.Sleep(1 * time.Second)
	}
}

func toSideView(g *gocui.Gui, v *gocui.View) error {
	cx, cy := v.Cursor()

	if cx > 0 || cy > 0 {
		v.SetCursor(cx-1, cy)
		return nil
	}

	_, err := g.SetCurrentView("side")
	return err
}

func toMainView(g *gocui.Gui, v *gocui.View) error {
	_, err := g.SetCurrentView("main")
	return err
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
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
			return err
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
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
				return err
			}
		}
	}
	return nil
}

func getLine(g *gocui.Gui, v *gocui.View) error {
	_, cy := v.Cursor()

	value := history.Value(cy)
	clipboard.Set(value)
	historySave(value)

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

func keybindings(g *gocui.Gui) error {
	if err := g.SetKeybinding("side", gocui.KeyTab, gocui.ModNone, toMainView); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowRight, gocui.ModNone, toMainView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyTab, gocui.ModNone, toSideView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main", gocui.KeyArrowLeft, gocui.ModNone, toSideView); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("side", gocui.KeyEnter, gocui.ModNone, getLine); err != nil {
		return err
	}

	if err := g.SetKeybinding("main", gocui.KeyCtrlS, gocui.ModNone, saveVisualMain); err != nil {
		return err
	}
	return nil
}

func saveVisualMain(g *gocui.Gui, v *gocui.View) error {
	return clipboard.Set(v.ViewBuffer())
}

func main() {
	go collectClipboard()

	var err error
	gui, err = gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer gui.Close()

	gui.Cursor = true

	gui.SetManagerFunc(func(g *gocui.Gui) error {
		maxX, maxY := g.Size()
		var err error

		sideView, err = g.SetView("side", -1, -1, 50, maxY-2)
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			sideView.Highlight = true
			sideView.SelBgColor = gocui.ColorGreen
			sideView.SelFgColor = gocui.ColorBlack
			if _, err := g.SetCurrentView("side"); err != nil {
				return err
			}
		}
		mainView, err = g.SetView("main", 50, -1, maxX, maxY-2)
		if err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			mainView.Editable = true
			mainView.Wrap = true
		}
		if helpView, err := g.SetView("help", -1, maxY-2, maxX, maxY); err != nil {
			if err != gocui.ErrUnknownView {
				return err
			}
			fmt.Fprintln(helpView, "Copy: c/Enter | Edit: e/→ | Delete: d/Del/⌫ | Up: ↑ | Down: ↓ | Exit: q/Ctrl-C")
			helpView.Editable = false
			helpView.Wrap = false
		}
		return nil
	})

	if err := keybindings(gui); err != nil {
		log.Panicln(err)
	}

	if err := gui.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}
