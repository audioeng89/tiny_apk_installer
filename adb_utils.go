package main

import "runtime"

func getAdbExeName() string {
	if runtime.GOOS == "windows" {
		return "adb.exe"
	}
	return "adb"
}
