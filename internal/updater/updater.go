package updater

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"daxwalkerfix/internal/output"
)

type Release struct {
	Assets []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
		Size        int64  `json:"size"`
	} `json:"assets"`
}

func getFileSize(filename string) int64 {
	info, err := os.Stat(filename)
	if err != nil {
		return 0
	}
	return info.Size()
}

func getLatestRelease() *Release {
	resp, err := http.Get("https://api.github.com/repos/kolief/Dax-Walker-Fix/releases/latest")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()

	var release Release
	json.NewDecoder(resp.Body).Decode(&release)
	return &release
}

func downloadUpdate(release *Release) bool {
	var downloadURL string
	for _, asset := range release.Assets {
		if asset.Name == "daxwalkerfix.exe" {
			downloadURL = asset.DownloadURL
			break
		}
	}

	if downloadURL == "" {
		return false
	}

	resp, err := http.Get(downloadURL)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	newFile := filepath.Join(exeDir, "daxwalkerfix_new.exe")
	
	file, err := os.Create(newFile)
	if err != nil {
		return false
	}
	defer file.Close()

	io.Copy(file, resp.Body)

	return true
}

func updateExe() {
	exePath, _ := os.Executable()
	exeDir := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)
	
	batchContent := fmt.Sprintf(`@echo off
cd /d "%s"
timeout /t 3 /nobreak > nul
if exist "daxwalkerfix_new.exe" (
    if exist "%s" (
        del "%s"
    )
    ren "daxwalkerfix_new.exe" "%s"
)
del "%%~f0"`, exeDir, exeName, exeName, exeName)

	batchFile := filepath.Join(exeDir, "update.bat")
	os.WriteFile(batchFile, []byte(batchContent), 0644)
	exec.Command("cmd", "/C", batchFile).Start()
	os.Exit(0)
}

func Check() {
	release := getLatestRelease()
	if release == nil {
		fmt.Println("Failed to check for updates")
		output.Info("Failed to check for updates")
		return
	}

	var remoteSize int64
	for _, asset := range release.Assets {
		if asset.Name == "daxwalkerfix.exe" {
			remoteSize = asset.Size
			break
		}
	}

	if remoteSize == 0 {
		fmt.Println("No release found")
		output.Info("No release found")
		return
	}

	exePath, _ := os.Executable()
	currentSize := getFileSize(exePath)

	if currentSize != remoteSize {
		fmt.Print("Update available. Download? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		if answer == "y" {
			if downloadUpdate(release) {
				fmt.Println("Update downloaded, restarting...")
				output.Info("Update downloaded, restarting...")
				updateExe()
			} else {
				fmt.Println("Update failed")
				output.Info("Update failed")
			}
		}
	} else {
		fmt.Println("Up to date")
		output.Info("Up to date")
	}
}
