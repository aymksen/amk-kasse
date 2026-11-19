package main

import (
	"log"
	"syscall"
	"unsafe"
)

// showMessage opens a plain Windows dialog. It blocks until it is closed.
func showMessage(text string) {
	const iconInformation = 0x40
	body, err := syscall.UTF16PtrFromString(text)
	if err != nil {
		log.Print(text)
		return
	}
	title, err := syscall.UTF16PtrFromString("AMK Kasse")
	if err != nil {
		log.Print(text)
		return
	}
	box := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")
	box.Call(0, uintptr(unsafe.Pointer(body)), uintptr(unsafe.Pointer(title)), iconInformation)
}
