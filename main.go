package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

type Config struct {
	QBittorrentUser   string `yaml:"qbittorrent_user" env:"QBITTORRENT_USER"`
	QBittorrentPass   string `yaml:"qbittorrent_pass" env:"QBITTORRENT_PASS"`
	QBittorrentServer string `yaml:"qbittorrent_server" env:"QBITTORRENT_SERVER"`
	QBittorrentPort   int    `yaml:"qbittorrent_port" env:"QBITTORRENT_PORT"`
	GluetunServer     string `yaml:"gluetun_server" env:"GLUETUN_SERVER"`
	GluetunPort       int    `yaml:"gluetun_port" env:"GLUETUN_PORT"`
}

type QBittorrentPreferences struct {
	ListenPort int `json:"listen_port"`
}

type GluetunPortResponse struct {
	Port int `json:"port"`
}

func loadConfig(configyaml string) (*Config, error) {
	config := &Config{
		QBittorrentServer: "localhost",
		QBittorrentPort:   8080,
		GluetunServer:     "localhost",
		GluetunPort:       8000,
	}

	// Load from YAML file if exists
	if _, err := os.Stat(configyaml); err == nil {
		data, err := os.ReadFile(configyaml)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// Override with environment variables if set
	if val, exists := os.LookupEnv("QBITTORRENT_USER"); exists {
		config.QBittorrentUser = val
	}
	if val, exists := os.LookupEnv("QBITTORRENT_PASS"); exists {
		config.QBittorrentPass = val
	}
	if val, exists := os.LookupEnv("QBITTORRENT_SERVER"); exists {
		config.QBittorrentServer = val
	}
	if val, exists := os.LookupEnv("QBITTORRENT_PORT"); exists {
		config.QBittorrentPort, _ = strconv.Atoi(val)
	}
	if val, exists := os.LookupEnv("GLUETUN_SERVER"); exists {
		config.GluetunServer = val
	}
	if val, exists := os.LookupEnv("GLUETUN_PORT"); exists {
		config.GluetunPort, _ = strconv.Atoi(val)
	}

	return config, nil
}

func findListenPort(config *Config, cookie string) (int, error) {
	url := fmt.Sprintf("http://%s:%d/api/v2/app/preferences", config.QBittorrentServer, config.QBittorrentPort)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Referer", fmt.Sprintf("http://%s:%d", config.QBittorrentServer, config.QBittorrentPort))
	req.Header.Set("Cookie", cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	var prefs QBittorrentPreferences
	if err := json.NewDecoder(resp.Body).Decode(&prefs); err != nil {
		return 0, err
	}

	return prefs.ListenPort, nil
}

func findFwdPort(config *Config) (int, error) {
	url := fmt.Sprintf("http://%s:%d/v1/openvpn/portforwarded", config.GluetunServer, config.GluetunPort)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	var portResp GluetunPortResponse
	if err := json.NewDecoder(resp.Body).Decode(&portResp); err != nil {
		return 0, err
	}

	return portResp.Port, nil
}

func qbtLogin(config *Config) (string, error) {
	url := fmt.Sprintf("http://%s:%d/api/v2/auth/login", config.QBittorrentServer, config.QBittorrentPort)
	data := fmt.Sprintf("username=%s&password=%s", config.QBittorrentUser, config.QBittorrentPass)
	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(data))
	req.Header.Set("Referer", fmt.Sprintf("http://%s:%d", config.QBittorrentServer, config.QBittorrentPort))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	cookies := resp.Header["Set-Cookie"]
	for _, cookie := range cookies {
		if strings.Contains(cookie, "SID=") {
			return cookie, nil
		}
	}

	return "", fmt.Errorf("failed to get session ID")
}

func qbtUpdatePort(config *Config, cookie string, port int) error {
	url := fmt.Sprintf("http://%s:%d/api/v2/app/setPreferences", config.QBittorrentServer, config.QBittorrentPort)
	data := fmt.Sprintf(`json={"listen_port":%d,"random_port":false,"upnp":false}`, port)
	req, _ := http.NewRequest("POST", url, bytes.NewBufferString(data))
	req.Header.Set("Referer", fmt.Sprintf("http://%s:%d", config.QBittorrentServer, config.QBittorrentPort))
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-OK HTTP status: %d", resp.StatusCode)
	}

	return nil
}

func findFwdPortFromFile(fwdFile string) (int, error) {
	fwdPortFile, err := os.Open(fwdFile)
	if err != nil {
		return 0, fmt.Errorf("unable to open forwarded port file: %s", err)
	}
	defer fwdPortFile.Close()

	fwdPortBytes, err := io.ReadAll(fwdPortFile)
	if err != nil {
		return 0, fmt.Errorf("unable to read forwarded port file: %s", err)
	}

	fwdPort, err := strconv.Atoi(strings.TrimSpace(string(fwdPortBytes)))
	if err != nil {
		return 0, fmt.Errorf("invalid forwarded port: %s", err)
	}

	return fwdPort, nil
}

func update(config *Config, fwdFile string) error {
	cookie, err := qbtLogin(config)
	if err != nil {
		return fmt.Errorf("unable to log in to qBittorrent: %s", err)
	}

	var fwdPort int

	if fwdFile != "" {
		fwdPort, err = findFwdPortFromFile(fwdFile)
		if err != nil {
			return fmt.Errorf("unable to find port number in file '%s': %s", fwdFile, err)
		}
	} else {
		fwdPort, err = findFwdPort(config)
		if err != nil {
			return fmt.Errorf("unable to find forwarded port in Gluetun: %s", err)
		}
	}

	if fwdPort < 1 || fwdPort > 65535 {
		return fmt.Errorf("invalid port number found for forwarded port: %d", fwdPort)
	}

	listenPort, err := findListenPort(config, cookie)
	if err != nil {
		return fmt.Errorf("unable to find listening port in Qbittorrent: %s", err)
	}

	if listenPort != fwdPort {
		if err := qbtUpdatePort(config, cookie, fwdPort); err != nil {
			return fmt.Errorf("port update failed: %s", err)
		} else {
			log.Printf("qBittorrent listen port updated: %d -> %d", listenPort, fwdPort)
		}
	} else {
		log.Printf("qBittorrent listen port (%d) already up to date", listenPort)
	}

	return nil
}

func main() {
	configYaml := flag.String("f", "config.yaml", "Path to config")
	fwdFile := flag.String("i", "", "Path to forward port file")
	job := flag.Bool("j", false, "Run as a job, updating once")
	updateFreq := flag.Int("u", 600, "Update frequency in seconds")
	flag.Parse()

	config, err := loadConfig(*configYaml)
	if err != nil {
		log.Fatal(err)
	}

	if *job {
		err := update(config, *fwdFile)
		if err != nil {
			log.Println(err)
		}
	} else {
		for {
			err := update(config, *fwdFile)
			if err != nil {
				log.Println(err)
			}
			time.Sleep(time.Duration(*updateFreq) * time.Second)
		}
	}
}
