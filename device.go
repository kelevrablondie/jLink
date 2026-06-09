package main

import (
	"fmt"
	"log"
)

var (
	deviceManager   devices
	selectedHeadset = -1
	selectedDongle  = -1

	startMenu          = []menuItem{}
	dongleSettingsMenu = []menuItem{}

	searchDeviceList = &pairingList{
		count:         0,
		listType:      searchResult,
		pairedDevices: make([]pairedDevice, 0),
	}

	stopUpdateBattery     = make(chan struct{})
	stopUpdatePairingList = make(chan struct{})
)

func updateDongleSettingsMenu() {
	dongleSettingsMenu = []menuItem{}

	if dongle, exists := deviceManager[selectedDongle]; exists {
		autoPairing, _ := getAutoPairing()
		label := "AutoPairing OFF"
		if autoPairing {
			label = "AutoPairing ON"
		}
		dongleSettingsMenu = append(dongleSettingsMenu, menuItem{id: 0, label: label})

		if dongle.featureFlags.factoryReset {
			dongleSettingsMenu = append(dongleSettingsMenu, menuItem{id: 1, label: "Factory Reset"})
		}
	}
}

func updateStartMenu() {
	startMenu = []menuItem{}

	if dongle, dongleExists := deviceManager[selectedDongle]; dongleExists {
		startMenu = append(startMenu, menuItem{id: 0, label: "Search For New Devices"})
		if dongle.featureFlags.pairingList && dongle.pairingList.count != 0 {
			startMenu = append(startMenu, menuItem{id: 1, label: "See Remembered Paired Devices"})
		}
		startMenu = append(startMenu, menuItem{id: 2, label: fmt.Sprintf("%s Settings", dongle.deviceName)})
	}

	// TODO: Device switching (requires len(deviceManager) > 2)
	// if len(deviceManager) > 2 {
	// 	startMenu = append(startMenu, menuItem{id: 3, label: "Switch Device"})
	// }

	// TODO: Headset settings
	// if device, deviceExists := deviceManager[selectedHeadset]; deviceExists {
	// 	startMenu = append(startMenu, menuItem{id: 4, label: fmt.Sprintf("%s Settings", device.deviceName)})
	// }

	startMenu = append(startMenu, menuItem{id: 5, label: "Exit"})
}

func serialNumberCheck(deviceInfo *jabra_DeviceInfo) bool {
	if deviceInfo.serialNumber == "" {
		return false
	}

	if deviceInfo.isDongle {
		if dongle, exists := deviceManager[selectedDongle]; exists {
			if dongle.serialNumber == deviceInfo.serialNumber && dongle.deviceConnection == deviceInfo.deviceConnection {
				dongle.deviceID = deviceInfo.deviceID
				return false
			}
			return true
		}
		if selectedDongle != -1 {
			log.Printf("selectedDongle=%d has no matching entry in deviceManager", selectedDongle)
		}
		return selectedDongle == -1
	}

	if device, exists := deviceManager[selectedHeadset]; exists {
		if device.serialNumber == deviceInfo.serialNumber && device.deviceConnection == deviceInfo.deviceConnection {
			device.deviceID = deviceInfo.deviceID
			return false
		}
		return true
	}
	if selectedHeadset != -1 {
		log.Printf("selectedHeadset=%d has no matching entry in deviceManager", selectedHeadset)
	}
	return selectedHeadset == -1
}

func (d *devices) add(deviceInfo *jabra_DeviceInfo) {
	if *d == nil {
		*d = make(map[int]*jabra_DeviceInfo)
	}

	id := len(*d)
	if deviceInfo.isDongle {
		if selectedDongle == -1 {
			selectedDongle = id
			go updatePairingList()
		}
	} else {
		if selectedHeadset == -1 {
			selectedHeadset = id
			go batteryStatusUpdate()
		}
	}

	(*d)[id] = deviceInfo
	updateStartMenu()
	updateDongleSettingsMenu()
}

func (d *devices) removed(deviceID uint16) {
	if *d == nil {
		return
	}

	var (
		checkDongleExists  bool
		checkHeadSetExists bool
	)

	newDevices := make(map[int]*jabra_DeviceInfo)
	nextIndex := 0

	for i := 0; i < len(*d); i++ {
		device, exists := (*d)[i]
		if !exists || device.deviceID == deviceID {
			continue
		}

		if device.isDongle {
			checkDongleExists = true
			selectedDongle = nextIndex
		} else {
			checkHeadSetExists = true
			selectedHeadset = nextIndex
		}

		newDevices[nextIndex] = device
		nextIndex++
	}

	if !checkDongleExists {
		stopUpdatePairingList <- struct{}{}
		selectedDongle = -1
	}
	if !checkHeadSetExists {
		stopUpdateBattery <- struct{}{}
		selectedHeadset = -1
	}

	*d = newDevices
	updateStartMenu()
}
