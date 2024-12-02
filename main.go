package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas" // Ensure correct import here
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/terminal"
)

var activeScreen string

func main() {
	// Create a new Fyne application
	myApp := app.New()
	myWindow := myApp.NewWindow("Dreamland")

	var mainMenu fyne.CanvasObject // Placeholder for main menu

	// Create the terminal widget
	t := terminal.New()

	go func() {
		// Run the terminal's local shell
		_ = t.RunLocalShell()
		// Handle cleanup after terminal exits (if needed)
	}()
	terminal := container.NewScroll(t)
	terminalMenu := container.NewBorder(
		nil,      // No top border
		nil,      // No bottom border
		nil,      // No left border
		nil,      // No right border
		terminal, // Scrollable terminal
	)

	// Search menu
	searchBar := widget.NewEntry()

	// Define a more structured layout for search results and borders
	top := canvas.NewText("Search Bar", color.White)        // Top border label for Search
	left := canvas.NewText("Left Border", color.White)      // Left border label
	middle := canvas.NewText("Search Results", color.White) // Main content (results area)

	// Set content inside the borders
	searchResults := container.NewBorder(top, nil, left, nil, middle)

	// The search menu, including search bar and results in a vertical layout
	searchMenu := container.NewVBox(
		searchBar,
		searchResults,
	)

	// Define Main Menu
	mainMenu = container.NewVBox(
		widget.NewButton("Install A Package", func() {
			activeScreen = "Search"
			myWindow.SetContent(searchMenu) // Switch to search menu
		}),
		widget.NewButton("Open A Terminal", func() {
			activeScreen = "Terminal"
			myWindow.SetContent(terminalMenu) // Switch to terminal menu
		}),
		widget.NewButton("Remove", func() { println("Remove selected") }),
	)

	// Define menu items
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
			fyne.NewMenuItem("Install", func() { println("Install selected") }),
		),
		fyne.NewMenu("Exit",
			fyne.NewMenuItem("About", func() { println("About selected") }),
		),
	)

	// Set the menu to the window
	myWindow.SetMainMenu(toolBar)

	// Set initial screen
	activeScreen = "Main"
	myWindow.SetContent(mainMenu)

	// Configure and show the window
	myWindow.Resize(fyne.NewSize(1000, 800))
	myWindow.ShowAndRun()
}
