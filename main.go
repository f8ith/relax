package main

import (
	"fmt"
	"net/url"
	"encoding/json"
	"golang.org/x/term"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"syscall"
	"strings"

	"github.com/99designs/keyring"
	wifiname "github.com/yelinaung/wifi-name"
)

type Config struct {
	DefaultProfile string
	Profiles       map[string]Profile
}

type Profile struct {
	Url           string
	Fields        map[string]string
	PasswordField string
	SuccessText   string
	SSID 		  string
}

var config Config

var SetId string

var UserConfigDir, _ = os.UserConfigDir()

var ConfigDir string = path.Join(UserConfigDir, "relax")

var ConfigFile string = path.Join(ConfigDir, "relax.json")

func LoadConfig() {
	if _, err := os.Stat(ConfigFile); os.IsNotExist(err) {
		log.Fatal("LoadConfig: ", err.Error())
	}
	ConfigData, err := os.ReadFile(ConfigFile)
	if err != nil {
		log.Fatal("LoadConfig: ", err.Error())
	}
	err = json.Unmarshal(ConfigData, &config)
	if err != nil {
		log.Fatal("LoadConfig: ", err.Error())
	}
}

func main() {
	LoadConfig()
	var profile Profile
	var profileName string
	if len(os.Args) <= 1 {
		profileName = config.DefaultProfile
	} else {
		profileName = os.Args[1]
	}
	profile = config.Profiles[profileName]
	if profile.SSID == "" || wifiname.WifiName() == profile.SSID {
		ring, _ := keyring.Open(keyring.Config{
			ServiceName: "relax",
		})
		password, err := ring.Get(profileName)
		if err != nil {
			fmt.Println("Enter password: ")
			passwordBytes, _ := term.ReadPassword(int(syscall.Stdin))
			password  := string(passwordBytes)
			_ = ring.Set(keyring.Item{
				Key: profileName,
				Data: []byte(password),
			})
		}
		profile.Fields[profile.PasswordField] = string(password.Data)
//		jsonData, _ := json.Marshal(profile.Fields)
		formData := url.Values{}
		for k, v := range profile.Fields {
		    formData.Add(k, v)
		}
		fmt.Println(profile)
		r, err  := http.PostForm(profile.Url, formData)
		if err != nil {
			log.Fatal("HttpPost: ", err.Error())
		}
		defer r.Body.Close()
		rBytes, _ := io.ReadAll(r.Body)
		if strings.Contains(string(rBytes), profile.SuccessText) {
			fmt.Println("online")
		}
		return
	}
}
