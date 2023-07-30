package main

// This is a small go-based nice demonstration of building rich console-based user interface.

// Version  : 1.0
// Author   : Jerome AMON
// Created  : 15 November 2021

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/jroimartin/gocui"
)

const (

	// Jobs box width.
	jobswidth = 22
)

// global datastore.
var dbs *databases

// struct of a datastore.
type databases struct {
	jobs    map[string]string
	actions map[string]string
	lock    *sync.RWMutex
}

// newDatabases creates new databases.
func newDatabases() *databases {
	return &databases{
		jobs:    map[string]string{},
		actions: map[string]string{},
		lock:    &sync.RWMutex{},
	}
}

// addJob generates random id for the job and save to jobs store.
func (db *databases) addJob(data string) {
	id := fmt.Sprintf("%x", time.Now().UnixNano())
	db.lock.Lock()
	db.jobs[id] = strings.TrimSpace(data)
	db.lock.Unlock()
}

// addAction generates random id for the action and save to actions store.
func (db *databases) addAction(data string) {
	id := fmt.Sprintf("%x", time.Now().UnixNano())
	db.lock.Lock()
	db.actions[id] = strings.TrimSpace(data)
	db.lock.Unlock()
}

// getJob retrieves a given job data based on its id from jobs store.
func (db *databases) getJob(id string) string {
	var data string
	db.lock.RLock()
	data = db.jobs[id]
	db.lock.RUnlock()
	return data
}

// getAction retrieves a given action data based on its id from actions store.
func (db *databases) getAction(id string) string {
	var data string
	db.lock.RLock()
	data = db.actions[id]
	db.lock.RUnlock()
	return data
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())
	rand.Seed(time.Now().UnixNano())

	f, err := os.OpenFile("logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("failed to create logs file.")
	}
	defer f.Close()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(f)

	dbs = newDatabases()

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	// active view to use its Sel properties.
	g.Highlight = true
	// view border color.
	g.SelFgColor = gocui.ColorRed
	g.BgColor = gocui.ColorBlack
	g.FgColor = gocui.ColorWhite
	// enable mouse / cursor / Esc as key.
	g.Cursor = true
	g.InputEsc = true
	g.Mouse = true

	g.SetManagerFunc(layout)

	err = g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit)
	if err != nil {
		log.Println("Could not set key binding:", err)
		return
	}

	maxX, maxY := g.Size()

	// Jobs list view.
	jobsView, err := g.SetView("jobs", 0, 0, jobswidth, maxY-12)
	// ErrUnknownView means the view did not exist before and needs initialization.
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create jobs list view:", err)
		return
	}
	jobsView.Title = "Jobs"
	jobsView.FgColor = gocui.ColorWhite
	jobsView.SelBgColor = gocui.ColorBlue
	jobsView.SelFgColor = gocui.ColorWhite
	jobsView.Highlight = true

	// Outputs view.
	outputsView, err := g.SetView("outputs", jobswidth+1, 0, maxX-1, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create outputs view:", err)
		return
	}
	outputsView.Title = "Outputs"
	outputsView.FgColor = gocui.ColorWhite
	outputsView.SelBgColor = gocui.ColorBlack
	outputsView.SelFgColor = gocui.ColorRed
	// scroll if the output exceeds the visible area.
	outputsView.Autoscroll = true
	outputsView.Wrap = true

	// Actions view.
	actionsView, err := g.SetView("actions", 0, maxY-11, jobswidth, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create actions view:", err)
		return
	}
	actionsView.Title = "Actions"
	actionsView.FgColor = gocui.ColorYellow
	actionsView.SelBgColor = gocui.ColorGreen
	actionsView.SelFgColor = gocui.ColorBlack
	actionsView.Highlight = true

	// Commands view.
	commandsView, err := g.SetView("commands", 0, maxY-3, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create commands view:", err)
		return
	}
	commandsView.Title = "Commands"
	commandsView.FgColor = gocui.ColorYellow
	commandsView.SelBgColor = gocui.ColorBlack
	commandsView.SelFgColor = gocui.ColorYellow
	commandsView.Editable = true
	commandsView.Autoscroll = true
	commandsView.Wrap = false

	// Apply keybindings to program.
	if err = keybindings(g); err != nil {
		log.Panicln(err)
	}

	// move the focus on the jobs list box.
	if _, err = g.SetCurrentView("jobs"); err != nil {
		log.Println("Failed to set focus on jobs view:", err)
		return
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		log.Panicln(err)
	}
}

func layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()

	// Jobs list view.
	_, err := g.SetView("jobs", 0, 0, jobswidth, maxY-12)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create jobs list view:", err)
		return err
	}

	// Outputs view.
	_, err = g.SetView("outputs", jobswidth+1, 0, maxX-1, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create outputs view:", err)
		return err
	}

	// Actions view.
	_, err = g.SetView("actions", 0, maxY-11, jobswidth, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create actions view:", err)
		return err
	}

	// Commands view.
	_, err = g.SetView("commands", 0, maxY-3, maxX-1, maxY-1)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create commands view:", err)
		return err
	}

	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrQuit
}

// keybindings binds multiple keys to views.
func keybindings(g *gocui.Gui) error {

	// keys binding on global terminal itself.
	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}

	// Ctrl+N key binding to create new data entry.
	if err := g.SetKeybinding("", gocui.KeyCtrlN, gocui.ModNone, inputView); err != nil {
		return err
	}

	if err := g.SetKeybinding("jobs", gocui.KeyCtrlN, gocui.ModNone, inputView); err != nil {
		return err
	}

	if err := g.SetKeybinding("actions", gocui.KeyCtrlN, gocui.ModNone, inputView); err != nil {
		return err
	}

	// arrow keys binding to navigate over the list of items.
	if err := g.SetKeybinding("jobs", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	if err := g.SetKeybinding("actions", gocui.KeyArrowUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	if err := g.SetKeybinding("jobs", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}

	if err := g.SetKeybinding("actions", gocui.KeyArrowDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}

	// Mouse bindings to select an item or scroll over the list of items.
	if err := g.SetKeybinding("jobs", gocui.MouseWheelUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	if err := g.SetKeybinding("actions", gocui.MouseWheelUp, gocui.ModNone, cursorUp); err != nil {
		return err
	}

	if err := g.SetKeybinding("jobs", gocui.MouseWheelDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}

	if err := g.SetKeybinding("actions", gocui.MouseWheelDown, gocui.ModNone, cursorDown); err != nil {
		return err
	}

	if err := g.SetKeybinding("jobs", gocui.MouseLeft, gocui.ModNone, mouseLeftClick); err != nil {
		return err
	}

	if err := g.SetKeybinding("actions", gocui.MouseLeft, gocui.ModNone, mouseLeftClick); err != nil {
		return err
	}

	return nil
}

// inputView displays a temporary input box to either add new job or actions.
func inputView(g *gocui.Gui, cv *gocui.View) error {
	maxX, maxY := g.Size()
	var title string
	var name string

	if cv == nil {
		// no focus view.
		title = "New Job"
		name = "addJob"
	} else {
		// process based on current view name.
		switch cv.Name() {
		case "jobs":
			title = "New Job"
			name = "addJob"

		case "actions":
			title = "New Action"
			name = "addAction"
		}
	}

	// construct the input box and position at the center of the screen.
	if inputView, err := g.SetView(name, maxX/2-12, maxY/2, maxX/2+12, maxY/2+2); err != nil {
		if err != gocui.ErrUnknownView {
			log.Println(err)
			return err
		}
		inputView.Title = title
		inputView.FgColor = gocui.ColorYellow
		inputView.SelBgColor = gocui.ColorBlack
		inputView.SelFgColor = gocui.ColorYellow
		inputView.Editable = true

		if _, err := g.SetCurrentView(name); err != nil {
			log.Println(err)
			return err
		}
		g.Cursor = true
		inputView.Highlight = true
		// bind Enter key to copyInput function.
		if err := g.SetKeybinding(name, gocui.KeyEnter, gocui.ModNone, copyInput); err != nil {
			log.Println(err)
			return err
		}

		// bind Ctrl+Q and Escape keys to close the input box.
		if err := g.SetKeybinding(name, gocui.KeyCtrlQ, gocui.ModNone, closeInputView); err != nil {
			log.Println(err)
			return err
		}

		if err := g.SetKeybinding(name, gocui.KeyEsc, gocui.ModNone, closeInputView); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}

// copyInput the input view (iv) buffer and add it to either jobs or actions views.
func copyInput(g *gocui.Gui, iv *gocui.View) error {
	var err error
	// read the view’s buffer from the beginning.
	iv.Rewind()
	// output view.
	var ov *gocui.View

	switch iv.Name() {

	case "addJob":
		ov, _ = g.View("jobs")
		if iv.Buffer() != "" {
			// data typed, add it.
			dbs.addJob(iv.Buffer())
		} else {
			// no data entered, so go back.
			inputView(g, ov)
			return nil
		}
	case "addAction":
		ov, _ = g.View("actions")
		if iv.Buffer() != "" {
			dbs.addAction(iv.Buffer())
		} else {
			inputView(g, ov)
			return nil
		}
	}

	// clear the temporary input view.
	iv.Clear()
	// no input, so disbale cursor.
	g.Cursor = false

	// must delete keybindings before the view, or fatal error.
	g.DeleteKeybindings(iv.Name())
	if err = g.DeleteView(iv.Name()); err != nil {
		return err
	}
	// set the view back.
	if _, err = g.SetCurrentView(ov.Name()); err != nil {
		return err
	}
	switch ov.Name() {
	case "jobs":
		redrawJobs(g, ov)
	case "actions":
		redrawActions(g, ov)
	}
	return err
}

func redrawJobs(g *gocui.Gui, v *gocui.View) {
	// Clear the view of content and redraw it with a fresh data.
	v.Clear()

	// Loop through jobs database to add their ids to the view.
	dbs.lock.RLock()
	for id, _ := range dbs.jobs {
		_, err := fmt.Fprintln(v, id)
		if err != nil {
			log.Println("Error writing to the jobs view:", err)
		}
	}
	dbs.lock.RUnlock()

	// add current focused job id with its output.
	_, cy := v.Cursor()
	l, _ := v.Line(cy)
	if len(l) == 0 {
		return
	}

	outputsView, _ := g.View("outputs")
	topJobOutput := fmt.Sprintf("%s :: Job :: %s", l, dbs.getJob(l))
	fmt.Fprintln(outputsView, topJobOutput)

}

func redrawActions(g *gocui.Gui, v *gocui.View) {
	// Clear the view of content and redraw it with a fresh data.
	v.Clear()

	// Loop through actions to add their ids to the view.
	dbs.lock.RLock()
	for id, _ := range dbs.actions {
		_, err := fmt.Fprintln(v, id)
		if err != nil {
			log.Println("Error writing to the actions view:", err)
		}
	}
	dbs.lock.RUnlock()

	_, cy := v.Cursor()
	l, _ := v.Line(cy)
	if len(l) == 0 {
		return
	}

	outputsView, _ := g.View("outputs")
	topActionOutput := fmt.Sprintf("%s :: Action :: %s", l, dbs.getAction(l))
	fmt.Fprintln(outputsView, topActionOutput)
}

// nextView moves the focus to another view.
func nextView(g *gocui.Gui, v *gocui.View) error {

	cv := g.CurrentView()

	if cv == nil {
		if _, err := g.SetCurrentView("jobs"); err != nil {
			log.Println("Failed to set focus on default (jobs) view:", err)
			return err
		}
		return nil
	}

	switch cv.Name() {

	case "jobs":
		// move the focus on the actions list box.
		if _, err := g.SetCurrentView("actions"); err != nil {
			log.Println("Failed to set focus on actions view:", err)
			return err
		}
	case "actions":
		// move back the focus on the jobs list box.
		if _, err := g.SetCurrentView("jobs"); err != nil {
			log.Println("Failed to set focus on jobs view:", err)
			return err
		}
	}

	return nil
}

// closeInputView close temporary input view and abort change.
func closeInputView(g *gocui.Gui, iv *gocui.View) error {
	// clear the temporary input view.
	iv.Clear()
	// no input, so disbale cursor.
	g.Cursor = false

	// must delete keybindings before the view, or fatal error.
	g.DeleteKeybindings(iv.Name())
	if err := g.DeleteView(iv.Name()); err != nil {
		log.Println("Failed to delete input view:", err)
		return err
	}

	if err := setCurrentDefaultView(g); err != nil {
		return err
	}

	return nil
}

// setCurrentDefaultView moves the focus on default view.
func setCurrentDefaultView(g *gocui.Gui) error {
	// move back the focus on the jobs list box.
	if _, err := g.SetCurrentView("jobs"); err != nil {
		log.Println("Failed to set focus on default view:", err)
		return err
	}
	return nil
}

// lineBelow returns true if there is a non-empty string in cursor position y+1.
func lineBelow(g *gocui.Gui, v *gocui.View) bool {
	_, cy := v.Cursor()
	if l, _ := v.Line(cy + 1); l != "" {
		return true
	}
	return false
}

// cursorDown moves cursor to (currentY + 1) position if there is data there.
func cursorDown(g *gocui.Gui, v *gocui.View) error {

	if v != nil && lineBelow(g, v) == true {
		// there is data to next line.
		v.MoveCursor(0, 1, false)
		// get cursor Y coordinate and data there.
		_, cy := v.Cursor()
		id, _ := v.Line(cy)

		if v.Name() == "jobs" {
			redrawOutputs(g, "jobs", id)

		} else if v.Name() == "actions" {
			redrawOutputs(g, "actions", id)
		}
	}

	return nil
}

// lineAbove returns true if there is a non-empty string in cursor position y-1.
func lineAbove(g *gocui.Gui, v *gocui.View) bool {
	_, cy := v.Cursor()
	if l, _ := v.Line(cy - 1); l != "" {
		return true
	}
	return false
}

// cursorUp moves cursor to (currentY - 1) position if there is data there.
func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil && lineAbove(g, v) == true {
		// there is data upper.
		v.MoveCursor(0, -1, false)
		// get cursor Y coordinate and data there.
		_, cy := v.Cursor()
		id, _ := v.Line(cy)

		if v.Name() == "jobs" {
			redrawOutputs(g, "jobs", id)

		} else if v.Name() == "actions" {
			redrawOutputs(g, "actions", id)
		}
	}

	return nil
}

func redrawOutputs(g *gocui.Gui, sourceViewName, id string) {
	// clean the outputs view and redraw it with focused id data.
	outputsView, _ := g.View("outputs")
	outputsView.Clear()

	// retreive data from database and add it to outputs view.
	switch sourceViewName {
	case "jobs":
		data := dbs.getJob(id)
		_, err := fmt.Fprintln(outputsView, data)
		if err != nil {
			log.Println("Error writing job data to outputs view:", err)
		}
	case "actions":
		data := dbs.getAction(id)
		_, err := fmt.Fprintln(outputsView, data)
		if err != nil {
			log.Println("Error writing action data to outputs view:", err)
		}
	}
}

// mouseLeftClick displays the output data of the item selected.
func mouseLeftClick(g *gocui.Gui, v *gocui.View) error {

	if v != nil {
		_, cy := v.Cursor()
		id, _ := v.Line(cy)

		if len(id) == 0 {
			// click on empty space, fun.
			displayIcon(g)
			return nil
		}

		if v.Name() == "jobs" {
			redrawOutputs(g, "jobs", id)

		} else if v.Name() == "actions" {
			redrawOutputs(g, "actions", id)
		}
	}

	return nil
}

// displayIcon displays symbol at random position.
func displayIcon(g *gocui.Gui) {
	outputsView, _ := g.View("outputs")
	// generate two random values.
	xLines, yLines := outputsView.Size()
	posX, posY := rand.Intn(xLines+1), rand.Intn(yLines+1)
	// change cursor coordinates.
	outputsView.SetCursor(posX, posY)
	// insert symbol at this position.
	outputsView.EditWrite('♣')
}
