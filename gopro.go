package main

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

var gRegexIPRange = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+\.`)

func findGoProInterface() (*net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range ifaces {
		if strings.HasPrefix(iface.Name, "enx") {
			return &iface, nil
		}
	}

	return nil, errors.New("unable to find GoPro interface (with enx prefix)")
}

func findGoProIP() (string, error) {
	iface, err := findGoProInterface()
	if err != nil {
		return "", err
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return "", err
	}

	if len(addrs) == 0 {
		return "", fmt.Errorf("found no IP address for the GoPro interface, is dhcp enabled on the interface? (/etc/netplan/10-gopro.yaml)")
	}

	return gRegexIPRange.FindString(addrs[0].String()) + "51", nil
}

func goProAPI(path string) error {
	ip, err := findGoProIP()
	if err != nil {
		return err
	}

	_, err = http.Get("http://" + ip + path)
	if err != nil {
		return err
	}

	return nil
}

func startGoPro() error {
	err := goProAPI("/gp/gpWebcam/START?res=" + viper.GetString("video.gopro.res"))
	if err != nil {
		return err
	}
	return goProAPI(
		fmt.Sprintf(
			"/gp/gpWebcam/SETTINGS?fov=%d&bitrate=%d",
			viper.GetInt("gopro.fov"),
			viper.GetInt("gopro.bitrate")*1000,
		),
	)
}

func stopGoPro() error {
	return goProAPI("/gp/gpWebcam/EXIT")
}
