package main

/*
#cgo CFLAGS: -Iheaders
#cgo LDFLAGS: -Llib -ljabra

#include "Common.h"
*/
import "C"
import "unsafe"

func getBatteryStatus(deviceID uint16) (*batteryStatus, error) {
	var cBatteryStatus *C.Jabra_BatteryStatus

	if err := returnCode(int(C.Jabra_GetBatteryStatusV2(C.ushort(deviceID), &cBatteryStatus))); err != nil {
		return nil, err
	}

	result := &batteryStatus{
		levelInPercent:  uint8(cBatteryStatus.levelInPercent),
		charging:        bool(cBatteryStatus.charging),
		batteryLow:      bool(cBatteryStatus.batteryLow),
		component:       batteryComponent(cBatteryStatus.component),
		extraUnitsCount: uint32(cBatteryStatus.extraUnitsCount),
	}

	if cBatteryStatus.extraUnitsCount > 0 && cBatteryStatus.extraUnits != nil {
		extraUnits := (*[1 << 30]C.Jabra_BatteryStatusUnit)(unsafe.Pointer(cBatteryStatus.extraUnits))[:cBatteryStatus.extraUnitsCount:cBatteryStatus.extraUnitsCount]
		for _, unit := range extraUnits {
			result.extraUnits = append(result.extraUnits, batteryStatusUnit{
				levelInPercent: uint8(unit.levelInPercent),
				component:      batteryComponent(unit.component),
			})
		}
	}

	C.Jabra_FreeBatteryStatus(cBatteryStatus)
	return result, nil
}
