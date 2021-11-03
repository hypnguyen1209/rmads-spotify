package main

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"time"
)

var (
	dirSpotify = ""
)

func argsPowerShell() []string {
	return []string{
		"Compress-Archive",
		"-Path",
		fmt.Sprintf("%s\\Apps\\xpui.js", dirSpotify),
		"-Update",
		"-DestinationPath",
		fmt.Sprintf("%s\\Apps\\xpui.zip", dirSpotify),
	}
}

func addFileToZip() {
	cmd := exec.Command("powershell", argsPowerShell()...)
	_, err := cmd.Output()
	if err != nil {
		log.Fatal(err)
	}
}

func checkSpotifyInstalled() bool {
	if _, err := os.Stat(dirSpotify); os.IsNotExist(err) {
		return false
	}
	return true
}

func downloadConfig() {
	path := fmt.Sprintf("%s\\chrome_elf.zip", dirSpotify)
	file, err := os.Create(path)
	if err != nil {
		log.Fatal("Không thể tạo được file")
	}
	defer file.Close()
	resp, err := http.Get("https://rmads-spotify.hypnguyen.workers.dev")
	if err != nil {
		log.Fatal("Không thể download, vui lòng kiểu tra lại internet")
	}
	defer resp.Body.Close()
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		log.Fatal("Khởi tạo file không thành công")
	}
}

func deleteFile(src string) error {
	err := os.Remove(src)
	if err != nil {
		return err
	}
	return nil
}

func extractFile() {
	src := fmt.Sprintf("%s\\chrome_elf.zip", dirSpotify)
	r, err := zip.OpenReader(src)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	err = os.Rename(fmt.Sprintf("%s\\chrome_elf.dll", dirSpotify), fmt.Sprintf("%s\\chrome_elf.bak.dll", dirSpotify))
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range r.File {
		file, err := f.Open()
		if err != nil {
			log.Println("Lỗi trong quá trình giải nén...")
		}
		defer file.Close()
		name := fmt.Sprintf("%s\\%s", dirSpotify, f.Name)
		create, err := os.Create(name)
		if err != nil {
			log.Println(err)
		}
		defer create.Close()
		create.ReadFrom(file)
	}
	log.Println("Ads removed!")
}

func copyFileBak(originPath string) {
	log.Println("Removing Ads banner and Upgrade button...")
	defer func() {
		if err := recover(); err != nil {
			log.Fatal("Lỗi khi tạo backup")
		}
	}()
	original, _ := os.Open(originPath)
	defer original.Close()
	new, _ := os.Create(fmt.Sprintf("%s.bak", originPath))
	defer new.Close()
	_, err := io.Copy(new, original)
	if err != nil {
		log.Fatal("Lỗi khi tạo backup")
	}
}

func writeFileXpui(data string) {
	f, err := os.Create(fmt.Sprintf("%s\\Apps\\xpui.js", dirSpotify))
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	_, err = f.WriteString(data)
	if err != nil {
		log.Fatal(err)
	}
}

func extractXpuiJS() {
	src := fmt.Sprintf("%s\\Apps\\xpui.spa", dirSpotify)
	r, err := zip.OpenReader(src)
	if err != nil {
		log.Fatal(err)
	}
	defer r.Close()
	for _, f := range r.File {
		file, err := f.Open()
		if err != nil {
			log.Printf("Lỗi trong quá trình giải nén %s...\n", src)
		}
		defer file.Close()
		if f.Name == "xpui.js" {
			name := fmt.Sprintf("%s\\Apps\\xpui.txt", dirSpotify)
			create, err := os.Create(name)
			if err != nil {
				log.Fatal(err)
			}
			defer create.Close()
			create.ReadFrom(file)
			break
		}
	}
	data, err := ioutil.ReadFile(fmt.Sprintf("%s\\Apps\\xpui.txt", dirSpotify))
	if err != nil {
		log.Fatal("Lỗi đọc file xpui.js")
	}
	m := regexp.MustCompile(`(\.ads\.leaderboard\.isEnabled)(}|\))`)
	res := m.ReplaceAllString(string(data), "$1&&false$2")
	m = regexp.MustCompile(`\.createElement\([^.,{]+,{onClick:[^.,]+,className:[^.]+\.[^.]+\.UpgradeButton}\),[^.(]+\(\)`)
	res = m.ReplaceAllString(res, "")
	writeFileXpui(res)
}

func remoteBanner() {
	src := fmt.Sprintf("%s\\Apps\\xpui.spa", dirSpotify)
	copyFileBak(src)
	extractXpuiJS()
}

func main() {
	fmt.Println("*********************************************************************")
	fmt.Println("*                                                                   *")
	fmt.Println("*                                                                   *")
	fmt.Println("*                       REMOVE ADS SPOTIFY                          *")
	fmt.Println("*                                                                   *")
	fmt.Println("*                                                                   *")
	fmt.Println("*********************************************************************")
	fmt.Println("")
	log.Println("Checking Spotify....")
	dirAppdata, err := os.UserConfigDir()
	if err != nil {
		log.Fatal(err)
	}
	dirSpotify = dirAppdata + "\\Spotify"
	isInstalled := checkSpotifyInstalled()
	if !isInstalled {
		log.Fatal("Spotify chưa được cài đặt...")
		log.Fatal("Nếu bạn cài đặt bản trên M$ Store thì vui lòng gỡ và cài đặt bản dành cho desktop...")
	}
	downloadConfig()
	extractFile()
	time.Sleep(1 * time.Second)
	deleteFile(fmt.Sprintf("%s\\config.zip", dirSpotify))
	remoteBanner()
	time.Sleep(1 * time.Second)
	deleteFile(fmt.Sprintf("%s\\Apps\\xpui.txt", dirSpotify))
	err = os.Rename(fmt.Sprintf("%s\\Apps\\xpui.spa", dirSpotify), fmt.Sprintf("%s\\Apps\\xpui.zip", dirSpotify))
	if err != nil {
		log.Fatal(err)
	}
	addFileToZip()
	err = os.Rename(fmt.Sprintf("%s\\Apps\\xpui.zip", dirSpotify), fmt.Sprintf("%s\\Apps\\xpui.spa", dirSpotify))
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second)
	deleteFile(fmt.Sprintf("%s\\Apps\\xpui.js", dirSpotify))
	log.Println("Done!")
	fmt.Scanln()
}
