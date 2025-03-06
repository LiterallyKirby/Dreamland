package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var desktopDirs = []string{
	filepath.Join(os.Getenv("HOME"), ".local/share/applications"),
	"/usr/share/applications",
	filepath.Join(os.Getenv("HOME"), ".local/share/flatpak/exports/share/applications"),
}

var myWindow fyne.Window
var mainMenu *fyne.Container

func main() {
	myApp := app.New()
	myWindow = myApp.NewWindow("Desktop File Maker")

	// Entry widgets for each field
	Type_Label := widget.NewLabel("Type*")
	Type := widget.NewEntry()
	Type.SetPlaceHolder("Enter App Type Here...")

	Version_Label := widget.NewLabel("App Version")
	Version := widget.NewEntry()
	Version.SetPlaceHolder("Enter App Version Here...")

	Name_Label := widget.NewLabel("App Name*")
	Name := widget.NewEntry()
	Name.SetPlaceHolder("Enter App Name Here...")

	Comment_Label := widget.NewLabel("App Comment")
	Comment := widget.NewEntry()
	Comment.SetPlaceHolder("Enter App Comment Here...")

	Exec_Label := widget.NewLabel("App Exec*")
	Exec := widget.NewEntry()
	Exec.SetPlaceHolder("Enter App Exec Path Here...")

	Icon_Label := widget.NewLabel("App Icon")
	Icon := widget.NewEntry()
	Icon.SetPlaceHolder("Enter App Icon Path Here...")

	BackButton := widget.NewButton("Back", func() {
		myWindow.SetContent(mainMenu)
	})

	// Button to open file picker (ai made cus i cant be bothered)
	IconPickerButton := widget.NewButton("Browse...", func() {
		fileDialog := dialog.NewFileOpen(func(reader fyne.URIReadCloser, err error) {
			if err != nil {
				log.Println("File selection error:", err)
				return
			}
			if reader == nil {
				return // User canceled
			}

			// Get the selected file path
			selectedFile := reader.URI().Path()
			Icon.SetText(selectedFile)
		}, myWindow)

		// Set the file filter to only show images
		fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".png", ".jpg", ".jpeg", ".svg", ".ico"}))
		fileDialog.Show()
	})

	Terminal_Label := widget.NewLabel("Does the app run in a Terminal?")
	Terminal := widget.NewCheck("Terminal?", func(b bool) {})

	Category_Label := widget.NewLabel("App's Categories")
	Category := widget.NewEntry()
	Category.SetPlaceHolder("Enter App Categories Here...")

	StartupWMClass_Label := widget.NewLabel("App's StartupWMClass (Recommended)")
	StartupWMClass := widget.NewEntry()
	StartupWMClass.SetPlaceHolder("Enter App's StartupWMClass Here...")

	// Set window size
	myWindow.Resize(fyne.NewSize(900, 600))

	// Create the UI layout
	makeDesktop := container.NewVBox(
		Type_Label, Type,
		Version_Label, Version,
		Name_Label, Name,
		Comment_Label, Comment,
		Exec_Label, Exec,
		Icon_Label, Icon, IconPickerButton,
		Terminal_Label, Terminal,
		Category_Label, Category,
		StartupWMClass_Label, StartupWMClass,

		widget.NewButton("Make Desktop", func() {
			// Get data from user input
			appName := Name.Text
			execPath := Exec.Text
			iconPath := Icon.Text
			comment := Comment.Text
			terminal := Terminal.Checked
			categories := Category.Text
			startupWMClass := StartupWMClass.Text

			Name.Refresh()
			Exec.Refresh()
			Icon.Refresh()
			Comment.Refresh()
			Terminal.Refresh()
			Category.Refresh()
			StartupWMClass.Refresh()

			// Create desktop file content
			desktopFileContent := fmt.Sprintf(`
			
[Desktop Entry]
Name=%s
Exec=%s
Icon=%s
Type=Application
Categories=%s
Terminal=%t
Comment=%s
StartupWMClass=%s
`, appName, execPath, iconPath, categories, terminal, comment, startupWMClass)

			// Save the desktop file
			err := os.WriteFile(appName+".desktop", []byte(desktopFileContent), 0644)
			if err != nil {
				log.Println("Error writing desktop file:", err)
			} else {
				log.Println("Desktop file created:", appName+".desktop")
			}
			// Mark the file as executable
			err = os.Chmod(appName+".desktop", 0755)
			if err != nil {
				log.Println("Error setting file as executable:", err)
			} else {
				log.Println("Desktop file is now executable")
			}

			// Copy the .desktop file to the trusted applications directory
			desktopDir := os.Getenv("HOME") + "/.local/share/applications/"
			err = ioutil.WriteFile(desktopDir+appName+".desktop", []byte(desktopFileContent), 0755)
			if err != nil {
				log.Println("Error copying desktop file to trusted directory:", err)
			} else {
				log.Println("Desktop file added to trusted directory")
			}
			//update the desktop data base
			cmd := exec.Command("update-desktop-database", "~/.local/share/applications")
			err = cmd.Run()
			if err != nil {
				log.Println("Error updating desktop database:", err)
			} else {
				log.Println("Desktop database updated")
			}
		}), BackButton,
	)

	createMenuButton := widget.NewButton("Make A Desktop", func() {
		myWindow.SetContent(makeDesktop)
	})
	removeMenuButton := widget.NewButton("Remove A Desktop", func() {
		//myWindow.SetContent(m)
		myWindow.SetContent(container.NewVScroll(RefreshRemoveFiles()))

	})
	editMenuButton := widget.NewButton("Edit A Desktop", func() {
		//myWindow.SetContent(m)
		fmt.Println("Not Made Yet")
	})

	mainMenu = container.NewVBox(createMenuButton, removeMenuButton, editMenuButton)

	myWindow.SetContent(mainMenu)
	myWindow.ShowAndRun()
}

func CollectingFiles() []string {
	fmt.Println("Searching for .desktop files...")

	var desktops []string

	for _, dir := range desktopDirs {
		err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				log.Println("Error accessing:", path, "->", err)
				return nil
			}
			if !d.IsDir() && filepath.Ext(path) == ".desktop" {
				fmt.Println("Found:", path)
				desktops = append(desktops, path)
			}
			return nil
		})
		if err != nil {
			log.Println("Error walking directory:", dir, "->", err)
		}
	}
	return desktops
}

func RefreshRemoveFiles() *fyne.Container {
	desktopFiles := CollectingFiles() // Assume this function gets the .desktop file paths
	content := container.NewVBox()
	BackButton := widget.NewButton("Back", func() {
		myWindow.SetContent(mainMenu)
	})
	content.Add(BackButton)

	for _, path := range desktopFiles {
		iconName := getIconPath(path) // Get the icon name from the desktop file

		// Try to construct a valid icon path
		iconPath := filepath.Join(filepath.Dir(path), iconName) // Check if it's in the same dir
		if !fileExists(iconPath) {                              // If not, check system icon dirs
			iconPath = resolveIconPath(iconName)
		}

		// Load icon resource
		icon, err := fyne.LoadResourceFromPath(iconPath)
		if err != nil {
			log.Println("Failed to load icon:", err)
			icon = theme.FyneLogo() // Use fallback icon
		}

		// Correct closure issue by capturing `path` in a new variable
		pathCopy := path

		// Create button with icon
		btn := widget.NewButtonWithIcon(filepath.Base(pathCopy), icon, func() {
			// Open a confirmation dialog before removing the file
			dialog.ShowConfirm("Remove File", "Are you sure you want to remove this desktop entry? (This WON'T actually uninstall the App)",
				func(confirm bool) {
					if confirm {
						err := os.Remove(pathCopy)
						if err != nil {
							log.Println("Error removing file:", err)
						} else {
							log.Println("Removed:", pathCopy)
							myWindow.SetContent(container.NewVScroll(RefreshRemoveFiles())) // Refresh UI after deletion
						}
					}
				}, myWindow)
		})

		content.Add(btn)
	}

	return content
}

// Helper function to resolve icon paths from system directories
func resolveIconPath(iconName string) string {
	commonPaths := []string{
		"/usr/share/icons/hicolor/256x256/apps/",
		"/usr/share/pixmaps/",
		"~/.local/share/icons/",
	}

	for _, dir := range commonPaths {
		fullPath := filepath.Join(dir, iconName+".png") // Assume PNG first
		if fileExists(fullPath) {
			return fullPath
		}
	}

	return "" // No valid path found
}

// Helper function to check if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func getIconPath(desktopFilePath string) string {
	file, err := os.Open(desktopFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Icon=") {
			return strings.TrimPrefix(line, "Icon=")
		}
	}
	return ""
}
