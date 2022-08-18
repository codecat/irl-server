package main

import (
	"net/http"

	"github.com/codecat/go-libs/log"
	"github.com/spf13/viper"
)

func main() {
	viper.AddConfigPath(".")
	viper.ReadInConfig()

	go loopOBS()

	if viper.GetString("video.type") == "gopro" {
		ip, err := findGoProIP()
		if err != nil {
			log.Error("Unable to find GoPro IP address: %s", err.Error())
		} else {
			log.Info("GoPro IP address: %s", ip)
		}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Access-Control-Allow-Origin", "*")
		w.Write([]byte("Nimble IRL"))
	})

	http.HandleFunc("/api/stats", httpStats)

	http.HandleFunc("/api/start", httpStart)
	http.HandleFunc("/api/stop", httpStop)

	http.HandleFunc("/api/obs-start", httpObsStart)
	http.HandleFunc("/api/obs-stop", httpObsStop)

	http.HandleFunc("/api/scenes", httpScenes)
	http.HandleFunc("/api/set-scene", httpSetScene)

	log.Info("Listening on %s", viper.GetString("server.listen"))
	http.ListenAndServe(viper.GetString("server.listen"), nil)
}
