package main

import (
	"io"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

type ResponseStats struct {
	Result string `json:"result"`

	LiveEncoding    bool `json:"live_encoding"`
	LiveOBS         bool `json:"live_obs"`
	LiveOBSDuration int  `json:"live_obs_duration"`

	Temperature string                       `json:"temp"`
	Interfaces  map[string]ResponseInterface `json:"ifs"`
}

type ResponseInterface struct {
	MBDown int64 `json:"d"`
	MBUp   int64 `json:"u"`
}

var gRegexSpacer = regexp.MustCompile(`\s+`)

func httpStats(w http.ResponseWriter, r *http.Request) {
	// Check if OBS is streaming
	streamStatus, err := sendOBSRequestWait("GetStreamStatus", nil)
	if err != nil {
		writeError(w, 500, "Unable to get OBS stream status: %s", err.Error())
		return
	}

	// Get temperature measurement
	cmd := exec.Command("vcgencmd", "measure_temp")
	res, err := cmd.Output()
	if err != nil {
		writeError(w, 500, "Unable to get temperature: %s", err.Error())
		return
	}

	// Get network transfer statistics
	resInterfaces := make(map[string]ResponseInterface)
	fh, err := os.Open("/proc/net/dev")
	if err != nil {
		writeError(w, 500, "Unable to read network interface statistics")
		return
	}
	data, _ := io.ReadAll(fh)
	ifaces := strings.Split(string(data), "\n")[2:]
	for _, iface := range ifaces {
		parse := strings.Split(gRegexSpacer.ReplaceAllString(strings.TrimLeft(iface, " "), " "), " ")
		if len(parse) < 17 {
			continue
		}

		bytes_down, _ := strconv.ParseInt(parse[1], 10, 64)
		bytes_up, _ := strconv.ParseInt(parse[9], 10, 64)

		if bytes_down == 0 || bytes_up == 0 {
			continue
		}

		mb_down := bytes_down / 1000 / 1000
		mb_up := bytes_up / 1000 / 1000

		iface_name := strings.TrimSuffix(parse[0], ":")
		if iface_name == "lo" {
			continue
		}

		resInterfaces[iface_name] = ResponseInterface{
			MBDown: mb_down,
			MBUp:   mb_up,
		}
	}

	writeResponse(w, ResponseStats{
		Result: "OK",

		LiveEncoding:    gStreamCmd != nil,
		LiveOBS:         streamStatus["outputActive"].(bool),
		LiveOBSDuration: int(streamStatus["outputDuration"].(float64) / 1000.0),

		Temperature: strings.Trim(strings.TrimPrefix(string(res), "temp="), "\r\n"),
		Interfaces:  resInterfaces,
	})
}
