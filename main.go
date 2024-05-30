package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

const (
	goURL    = "https://go.dev/dl/go1.22.3.src.tar.gz"
	hashURL  = "https://go.dev/dl/?mode=json"
	filename = "go1.22.3.src.tar.gz"
)

type File struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Sha256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

type GoRelease struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []File `json:"files"`
}

func downloadFile(url, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

func getExpectedHash(url, filename string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var releases []GoRelease
	if err := json.NewDecoder(resp.Body).Decode(&releases); err != nil {
		return "", err
	}

	for _, release := range releases {
		for _, file := range release.Files {
			if file.Filename == filename {
				return file.Sha256, nil
			}
		}
	}
	return "", fmt.Errorf("hash for file %s not found", filename)
}

func calculateSHA256(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func main() {
	if err := downloadFile(goURL, filename); err != nil {
		fmt.Println("Error downloading file:", err)
		return
	}

	expectedHash, err := getExpectedHash(hashURL, filename)
	if err != nil {
		fmt.Println("Error getting expected hash:", err)
		return
	}

	actualHash, err := calculateSHA256(filename)
	if err != nil {
		fmt.Println("Error calculating hash:", err)
		return
	}

	isValid := actualHash == expectedHash

	fmt.Println(isValid)
}
