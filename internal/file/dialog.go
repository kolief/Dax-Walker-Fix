package fileselect

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"
)

var (
	comdlg32            = syscall.NewLazyDLL("comdlg32.dll")
	procGetOpenFileName = comdlg32.NewProc("GetOpenFileNameW")
)

type openFileName struct {
	lStructSize       uint32
	hwndOwner         uintptr
	hInstance         uintptr
	lpstrFilter       *uint16
	lpstrCustomFilter *uint16
	nMaxCustFilter    uint32
	nFilterIndex      uint32
	lpstrFile         *uint16
	nMaxFile          uint32
	lpstrFileTitle    *uint16
	nMaxFileTitle     uint32
	lpstrInitialDir   *uint16
	lpstrTitle        *uint16
	flags             uint32
	nFileOffset       uint16
	nFileExtension    uint16
	lpstrDefExt       *uint16
	lCustData         uintptr
	lpfnHook          uintptr
	lpTemplateName    *uint16
}

func SelectProxyFile() (string, error) {
	filter := "Text Files (*.txt)\x00*.txt\x00All Files (*.*)\x00*.*\x00\x00"
	filterUTF16, _ := syscall.UTF16PtrFromString(filter)
	titleUTF16, _ := syscall.UTF16PtrFromString("Select proxy.txt file")
	fileBuffer := make([]uint16, 260)

	ofn := openFileName{
		lStructSize:  uint32(unsafe.Sizeof(openFileName{})),
		lpstrFilter:  filterUTF16,
		lpstrFile:    &fileBuffer[0],
		nMaxFile:     uint32(len(fileBuffer)),
		lpstrTitle:   titleUTF16,
		flags:        0x00080000 | 0x00001000 | 0x00000004,
		nFilterIndex: 1,
	}

	ret, _, _ := procGetOpenFileName.Call(uintptr(unsafe.Pointer(&ofn)))
	if ret == 0 {
		return "", fmt.Errorf("file selection cancelled")
	}

	filename := syscall.UTF16ToString(fileBuffer)
	return filename, nil
}

func SavePathWithType(path string, proxyType int) {
	home, _ := os.UserHomeDir()
	daxDir := filepath.Join(home, "Desktop", "DaxWalkerFix")
	os.MkdirAll(daxDir, 0755)
	rememberFile := filepath.Join(daxDir, "remember.dat")
	data := fmt.Sprintf("%s|%d", path, proxyType)
	os.WriteFile(rememberFile, []byte(data), 0644)
}

func LoadPathWithType() (string, int) {
	home, _ := os.UserHomeDir()
	rememberFile := filepath.Join(home, "Desktop", "DaxWalkerFix", "remember.dat")
	data, err := os.ReadFile(rememberFile)
	if err != nil {
		return "", -1
	}
	content := strings.TrimSpace(string(data))
	parts := strings.Split(content, "|")
	if len(parts) != 2 {
		return "", -1
	}
	path := parts[0]
	if _, err := os.Stat(path); err != nil {
		return "", -1
	}
	proxyType := 0
	if parts[1] == "1" {
		proxyType = 1
	}
	return path, proxyType
}

func LoadPath() string {
	path, _ := LoadPathWithType()
	return path
}
