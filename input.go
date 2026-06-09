package main

import (
	"fmt"
	"os"
)

func startKeysPressedListener() {
	buf := make([]byte, 3)
	for {
		n, err := os.Stdin.Read(buf)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}

		if n >= 3 && buf[0] == 0x1B && buf[1] == '[' {
			switch buf[2] {
			case 'A':
				handleUpKey()
			case 'B':
				handleDownKey()
			}
			continue
		}

		key := buf[0]
		switch menuState {
		case 0:
			switch key {
			case 'w':
				handleUpKey()
			case 's':
				handleDownKey()
			case '\r':
				startMenuSelected = currentSelection
			}
		case 1:
			switch key {
			case 'q':
				if err = setDongleInBTPairing(false); err != nil {
					fmt.Println(err)
				}
				startMenuSelected = -1
			case '1':
				selectedItemsSearchForNewDevices = 1
				if len(searchDeviceList.pairedDevices) != 0 {
					if err := connectNewDevice(uint16(currentSelection)); err != nil {
						fmt.Println(err)
					} else {
						startMenuSelected = -1
					}
				}
			}
		case 2:
			switch key {
			case 'q':
				startMenuSelected = -1
			case 'w':
				handleUpKey()
			case 's':
				handleDownKey()
			case '1':
				if err := connectDeviceFromPairedlist(uint16(currentSelection)); err != nil {
					fmt.Println(err)
				}
				selectedItemsPairedDevices = 1
			case '2':
				if err := disconnectDeviceFromPairedlist(uint16(currentSelection)); err != nil {
					fmt.Println(err)
				}
				selectedItemsPairedDevices = 2
			case '3':
				if err := removeDeviceFromPairedlist(uint16(currentSelection)); err != nil {
					fmt.Println(err)
				}
				selectedItemsPairedDevices = 3
			case '4':
				if err := clearPairingList(); err != nil {
					fmt.Println(err)
				}
				selectedItemsPairedDevices = 4
			}
		case 3:
			switch key {
			case 'q':
				startMenuSelected = -1
			case 'w':
				handleUpKey()
			case 's':
				handleDownKey()
			case '\r':
				switch dongleSettingsMenu[currentSelection].id {
				case 0:
					autoPairing, _ := getAutoPairing()
					if err := setAutoPairing(!autoPairing); err != nil {
						fmt.Println(err)
					}
					updateDongleSettingsMenu()
				case 1:
					if dongle, exists := deviceManager[selectedDongle]; exists {
						if err := factoryReset(dongle.deviceID); err != nil {
							fmt.Println(err)
						}
						startMenuSelected = -1
					}
				}
			}
		case 4:
			switch key {
			case 'q':
				startMenuSelected = -1
			}
		}
	}
}

func handleUpKey() {
	if currentSelection > 0 {
		currentSelection--
	}
}

func handleDownKey() {
	switch menuState {
	case 0:
		if currentSelection < len(startMenu)-1 {
			currentSelection++
		}
	case 2:
		if dongle, exists := deviceManager[selectedDongle]; exists {
			if currentSelection < len(dongle.pairingList.pairedDevices)-1 {
				currentSelection++
			}
		}
	case 3:
		if currentSelection < len(dongleSettingsMenu)-1 {
			currentSelection++
		}
	}
}
