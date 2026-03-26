package main

type DiscoveredDevicesMsg struct {
	Devices []Device
	Err     error
}

type PairingResultMsg struct {
	Success bool
	Error   string
}

type InstallResultMsg struct {
	Success bool
	Error   string
}

type SplashContinueMsg struct{}

type ADBPathSelectedMsg struct {
	Path string
	Err  error
}

type ADBDownloadMsg struct{}

type ADBDownloadResultMsg struct {
	Path string
	Err  error
}
