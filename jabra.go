package main

/*
#cgo CFLAGS: -Iheaders
#cgo LDFLAGS: -Llib -ljabra

#include "Common.h"
#include "JabraDeviceConfig.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"log"
	"time"
	"unsafe"
)

/****************************************************************************/
/*                             C CALLBACKS                                  */
/****************************************************************************/

//export deviceAttachedFunc
func deviceAttachedFunc(deviceInfo C.Jabra_DeviceInfo) {
	goDeviceInfo := &jabra_DeviceInfo{
		deviceID:               uint16(deviceInfo.deviceID),
		productID:              uint16(deviceInfo.productID),
		vendorID:               uint16(deviceInfo.vendorID),
		deviceName:             C.GoString(deviceInfo.deviceName),
		usbDevicePath:          C.GoString(deviceInfo.usbDevicePath),
		parentInstanceId:       C.GoString(deviceInfo.parentInstanceId),
		errStatus:              errorStatusCode(deviceInfo.errStatus),
		isDongle:               bool(deviceInfo.isDongle),
		dongleName:             C.GoString(deviceInfo.dongleName),
		variant:                C.GoString(deviceInfo.variant),
		serialNumber:           C.GoString(deviceInfo.serialNumber),
		isInFirmwareUpdateMode: bool(deviceInfo.isInFirmwareUpdateMode),
		deviceConnection:       deviceConnectionType(deviceInfo.deviceconnection),
		connectionID:           uint32(deviceInfo.connectionId),
		parentDeviceID:         uint16(deviceInfo.parentDeviceId),
	}
	goDeviceInfo.deviceEventsMask = getDeviceEventsMask(goDeviceInfo.deviceID)
	goDeviceInfo.featureFlags = getSupportedFeature(goDeviceInfo.deviceID)

	if !goDeviceInfo.isDongle {
		battery, err := getBatteryStatus(goDeviceInfo.deviceID)
		if err != nil {
			fmt.Printf("Get Battery Status for %s: %s\n", goDeviceInfo.deviceName, err)
		} else {
			goDeviceInfo.batteryStatus = battery
		}
	} else {
		if goDeviceInfo.featureFlags.pairingList {
			goDeviceInfo.pairingList = getPairingList(goDeviceInfo.deviceID)
		}
	}

	if serialNumberCheck(goDeviceInfo) {
		deviceManager.add(goDeviceInfo)
	}
	C.Jabra_FreeDeviceInfo(deviceInfo)
}

//export deviceRemovedFunc
func deviceRemovedFunc(deviceID uint16) {
	deviceManager.removed(deviceID)
}

/****************************************************************************/
/*                           BACKGROUND POLLERS                             */
/****************************************************************************/

func batteryStatusUpdate() {
	for {
		select {
		case <-stopUpdateBattery:
			return
		default:
			if device, exists := deviceManager[selectedHeadset]; exists {
				battery, err := getBatteryStatus(device.deviceID)
				if err != nil {
					log.Printf("batteryStatusUpdate: %v", err)
				} else {
					// Note: battery percentage can jump when charging due to charge cycle behaviour.
					device.batteryStatus.levelInPercent = battery.levelInPercent
					device.batteryStatus.charging = battery.charging
					device.batteryStatus.batteryLow = battery.batteryLow
					device.batteryStatus.component = battery.component
					device.batteryStatus.extraUnitsCount = battery.extraUnitsCount
					device.batteryStatus.extraUnits = battery.extraUnits
				}
			}
			time.Sleep(time.Second)
		}
	}
}

func updatePairingList() {
	for {
		select {
		case <-stopUpdatePairingList:
			return
		default:
			if dongle, exists := deviceManager[selectedDongle]; exists {
				if updated := getPairingList(dongle.deviceID); updated != nil {
					dongle.pairingList.count = updated.count
					dongle.pairingList.listType = updated.listType
					dongle.pairingList.pairedDevices = updated.pairedDevices
				}
			}
			time.Sleep(time.Second)
		}
	}
}

/****************************************************************************/
/*                             SDK UTILITIES                                */
/****************************************************************************/

func uninitialize() {
	if !C.Jabra_Uninitialize() {
		fmt.Println("Failed Uninitialize")
	}
}

func factoryReset(deviceID uint16) error {
	return returnCode(int(C.Jabra_FactoryReset(C.ushort(deviceID))))
}

func getJabraSdkVersion() string {
	const bufferSize = 16
	buffer := make([]byte, bufferSize)
	cBuffer := (*C.char)(unsafe.Pointer(&buffer[0]))
	if err := returnCode(int(C.Jabra_GetVersion(cBuffer, C.int(bufferSize)))); err != nil {
		log.Fatalln(err)
	}
	return C.GoString(cBuffer)
}

func getDeviceEventsMask(deviceID uint16) uint32 {
	return uint32(C.Jabra_GetSupportedDeviceEvents(C.ushort(deviceID)))
}

func getSupportedFeature(deviceID uint16) *featureFlags {
	var f featureFlags
	var count C.uint32_t

	cFeatures := C.Jabra_GetSupportedFeatures(C.ushort(deviceID), &count)
	if cFeatures == nil {
		return &f
	}
	features := (*[1 << 30]uint32)(unsafe.Pointer(cFeatures))[:count:count]

	setters := map[uint32]*bool{
		1000: &f.busyLight,
		1001: &f.factoryReset,
		1002: &f.pairingList,
		1003: &f.remoteMMI,
		1004: &f.musicEqualizer,
		1005: &f.earbudInterconnectionStatus,
		1006: &f.stepRate,
		1007: &f.heartRate,
		1008: &f.rrInterval,
		1009: &f.ringtoneUpload,
		1010: &f.imageUpload,
		1011: &f.needsExplicitRebootAfterOta,
		1012: &f.needsToBePutIncCradleToCompleteFwu,
		1013: &f.remoteMMIv2,
		1014: &f.logging,
		1015: &f.preferredSoftphoneListInDevice,
		1016: &f.voiceAssistant,
		1017: &f.playRingtone,
		1018: &f.setDateTime,
		1019: &f.fullWizardMode,
		1020: &f.limitedWizardMode,
		1021: &f.onHeadDetection,
		1022: &f.settingsChangeNotification,
		1023: &f.audioStreaming,
		1024: &f.customerSupport,
		1025: &f.mySound,
		1026: &f.uiConfigurableButtons,
		1027: &f.manualBusyLight,
		1028: &f.whiteboard,
		1029: &f.video,
		1030: &f.ambienceModes,
		1031: &f.sealingTest,
		1032: &f.amasupport,
		1033: &f.ambienceModesLoop,
		1034: &f.ffanc,
		1035: &f.googleBisto,
		1036: &f.virtualDirector,
		1037: &f.pictureInPicture,
		1038: &f.dateTimeIsUTC,
		1039: &f.remoteControl,
		1040: &f.userConfigurableHdr,
		1041: &f.dectBasicPairing,
		1042: &f.dectSecurePairing,
		1043: &f.dectOtaFwuSupported,
		1044: &f.xpressURL,
		1045: &f.passwordProvisioning,
		1046: &f.ethernet,
		1047: &f.wlan,
		1048: &f.ethernetAuthenticationCertificate,
		1049: &f.ethernetAuthenticationMschapv2,
		1050: &f.wlanAuthenticationCertificate,
		1051: &f.wlanAuthenticationMschapv2,
	}

	for _, id := range features {
		if b, ok := setters[id]; ok {
			*b = true
		}
	}

	C.Jabra_FreeSupportedFeatures(cFeatures)
	return &f
}
