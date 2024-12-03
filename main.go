package main

import (
	"fmt"
	"image/color"

	quickTools "dreamland/backend"

	"sort"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/fyne-io/terminal"
)

var activeScreen string

// Package represents a reusable package with its information
type Package struct {
	Name        string
	Description string
	Author      string
	Popularity  float32
	Version     string
}

// Convert PackageInfo to Package
func convertToPackage(pkgInfo quickTools.PackageInfo) Package {
	return Package{
		Name:        pkgInfo.Name,
		Description: pkgInfo.Description,
		Author:      pkgInfo.Author,
		Popularity:  pkgInfo.Popularity,
		Version:     pkgInfo.Version,
	}
}

// GenerateScreen generates a detailed screen for a selected package
func GenerateScreen(pkg Package, myWindow fyne.Window) fyne.CanvasObject {
	// Title
	Title := widget.NewLabelWithStyle(pkg.Name, fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Description
	Description := widget.NewLabel(pkg.Description)

	// Author
	author := widget.NewLabel(fmt.Sprintf("Author: %s", pkg.Author))

	// Version
	version := widget.NewLabel(fmt.Sprintf("Version: %s", pkg.Version))

	// Layout the components
	content := container.NewVBox(
		Title,
		Description,
		author,
		version,
		widget.NewButton("Install", func() {
			quickTools.Install(pkg.Name) // Install the package
		}),
		widget.NewButton("Back", func() {
			myWindow.SetContent(mainMenu(myWindow)) // Go back to the main menu
		}),
	)

	// Scrollable container for long content
	return container.NewScroll(content)
}

// NewItem creates a new search result item
func NewItem(Title, Description string, onClick func()) *item {
	it := &item{
		Title:       Title,
		Description: Description,
		onClick:     onClick,
	}
	it.ExtendBaseWidget(it)
	return it
}

// item represents a reusable search result component
type item struct {
	widget.BaseWidget
	Title       string
	Description string
	onClick     func()
}

// CreateRenderer implements the fyne.WidgetRenderer interface for item
func (it *item) CreateRenderer() fyne.WidgetRenderer {
	// Title and Description text
	TitleText := canvas.NewText(it.Title, color.White)
	TitleText.TextStyle = fyne.TextStyle{Bold: true}
	DescriptionText := canvas.NewText(it.Description, color.Gray{Y: 180})

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
	background.StrokeColor = color.RGBA{R: 255, G: 182, B: 193, A: 255}

	// Content layout
	TitleText.Color = color.RGBA{R: 255, G: 182, B: 193, A: 255}

	content := container.NewVBox(TitleText, DescriptionText, button)
	layout := container.NewMax(background, container.NewPadded(content))

	return widget.NewSimpleRenderer(layout)
}

// Search screen
func searchMenu(myWindow fyne.Window) fyne.CanvasObject {
	// Generate search results dynamically
	searchResults := container.NewVBox() // Dynamic search results container
	searchBar := widget.NewEntry()
	searchBar.PlaceHolder = "Type your search here..."
	searchBar.OnSubmitted = func(text string) {
		go func() {
			// Perform the search asynchronously
			results, err := quickTools.Search(text)
			if err != nil {
				fyne.CurrentApp().SendNotification(&fyne.Notification{
					Title:   "Error",
					Content: fmt.Sprintf("Failed to search: %v", err),
				})
				return
			}

			// Sort by popularity
			sort.Slice(results, func(i, j int) bool {
				return results[i].Popularity > results[j].Popularity
			})

			// Update the UI in the main thread
			searchResults.Objects = nil // Clear old results

			for _, pkg := range results {
				pkgCopy := pkg // Avoid closure issues
				convertedPkg := convertToPackage(pkgCopy)
				searchResults.Add(NewItem(pkgCopy.Name, pkgCopy.Description, func() {
					myWindow.SetContent(GenerateScreen(convertedPkg, myWindow))
				}))
			}
			searchResults.Refresh() // Refresh the container to show updates
		}()
	}

	// Combine search bar and results
	return container.NewBorder(
		searchBar,                          // Top
		nil,                                // Bottom
		nil,                                // Left
		nil,                                // Right
		container.NewScroll(searchResults), // Center
	)
}

func sortByPopularity(packages []Package) {
	sort.Slice(packages, func(i, j int) bool {
		return packages[i].Popularity > packages[j].Popularity // Descending order
	})
}

// Main Menu
func mainMenu(myWindow fyne.Window) fyne.CanvasObject {
	return container.NewVBox(
		widget.NewButton("Install A Package", func() {
			activeScreen = "Search"
			myWindow.SetContent(searchMenu(myWindow)) // Set content to the search menu
		}),
		widget.NewButton("Open A Terminal", func() {
			activeScreen = "Terminal"
			// Implement terminal menu here
		}),
		widget.NewButton("Remove", func() { println("Remove selected") }),
	)
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

	// Initially show the main menu
	myWindow.SetContent(mainMenu(myWindow))

	// Toolbar
	toolBar := fyne.NewMainMenu(
		fyne.NewMenu("Menus",
			fyne.NewMenuItem("Main Menu", func() {
				activeScreen = "Main"
				myWindow.SetContent(mainMenu(myWindow))
			}),
			fyne.NewMenuItem("Terminal", func() {
				activeScreen = "Terminal"
				myWindow.SetContent(terminalMenu)
			}),
		),
	)

	myWindow.SetMainMenu(toolBar)
	myWindow.Resize(fyne.NewSize(1000, 800))
	myWindow.ShowAndRun()
}
