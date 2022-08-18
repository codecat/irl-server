package main

import "net/http"

type ResponseScenes struct {
	Result string `json:"result"`
}

func httpObsStart(w http.ResponseWriter, r *http.Request) {
	_, err := sendOBSRequestWait("StartStream", nil)
	if err != nil {
		writeError(w, 500, "Unable to start OBS stream: %s", err.Error())
		return
	}
	writeOK(w)
}

func httpObsStop(w http.ResponseWriter, r *http.Request) {
	_, err := sendOBSRequestWait("StopStream", nil)
	if err != nil {
		writeError(w, 500, "Unable to stop OBS stream: %s", err.Error())
		return
	}
	writeOK(w)
}

func httpScenes(w http.ResponseWriter, r *http.Request) {
	res, err := sendOBSRequestWait("GetSceneList", nil)
	if err != nil {
		writeError(w, 500, "Unable to get scene list from OBS: %s", err.Error())
		return
	}
	writeResponse(w, res)
}

func httpSetScene(w http.ResponseWriter, r *http.Request) {
	values := r.URL.Query()

	data := make(map[string]interface{})
	data["sceneName"] = values.Get("scene")
	_, err := sendOBSRequest("SetCurrentProgramScene", data)
	if err != nil {
		writeError(w, 500, "Unable to set current scene: %s", err.Error())
		return
	}
	writeOK(w)
}
