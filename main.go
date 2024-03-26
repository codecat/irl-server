package main

import (
	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/viper"
)

func main() {
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	log.SetLevel(log.DebugLevel)

	go loopOBS()

	if viper.GetString("video.type") == "gopro" {
		ip, err := findGoProIP()
		if err != nil {
			log.Error("Unable to find GoPro IP address", "err", err.Error())
		} else {
			log.Info("Found GoPro IP address", "ip", ip)
		}
	}

	app := fiber.New()
	app.Use(func(c *fiber.Ctx) error {
		c.Set("Access-Control-Allow-Origin", "*")
		return c.Next()
	})

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Nimble IRL v2")
	})

	app.Get("/api/stats", httpStats)

	app.Get("/api/start", httpStart)
	app.Get("/api/stop", httpStop)

	app.Get("/api/obs-start", httpObsStart)
	app.Get("/api/obs-stop", httpObsStop)

	app.Get("/api/scenes", httpScenes)
	app.Get("/api/set-scene", httpSetScene)

	log.Infof("Listening on %s", viper.GetString("server.listen"))
	app.Listen(viper.GetString("server.listen"))
}
