package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

var gStreamCmd *exec.Cmd

func streamMonitor() {
	gStreamCmd.Wait()
	gStreamCmd = nil

	log.Warn("Stream command stopped!")
}

func httpStart(c *fiber.Ctx) error {
	if gStreamCmd != nil {
		return fiber.NewError(fiber.StatusBadRequest, "stream has already started")
	}

	videoType := viper.GetString("video.type")

	// If we're using the GoPro video type, we have to start webcam mode on the GoPro first
	if videoType == "gopro" {
		err := startGoPro()
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "unable to start GoPro: %s", err.Error())
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

	log.Info("Stream starting", "ffmpeg", strings.Join(args, " "))

	gStreamCmd = exec.Command("ffmpeg", args...)
	gStreamCmd.Start()

	go streamMonitor()

	return c.JSON(fiber.Map{"result": "OK"})
}

func httpStop(c *fiber.Ctx) error {
	if gStreamCmd == nil {
		return fiber.NewError(fiber.StatusBadRequest, "stream has not started yet")
	}

	log.Info("Stream stopping")

	// Send interrupt signal to the stream command
	gStreamCmd.Process.Signal(os.Interrupt)

	// If we're using the GoPro video type, we can stop webcam mode here
	if viper.GetString("video.type") == "gopro" {
		err := stopGoPro()
		if err != nil {
			return fmt.Errorf("unable to stop GoPro: %s", err.Error())
		}
	}

	log.Info("Stream stopped")

	return c.JSON(fiber.Map{"result": "OK"})
}
