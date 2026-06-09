package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

var (
	loading      = [10]string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	loadingIndex = 0

	horizontalLine = "" +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" +
		"━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" // 240 total, 3 bytes per ━

	verticalLine      = "┃"
	leftCornerTop     = "┏"
	rightCornerTop    = "┓"
	leftCornerBottom  = "┗"
	rightCornerBottom = "┛"

	width, height = 0, 0

	resetCurrentSelection = false
	currentSelection      = 0
	menuState             = 0
	startMenuSelected     = -1

	selectedItemsPairedDevices = -1
	menuItemsPairedDevices     = [5]string{"Q Back", "1 Connect", "2 Disconnect", "3 Remove", "4 Clear"}

	selectedItemsSearchForNewDevices = -1
	menuItemsSearchForNewDevices     = [2]string{"Q Back", "1 Connect"}
)

const (
	batteryFullChar     = "◼"
	batteryEmptyChar    = "◻"
	batteryWidth        = 10
	lowBatteryThreshold = 20
)

func enableRawMode() (*unix.Termios, error) {
	fd := int(os.Stdin.Fd())

	oldSettings, err := unix.IoctlGetTermios(fd, unix.TCGETS)
	if err != nil {
		return nil, err
	}

	newSettings := *oldSettings
	newSettings.Lflag &^= unix.ECHO | unix.ICANON
	newSettings.Iflag &^= unix.ICRNL

	if err := unix.IoctlSetTermios(fd, unix.TCSETS, &newSettings); err != nil {
		return nil, err
	}

	return oldSettings, nil
}

func restoreTerminal(oldSettings *unix.Termios) {
	unix.IoctlSetTermios(int(os.Stdin.Fd()), unix.TCSETS, oldSettings)
}

func moveCursor(row, col int) {
	fmt.Printf("\033[%d;%dH", row, col)
}

func clearScreen() {
	fmt.Print("\033[2J")
	fmt.Print("\033[H")
}

func getScreenSize() {
	getWidth, getHeight, err := term.GetSize(1)
	if err != nil {
		log.Fatalln(err)
	}
	width, height = getWidth, getHeight
}

func drawingBox() {
	calcHeight := height - 4
	calcWidth := (width - 11) * 3
	if calcWidth > len(horizontalLine) {
		return
	}

	// Using a pre-computed string slice is faster than strings.Repeat.
	// The factor 3 accounts for each Unicode character taking 3 bytes.
	moveCursor(3, 5)
	fmt.Printf("%s%s%s", leftCornerTop, horizontalLine[:calcWidth], rightCornerTop)

	for i := 4; i < calcHeight; i++ {
		moveCursor(i, 5)
		fmt.Print(verticalLine)
		moveCursor(i, width-5)
		fmt.Print(verticalLine)
		moveCursor(i, 6)
	}

	moveCursor(calcHeight, 5)
	fmt.Printf("%s%s%s", leftCornerBottom, horizontalLine[:(width-11)*3], rightCornerBottom)
}

func header() {
	moveCursor(2, 5)
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		fmt.Printf("Looking For Dongle %s", loading[loadingIndex])
		loadingIndex = (loadingIndex + 1) % len(loading)
		return
	}
	fmt.Printf("%s", dongle.deviceName)

	headset, exists := deviceManager[selectedHeadset]
	if !exists {
		moveCursor(2, width-25)
		fmt.Printf("Looking For HeadSet %s", loading[loadingIndex])
		loadingIndex = (loadingIndex + 1) % len(loading)
		return
	}
	if headset.batteryStatus == nil {
		return
	}

	levelInPercent := headset.batteryStatus.levelInPercent
	filledSegments := int(math.Round(float64(levelInPercent) / 100 * batteryWidth))
	emptySegments := batteryWidth - filledSegments

	var color string
	switch {
	case headset.batteryStatus.batteryLow:
		color = "\033[31m"
	case levelInPercent <= 65:
		color = "\033[33m"
	default:
		color = "\033[32m"
	}

	batteryBar := color +
		strings.Repeat(batteryFullChar, filledSegments) +
		strings.Repeat(batteryEmptyChar, emptySegments) +
		"\033[0m"

	if headset.batteryStatus.charging {
		moveCursor(2, width-50)
		fmt.Printf("%s - Battery : [%s]🗲 %d%%", headset.deviceName, batteryBar, levelInPercent)
	} else {
		moveCursor(2, width-48)
		fmt.Printf("%s - Battery: [%s] %d%%", headset.deviceName, batteryBar, levelInPercent)
	}
}

func menu(width int) {
	resetCurrentSelection = false
	drawingBox()
	for i, option := range startMenu {
		mid := (width - len(option.label)) / 2

		if i == currentSelection {
			moveCursor(5+i, mid-1)
			fmt.Println("\033[42m", option.label, "\033[0m")
		} else {
			moveCursor(5+i, mid)
			fmt.Println(option.label)
		}
	}
}

func pollSearchDeviceList() {
	for {
		if menuState != 1 {
			return
		}
		if dongle, exists := deviceManager[selectedDongle]; exists {
			if updated := getSearchDeviceList(dongle.deviceID); updated != nil {
				searchDeviceList.count = updated.count
				searchDeviceList.listType = updated.listType
				searchDeviceList.pairedDevices = updated.pairedDevices
			}
		}
		time.Sleep(time.Second)
	}
}

func menuSearchForNewDevices() {
	if !resetCurrentSelection {
		currentSelection = 0
		resetCurrentSelection = true
		if err := searchForNewDevices(); err != nil {
			fmt.Println(err)
		}
		go pollSearchDeviceList()
	}

	drawingBox()

	if len(searchDeviceList.pairedDevices) != 0 {
		for i, pd := range searchDeviceList.pairedDevices {
			moveCursor(4+i, 10)
			device := fmt.Sprintf("%d %s", i+1, pd.deviceName)
			if i == currentSelection {
				fmt.Println("\033[42m", device, "\033[0m")
			} else {
				fmt.Println(device)
			}
		}
	}

	calcWidth := 0
	for i, item := range menuItemsSearchForNewDevices {
		moveCursor(height-3, 7+calcWidth)
		if i == selectedItemsSearchForNewDevices {
			fmt.Println("\033[44m", item, "\033[0m")
			go func() {
				time.Sleep(time.Millisecond * 200)
				selectedItemsSearchForNewDevices = -1
			}()
		} else {
			fmt.Println("\033[42m", item, "\033[0m")
		}
		calcWidth += len(item) + 3
	}
}

func menuPairedDevices() {
	if !resetCurrentSelection {
		currentSelection = 0
		resetCurrentSelection = true
	}

	drawingBox()

	if dongle, exists := deviceManager[selectedDongle]; exists {
		for i, pd := range dongle.pairingList.pairedDevices {
			moveCursor(4+i, 10)
			device := fmt.Sprintf("%d %s", i+1, pd.deviceName)
			if pd.isConnected {
				device += " (Connected)"
			}
			if i == currentSelection {
				fmt.Println("\033[42m", device, "\033[0m")
			} else {
				fmt.Println(device)
			}
		}

		calcWidth := 0
		for i, item := range menuItemsPairedDevices {
			moveCursor(height-3, 7+calcWidth)
			if i == selectedItemsPairedDevices {
				fmt.Println("\033[44m", item, "\033[0m")
				go func() {
					time.Sleep(time.Millisecond * 200)
					selectedItemsPairedDevices = -1
				}()
			} else {
				fmt.Println("\033[42m", item, "\033[0m")
			}
			calcWidth += len(item) + 3
		}
	} else {
		startMenuSelected = -1
	}
}

func dongleSettings() {
	if !resetCurrentSelection {
		currentSelection = 0
		resetCurrentSelection = true
	}

	drawingBox()

	for i, item := range dongleSettingsMenu {
		if i == currentSelection {
			moveCursor(4+i, 9)
			fmt.Println("\033[42m", item.label, "\033[0m")
		} else {
			moveCursor(4+i, 10)
			fmt.Println(item.label)
		}
	}

	moveCursor(height-3, 7)
	fmt.Println("\033[42m", "Q Back", "\033[0m")
}

func startUi() {
	sigChan := make(chan os.Signal, 1)
	go func() {
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	}()

	for {
		select {
		case <-sigChan:
			return
		default:
			clearScreen()
			getScreenSize()
			header()

			if startMenuSelected != -1 {
				switch startMenu[startMenuSelected].id {
				case 0:
					menuState = 1
					menuSearchForNewDevices()
				case 1:
					menuState = 2
					menuPairedDevices()
				case 2:
					menuState = 3
					dongleSettings()
				case 3:
					moveCursor(4, 5)
					fmt.Println("Switch Device")
					menuState = 4
				case 4:
					fmt.Println("HeadSet Settings")
				case 5:
					return
				}
			} else {
				menuState = 0
				menu(width)
			}

			time.Sleep(time.Second / 12) // 12 FPS
		}
	}
}
