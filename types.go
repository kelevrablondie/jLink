package main

type jabra_DeviceInfo struct {
	deviceID               uint16
	productID              uint16
	vendorID               uint16
	deviceName             string
	usbDevicePath          string
	parentInstanceId       string
	errStatus              errorStatusCode
	isDongle               bool
	dongleName             string
	variant                string
	serialNumber           string
	isInFirmwareUpdateMode bool
	deviceConnection       deviceConnectionType
	connectionID           uint32
	parentDeviceID         uint16
	deviceEventsMask       uint32
	featureFlags           *featureFlags
	batteryStatus          *batteryStatus
	pairingList            *pairingList
}

type batteryComponent int

const (
	unknown       batteryComponent = iota
	headband
	combinde
	right
	left
	cradleBattery
	remoteControlBattery
)

type batteryStatusUnit struct {
	levelInPercent uint8
	component      batteryComponent
}

type batteryStatus struct {
	levelInPercent  uint8
	charging        bool
	batteryLow      bool
	component       batteryComponent
	extraUnitsCount uint32
	extraUnits      []batteryStatusUnit
}

type deviceListType int

const (
	searchResult   deviceListType = iota
	pairedDevices
	searchComplete
)

type pairedDevice struct {
	deviceName   string
	deviceBTAddr [6]byte
	isConnected  bool
}

type pairingList struct {
	count         uint16
	listType      deviceListType
	pairedDevices []pairedDevice
}

type secureConnectionMode int

const (
	legacyMode     secureConnectionMode = iota
	secureMode
	restrictedMode
)

type featureFlags struct {
	busyLight                          bool
	factoryReset                       bool
	pairingList                        bool
	remoteMMI                          bool
	musicEqualizer                     bool
	earbudInterconnectionStatus        bool
	stepRate                           bool
	heartRate                          bool
	rrInterval                         bool
	ringtoneUpload                     bool
	imageUpload                        bool
	needsExplicitRebootAfterOta        bool
	needsToBePutIncCradleToCompleteFwu bool
	remoteMMIv2                        bool
	logging                            bool
	preferredSoftphoneListInDevice     bool
	voiceAssistant                     bool
	playRingtone                       bool
	setDateTime                        bool
	fullWizardMode                     bool
	limitedWizardMode                  bool
	onHeadDetection                    bool
	settingsChangeNotification         bool
	audioStreaming                     bool
	customerSupport                    bool
	mySound                            bool
	uiConfigurableButtons              bool
	manualBusyLight                    bool
	whiteboard                         bool
	video                              bool
	ambienceModes                      bool
	sealingTest                        bool
	amasupport                         bool
	ambienceModesLoop                  bool
	ffanc                              bool
	googleBisto                        bool
	virtualDirector                    bool
	pictureInPicture                   bool
	dateTimeIsUTC                      bool
	remoteControl                      bool
	userConfigurableHdr                bool
	dectBasicPairing                   bool
	dectSecurePairing                  bool
	dectOtaFwuSupported                bool
	xpressURL                          bool
	passwordProvisioning               bool
	ethernet                           bool
	wlan                               bool
	ethernetAuthenticationCertificate  bool
	ethernetAuthenticationMschapv2     bool
	wlanAuthenticationCertificate      bool
	wlanAuthenticationMschapv2         bool
}

type hidInput int

const (
	undefined hidInput = iota
	offHook
	mute
	flash
	redial
	key0
	key1
	key2
	key3
	key4
	key5
	key6
	key7
	key8
	key9
	keyStar
	keyPound
	keyClear
	online
	speedDial
	voiceMail
	lineBusy
	rejectCall
	outOfRange
	pseudoOffHook
	button1
	button2
	button3
	volumeUp
	volumeDown
	fireAlarm
	jackConnection
	qdConnection
	headsetConnection
)

type devices map[int]*jabra_DeviceInfo

type deviceConnectionType int

const (
	deviceConnectionType_USB  deviceConnectionType = iota
	deviceConnectionType_BT
	deviceConnectionType_DECT
)

type menuItem struct {
	id    int
	label string
}
