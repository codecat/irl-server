package main

import (
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/codecat/go-libs/log"
	"github.com/spf13/viper"
)

var gStreamCmd *exec.Cmd

func streamMonitor() {
	gStreamCmd.Wait()
	gStreamCmd = nil

	log.Warn("Stream command stopped!")
}

func httpStart(w http.ResponseWriter, r *http.Request) {
	if gStreamCmd != nil {
		writeError(w, 400, "Stream has already started")
		return
	}

	videoType := viper.GetString("video.type")

	// If we're using the GoPro video type, we have to start webcam mode on the GoPro first
	if videoType == "gopro" {
		err := startGoPro()
		if err != nil {
			writeError(w, 500, "Unable to start GoPro: %s", err.Error())
			return
		}
	}

	args := make([]string, 0)

	switch videoType {
	case "gopro":
		// GoPro input (an RTSP stream)
		args = append(args, "-i", "udp://@0.0.0.0:8554?overrun_nonfatal=1&fifo_size=50000000")
		args = append(args, "-c:v", "copy")

	case "webcam":
		// Webcam input
		args = append(args, "-f", "v4l2")
		args = append(args, "-input_format", viper.GetString("video.webcam.format"))
		args = append(args, "-framerate", strconv.Itoa(viper.GetInt("video.webcam.fps")))
		args = append(args, "-video_size", viper.GetString("video.webcam.size"))
		args = append(args, "-i", viper.GetString("video.webcam.path"))

		// Encode using specific bitrate
		args = append(args, "-b:v", viper.GetString("stream.bitrate"))
	}

	// Re-encode audio
	args = append(args, "-c:a", viper.GetString("audio.codec"))
	args = append(args, "-ar", strconv.Itoa(viper.GetInt("audio.rate")))

	// Output to RTMP endpoint
	args = append(args, "-f", "rtsp", viper.GetString("stream.endpoint"))

	log.Info("Stream starting: ffmpeg %s", strings.Join(args, " "))

	gStreamCmd = exec.Command("ffmpeg", args...)
	gStreamCmd.Start()

	go streamMonitor()

	writeOK(w)
}

func httpStop(w http.ResponseWriter, r *http.Request) {
	if gStreamCmd == nil {
		writeError(w, 400, "Stream has not started yet")
		return
	}

	log.Info("Stream stopping")

	// Send interrupt signal to the stream command
	gStreamCmd.Process.Signal(os.Interrupt)

	// If we're using the GoPro video type, we can stop webcam mode here
	if viper.GetString("video.type") == "gopro" {
		err := stopGoPro()
		if err != nil {
			writeError(w, 500, "Unable to stop GoPro: %s", err.Error())
			return
		}
	}

	log.Info("Stream stopped")

	writeOK(w)
}
