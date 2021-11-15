package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/jroimartin/gocui"
)

const (

	// Jobs box width.
	jobswidth = 22
)

var jobsDatabase map[string]string
var actionsDatabase map[string]string

func init() {
	if jobsDatabase == nil {
		jobsDatabase = make(map[string]string)
	}

	if actionsDatabase == nil {
		actionsDatabase = make(map[string]string)
	}
}

func main() {

	runtime.GOMAXPROCS(runtime.NumCPU())

	f, err := os.OpenFile("logs.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Println("file no exists")
	}
	defer f.Close()
	log.SetFlags(log.LstdFlags)
	log.SetOutput(f)

	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		log.Panicln(err)
	}
	defer g.Close()

	// Highlight active view.
	g.Highlight = true
	g.SelFgColor = gocui.ColorRed
	g.BgColor = gocui.ColorBlack
	g.FgColor = gocui.ColorWhite
	//g.Cursor = true

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
	jobsView.SelBgColor = gocui.ColorBlack
	jobsView.SelFgColor = gocui.ColorGreen
	jobsView.Autoscroll = true

	// Outputs view.
	outputsView, err := g.SetView("outputs", jobswidth+1, 0, maxX-1, maxY-4)
	if err != nil && err != gocui.ErrUnknownView {
		log.Println("Failed to create outputs view:", err)
		return
	}
	outputsView.Title = "Outputs"
	outputsView.FgColor = gocui.ColorWhite
	outputsView.SelBgColor = gocui.ColorBlack
	outputsView.SelFgColor = gocui.ColorYellow
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
	actionsView.SelBgColor = gocui.ColorBlack
	actionsView.SelFgColor = gocui.ColorRed

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

// keybindings binds keys to views.
func keybindings(g *gocui.Gui) error {

	if err := g.SetKeybinding("", gocui.KeyCtrlC, gocui.ModNone, quit); err != nil {
		return err
	}

	if err := g.SetKeybinding("", gocui.KeyTab, gocui.ModNone, nextView); err != nil {
		return err
	}

	if err := g.SetKeybinding("jobs", gocui.KeyCtrlA, gocui.ModNone, inputView); err != nil {
		return err
	}

	if err := g.SetKeybinding("actions", gocui.KeyCtrlA, gocui.ModNone, inputView); err != nil {
		return err
	}

	return nil
}

// inputView displays a temporary input box to either add new job or actions.
func inputView(g *gocui.Gui, cv *gocui.View) error {
	maxX, maxY := g.Size()
	var title string
	var name string

	// process based on current view name.
	switch cv.Name() {

	case "jobs":
		title = "New Job"
		name = "addJob"

	case "actions":
		title = "New Action"
		name = "addAction"
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
			addJob(iv.Buffer())
		} else {
			// no data entered, so go back.
			inputView(g, ov)
			return nil
		}
	case "addAction":
		ov, _ = g.View("actions")
		if iv.Buffer() != "" {
			addAction(iv.Buffer())
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

// addJob generates random id for the job and save to jobs store.
func addJob(data string) {
	id := fmt.Sprintf("%x", time.Now().UnixNano())
	output := strings.TrimSpace(data)
	jobsDatabase[id] = output
}

// addAction generates randim id for the action and save to actions store.
func addAction(data string) {
	id := fmt.Sprintf("%x", time.Now().UnixNano())
	output := strings.TrimSpace(data)
	actionsDatabase[id] = output
}

func redrawJobs(g *gocui.Gui, v *gocui.View) {
	// Clear the view of content and redraw it with a fresh data.
	v.Clear()

	// Loop through jobs to add their ids to the view.
	for id, _ := range jobsDatabase {
		_, err := fmt.Fprintln(v, id)
		if err != nil {
			log.Println("Error writing to the jobs view:", err)
		}
	}

	_, cy := v.Cursor()
	l, _ := v.Line(cy)
	if len(l) == 0 {
		return
	}

	outputsView, _ := g.View("outputs")
	fmt.Fprintln(outputsView, jobsDatabase[l])
}

func redrawActions(g *gocui.Gui, v *gocui.View) {
	// Clear the view of content and redraw it with a fresh data.
	v.Clear()

	// Loop through jobs to add their ids to the view.
	for id, _ := range actionsDatabase {
		_, err := fmt.Fprintln(v, id)
		if err != nil {
			log.Println("Error writing to the actions view:", err)
		}
	}

	_, cy := v.Cursor()
	l, _ := v.Line(cy)
	if len(l) == 0 {
		return
	}

	outputsView, _ := g.View("outputs")
	fmt.Fprintln(outputsView, actionsDatabase[l])
}

// nextView moves the focus to another view.
func nextView(g *gocui.Gui, v *gocui.View) error {

	cv := g.CurrentView()

	switch cv.Name() {

	case "nil":
		// move the focus on the jobs list box.
		if _, err := g.SetCurrentView("jobs"); err != nil {
			log.Println("Failed to set focus on default (jobs) view:", err)
			return err
		}
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
