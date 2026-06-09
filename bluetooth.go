package main

/*
#cgo CFLAGS: -Iheaders
#cgo LDFLAGS: -Llib -ljabra

#include "Common.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"unsafe"
)

// toCPairedDevices converts a Go pairedDevice slice to a C array. The returned
// free function must be called once the C array is no longer needed.
func toCPairedDevices(list []pairedDevice) ([]C.Jabra_PairedDevice, func()) {
	cDevices := make([]C.Jabra_PairedDevice, len(list))
	for i, d := range list {
		cDevices[i].deviceName = C.CString(d.deviceName)
		cDevices[i].isConnected = C.bool(d.isConnected)
		for j := 0; j < 6; j++ {
			cDevices[i].deviceBTAddr[j] = C.uint8_t(d.deviceBTAddr[j])
		}
	}
	return cDevices, func() {
		for _, cd := range cDevices {
			C.free(unsafe.Pointer(cd.deviceName))
		}
	}
}

func searchForNewDevices() error {
	if err := setDongleInBTPairing(true); err != nil {
		return err
	}
	if dongle, exists := deviceManager[selectedDongle]; exists {
		return returnCode(int(C.Jabra_SearchNewDevices(C.ushort(dongle.deviceID))))
	}
	return nil
}

func setDongleInBTPairing(pairing bool) error {
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		return fmt.Errorf("no dongle found")
	}
	if pairing {
		return returnCode(int(C.Jabra_SetBTPairing(C.ushort(dongle.deviceID))))
	}
	return returnCode(int(C.Jabra_StopBTPairing(C.ushort(dongle.deviceID))))
}

func connectNewDevice(pairingID uint16) error {
	if searchDeviceList.count == 0 {
		return nil
	}
	cDevices, free := toCPairedDevices(searchDeviceList.pairedDevices)
	defer free()

	var returnErr error
	if err := returnCode(int(C.Jabra_ConnectNewDevice(C.ushort(pairingID), (*C.Jabra_PairedDevice)(unsafe.Pointer(&cDevices[0]))))); err != nil {
		returnErr = err
	}
	if err := setDongleInBTPairing(false); err != nil {
		returnErr = err
	}
	return returnErr
}

func getSearchDeviceList(deviceID uint16) *pairingList {
	cPairingList := C.Jabra_GetSearchDeviceList(C.ushort(deviceID))
	if cPairingList == nil {
		return nil
	}

	result := &pairingList{
		count:    uint16(cPairingList.count),
		listType: deviceListType(cPairingList.listType),
	}
	cDevices := (*[1 << 30]C.Jabra_PairedDevice)(unsafe.Pointer(cPairingList.pairedDevice))[:cPairingList.count:cPairingList.count]
	for _, cd := range cDevices {
		d := pairedDevice{
			deviceName:  C.GoString(cd.deviceName),
			isConnected: bool(cd.isConnected),
		}
		copy(d.deviceBTAddr[:], C.GoBytes(unsafe.Pointer(&cd.deviceBTAddr[0]), 6))
		result.pairedDevices = append(result.pairedDevices, d)
	}
	C.Jabra_FreePairingList(cPairingList)
	return result
}

func getPairingList(deviceID uint16) *pairingList {
	cPairingList := C.Jabra_GetPairingList(C.ushort(deviceID))
	if cPairingList == nil {
		return &pairingList{
			count:         0,
			listType:      -1,
			pairedDevices: make([]pairedDevice, 0),
		}
	}

	result := &pairingList{
		count:    uint16(cPairingList.count),
		listType: deviceListType(cPairingList.listType),
	}
	cDevices := (*[1 << 30]C.Jabra_PairedDevice)(unsafe.Pointer(cPairingList.pairedDevice))[:cPairingList.count:cPairingList.count]
	for _, cd := range cDevices {
		d := pairedDevice{
			deviceName:  C.GoString(cd.deviceName),
			isConnected: bool(cd.isConnected),
		}
		copy(d.deviceBTAddr[:], C.GoBytes(unsafe.Pointer(&cd.deviceBTAddr[0]), 6))
		result.pairedDevices = append(result.pairedDevices, d)
	}
	C.Jabra_FreePairingList(cPairingList)
	return result
}

func clearPairingList() error {
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		return fmt.Errorf("no dongle found")
	}
	return returnCode(int(C.Jabra_ClearPairingList(C.ushort(dongle.deviceID))))
}

func removeDeviceFromPairedlist(pairingID uint16) error {
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		return fmt.Errorf("no dongle found")
	}
	if dongle.pairingList == nil {
		return nil
	}
	cDevices, free := toCPairedDevices(dongle.pairingList.pairedDevices)
	defer free()
	return returnCode(int(C.Jabra_ClearPairedDevice(C.ushort(pairingID), (*C.Jabra_PairedDevice)(unsafe.Pointer(&cDevices[0])))))
}

func connectDeviceFromPairedlist(pairingID uint16) error {
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		return fmt.Errorf("no dongle found")
	}
	if dongle.pairingList == nil {
		return nil
	}
	cDevices, free := toCPairedDevices(dongle.pairingList.pairedDevices)
	defer free()
	return returnCode(int(C.Jabra_ConnectPairedDevice(C.ushort(pairingID), (*C.Jabra_PairedDevice)(unsafe.Pointer(&cDevices[0])))))
}

func disconnectDeviceFromPairedlist(pairingID uint16) error {
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		return fmt.Errorf("no dongle found")
	}
	if dongle.pairingList == nil {
		return nil
	}
	cDevices, free := toCPairedDevices(dongle.pairingList.pairedDevices)
	defer free()
	return returnCode(int(C.Jabra_DisConnectPairedDevice(C.ushort(pairingID), (*C.Jabra_PairedDevice)(unsafe.Pointer(&cDevices[0])))))
}

func reconnectToDevice() error {
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		return fmt.Errorf("no dongle found")
	}
	return returnCode(int(C.Jabra_ConnectBTDevice(C.ushort(dongle.deviceID))))
}

func disconnectBTDeviceFromDongle() error {
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		return fmt.Errorf("no dongle found")
	}
	return returnCode(int(C.Jabra_DisconnectBTDevice(C.ushort(dongle.deviceID))))
}

func getAutoPairing() (bool, error) {
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		return false, fmt.Errorf("no dongle")
	}
	return bool(C.Jabra_GetAutoPairing(C.ushort(dongle.deviceID))), nil
}

func setAutoPairing(autoPairing bool) error {
	dongle, exists := deviceManager[selectedDongle]
	if !exists {
		return fmt.Errorf("no dongle")
	}
	return returnCode(int(C.Jabra_SetAutoPairing(C.ushort(dongle.deviceID), C.bool(autoPairing))))
}
