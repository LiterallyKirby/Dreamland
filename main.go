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
	"sync"

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

var iconCache = make(map[string]string)

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
		myWindow.SetContent(container.NewVScroll(EditMenu()))
	})

	mainMenu = container.NewVBox(createMenuButton, removeMenuButton, editMenuButton)

	myWindow.SetContent(mainMenu)
	myWindow.ShowAndRun()
}

func CollectingFiles() []string {
	fmt.Println("Searching for .desktop files...")

	var desktops []string
	var mu sync.Mutex // Mutex to synchronize access to the desktops slice

	// Create a channel to collect the results
	results := make(chan string)

	var wg sync.WaitGroup // WaitGroup to wait for all goroutines to finish

	// Iterate through each directory
	for _, dir := range desktopDirs {
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					log.Println("Error accessing:", path, "->", err)
					return nil
				}
				if !d.IsDir() && filepath.Ext(path) == ".desktop" {
					results <- path
				}
				return nil
			})
			if err != nil {
				log.Println("Error walking directory:", dir, "->", err)
			}
		}(dir)
	}

	// Close the results channel once all goroutines are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Collect results from the channel
	for path := range results {
		mu.Lock()
		desktops = append(desktops, path)
		mu.Unlock()
	}

	return desktops
}

func EditMenu() *fyne.Container {
	desktopFiles := CollectingFiles()
	content := container.NewVBox()
	BackButton := widget.NewButton("Back", func() {
		myWindow.SetContent(mainMenu)
	})
	content.Add(BackButton)

	// Create a button for each desktop file
	for _, path := range desktopFiles {
		iconName := getIconPath(path)
		iconPath := filepath.Join(filepath.Dir(path), iconName)
		if !fileExists(iconPath) {
			iconPath = resolveIconPath(iconName)
		}

		// Load icon resource
		icon, err := fyne.LoadResourceFromPath(iconPath)
		if err != nil {
			log.Println("Failed to load icon:", err)
			icon = theme.FyneLogo() // Use fallback icon
		}

		// Create a button for each .desktop file
		btn := widget.NewButtonWithIcon(filepath.Base(path), icon, func() {
			// When clicked, open the EditForm for the selected file
			myWindow.SetContent(container.NewVScroll(EditDesktopForm(path)))
		})

		content.Add(btn)
	}

	return content
}

func EditDesktopForm(filePath string) *fyne.Container {
	// Load the content of the selected .desktop file
	fileContent, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Println("Error reading file:", err)
		return nil
	}

	// Parse the .desktop file into a map for easy access
	desktopData := parseDesktopFile(fileContent)

	// Create entry widgets for each field
	nameEntry := widget.NewEntry()
	nameEntry.SetText(desktopData["Name"])

	execEntry := widget.NewEntry()
	execEntry.SetText(desktopData["Exec"])

	iconEntry := widget.NewEntry()
	iconEntry.SetText(desktopData["Icon"])

	commentEntry := widget.NewEntry()
	commentEntry.SetText(desktopData["Comment"])

	categoriesEntry := widget.NewEntry()
	categoriesEntry.SetText(desktopData["Categories"])

	terminalEntry := widget.NewEntry()
	terminalEntry.SetText(desktopData["Terminal"])

	// Button to save the edited file
	saveButton := widget.NewButton("Save", func() {
		// Save the edited content back to the .desktop file
		editedContent := fmt.Sprintf(`
[Desktop Entry]
Name=%s
Exec=%s
Icon=%s
Type=Application
Categories=%s
Terminal=%s
Comment=%s
`, nameEntry.Text, execEntry.Text, iconEntry.Text, categoriesEntry.Text, terminalEntry.Text, commentEntry.Text)

		err := ioutil.WriteFile(filePath, []byte(editedContent), 0644)
		if err != nil {
			log.Println("Error writing to file:", err)
			return
		}

		// Mark the file as executable
		err = os.Chmod(filePath, 0755)
		if err != nil {
			log.Println("Error setting file as executable:", err)
		} else {
			log.Println("Desktop file updated and set to executable")
		}

		// Update the desktop database
		cmd := exec.Command("update-desktop-database", "~/.local/share/applications")
		err = cmd.Run()
		if err != nil {
			log.Println("Error updating desktop database:", err)
		} else {
			log.Println("Desktop database updated")
		}

		// Go back to the edit menu after saving
		myWindow.SetContent(EditMenu())
	})

	// Button to go back without saving
	backButton := widget.NewButton("Back", func() {
		myWindow.SetContent(container.NewVScroll(EditMenu())) // Go back to the list of desktop files
	})

	return container.NewVBox(
		widget.NewLabel("Edit Desktop File"),
		widget.NewLabel("Name:"),
		nameEntry,
		widget.NewLabel("Exec:"),
		execEntry,
		widget.NewLabel("Icon:"),
		iconEntry,
		widget.NewLabel("Comment:"),
		commentEntry,
		widget.NewLabel("Categories"),
		categoriesEntry,
		widget.NewLabel("Terminal"),
		terminalEntry,
		saveButton,
		backButton,
	)
}

func parseDesktopFile(content []byte) map[string]string {
	desktopData := make(map[string]string)
	scanner := bufio.NewScanner(strings.NewReader(string(content)))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Name=") {
			desktopData["Name"] = strings.TrimPrefix(line, "Name=")
		} else if strings.HasPrefix(line, "Exec=") {
			desktopData["Exec"] = strings.TrimPrefix(line, "Exec=")
		} else if strings.HasPrefix(line, "Icon=") {
			desktopData["Icon"] = strings.TrimPrefix(line, "Icon=")
		} else if strings.HasPrefix(line, "Comment=") {
			desktopData["Comment"] = strings.TrimPrefix(line, "Comment=")
		} else if strings.HasPrefix(line, "Categories=") {
			desktopData["Categories"] = strings.TrimPrefix(line, "Categories=")
		} else if strings.HasPrefix(line, "Terminal=") {
			desktopData["Terminal"] = strings.TrimPrefix(line, "Terminal=")
		}
	}

	return desktopData
}

func RefreshRemoveFiles() *fyne.Container {
	desktopFiles := CollectingFiles()
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
	// Check if we have a cached icon path for the file
	if cachedIcon, exists := iconCache[desktopFilePath]; exists {
		return cachedIcon
	}

	// If not cached, open the file and parse it
	file, err := os.Open(desktopFilePath)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return ""
	}
	defer file.Close()

	// Parse the file for the icon path
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Icon=") {
			iconPath := strings.TrimPrefix(line, "Icon=")
			// Cache the icon path for future use
			iconCache[desktopFilePath] = iconPath
			return iconPath
		}
	}
	return ""
}
