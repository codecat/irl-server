package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

var gRegexSpacer = regexp.MustCompile(`\s+`)

func httpStats(c *fiber.Ctx) error {
	// Check if OBS is streaming
	streamStatus, err := sendOBSRequestWait("GetStreamStatus", nil)
	if err != nil {
		return fmt.Errorf("unable to get OBS stream status: %s", err.Error())
	}

	// Get temperature measurement
	cmd := exec.Command("vcgencmd", "measure_temp")
	res, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("unable to get temperature: %s", err.Error())
	}

	// Get network transfer statistics
	resInterfaces := make(map[string]fiber.Map)
	fh, err := os.Open("/proc/net/dev")
	if err != nil {
		return fmt.Errorf("unable to read network interface statistics")
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

		resInterfaces[iface_name] = fiber.Map{
			"d": mb_down,
			"u": mb_up,
		}
	}

	return c.JSON(fiber.Map{
		"result": "OK",

		"live_encoding":     gStreamCmd != nil,
		"live_obs":          streamStatus["outputActive"].(bool),
		"live_obs_duration": int(streamStatus["outputDuration"].(float64) / 1000.0),

		"temp": strings.Trim(strings.TrimPrefix(string(res), "temp="), "\r\n"),
		"ifs":  resInterfaces,
	})
}
