package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/terminal"
)

var activeScreen string

// item represents a reusable search result component
type item struct {
	widget.BaseWidget
	title   string
	desc    string
	onClick func()
}

// NewItem creates a new search result item
func NewItem(title, desc string, onClick func()) *item {
	it := &item{
		title:   title,
		desc:    desc,
		onClick: onClick,
	}
	it.ExtendBaseWidget(it)
	return it
}

// CreateRenderer implements the fyne.WidgetRenderer interface for item
func (it *item) CreateRenderer() fyne.WidgetRenderer {
	// Title and description text
	titleText := canvas.NewText(it.title, color.White)
	titleText.TextStyle = fyne.TextStyle{Bold: true}
	descText := canvas.NewText(it.desc, color.Gray{Y: 180})

	// Button for interaction
	button := widget.NewButton("Select", func() {
		if it.onClick != nil {
			it.onClick()
		}
	})

	// Background for rounded border
	background := canvas.NewRectangle(color.RGBA{R: 30, G: 30, B: 30, A: 255})
	background.CornerRadius = 10
	background.StrokeWidth = 2
	background.StrokeColor = color.RGBA{R: 100, G: 71, B: 76, A: 255}

	// Content layout
	content := container.NewVBox(titleText, descText, button)
	layout := container.NewMax(background, container.NewPadded(content))

	return widget.NewSimpleRenderer(layout)
}

func main() {
	// Create a new Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow("Dreamland")

	// Terminal widget
	t := terminal.New()
	go func() {
		_ = t.RunLocalShell()
	}()
	terminalMenu := container.NewScroll(t)

	// Generate search results dynamically
	results := []fyne.CanvasObject{
		NewItem("Result 1", "Details about result 1", func() { println("Clicked Result 1") }),
		NewItem("Result 2", "Details about result 2", func() { println("Clicked Result 2") }),
		NewItem("Result 3", "Details about result 3", func() { println("Clicked Result 3") }),
	}

	searchBar := widget.NewEntry()
	searchResults := container.NewVBox(results...)

	// Use a Border layout to structure the search menu
	searchMenu := container.NewBorder(
		searchBar,                          // Top (search bar)
		nil,                                // Bottom
		nil,                                // Left
		nil,                                // Right
		container.NewScroll(searchResults), // Center (search results)
	)

	// Define main menu
	mainMenu := container.NewVBox(
		widget.NewButton("Install A Package", func() {
			activeScreen = "Search"
			myWindow.SetContent(searchMenu)
		}),
		widget.NewButton("Open A Terminal", func() {
			activeScreen = "Terminal"
			myWindow.SetContent(terminalMenu)
		}),
		widget.NewButton("Remove", func() { println("Remove selected") }),
	)

	// Toolbar
	toolBar := fyne.NewMainMenu(
		fyne.NewMenu("Menus",
			fyne.NewMenuItem("Main Menu", func() {
				activeScreen = "Main"
				myWindow.SetContent(mainMenu)
			}),
			fyne.NewMenuItem("Terminal", func() {
				activeScreen = "Terminal"
				myWindow.SetContent(terminalMenu)
			}),
		),
	)

	myWindow.SetMainMenu(toolBar)
	myWindow.SetContent(mainMenu)
	myWindow.Resize(fyne.NewSize(1000, 800))
	myWindow.ShowAndRun()
}
