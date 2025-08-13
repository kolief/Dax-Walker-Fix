package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Release struct {
	Assets []struct {
		Name        string `json:"name"`
		DownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

func GetFileChecksum(filename string) string {
	file, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer file.Close()

	hash := sha256.New()
	io.Copy(hash, file)
	return hex.EncodeToString(hash.Sum(nil))
}

func GetRemoteChecksum() string {
	resp, err := http.Get("https://api.github.com/repos/kolief/Dax-Walker-Fix/releases/latest")
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	var release Release
	json.NewDecoder(resp.Body).Decode(&release)

	for _, asset := range release.Assets {
		if asset.Name == "daxwalkerfix.exe" {
			resp2, err := http.Get(asset.DownloadURL)
			if err != nil {
				return ""
			}
			defer resp2.Body.Close()

			hash := sha256.New()
			io.Copy(hash, resp2.Body)
			return hex.EncodeToString(hash.Sum(nil))
		}
	}
	return ""
}

func DownloadUpdate() bool {
	resp, err := http.Get("https://api.github.com/repos/kolief/Dax-Walker-Fix/releases/latest")
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	var release Release
	json.NewDecoder(resp.Body).Decode(&release)

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

	resp2, err := http.Get(downloadURL)
	if err != nil {
		return false
	}
	defer resp2.Body.Close()

	file, err := os.Create("daxwalkerfix_new.exe")
	if err != nil {
		return false
	}
	defer file.Close()

	io.Copy(file, resp2.Body)

	currentExe, _ := os.Executable()
	os.Rename(currentExe, "daxwalkerfix_old.exe")
	os.Rename("daxwalkerfix_new.exe", "daxwalkerfix.exe")
	os.Remove("daxwalkerfix_old.exe")

	return true
}

func Check() {
	currentHash := GetFileChecksum("daxwalkerfix.exe")
	remoteHash := GetRemoteChecksum()

	if currentHash != remoteHash && remoteHash != "" {
		fmt.Print("Update available. Download? (y/n): ")
		var answer string
		fmt.Scanln(&answer)
		if answer == "y" {
			if DownloadUpdate() {
				fmt.Println("Updated successfully")
			} else {
				fmt.Println("Update failed")
			}
		}
	} else {
		fmt.Println("Up to date")
	}
}