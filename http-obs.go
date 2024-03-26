package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type ResponseScenes struct {
	Result string `json:"result"`
}

func httpObsStart(c *fiber.Ctx) error {
	_, err := sendOBSRequestWait("StartStream", nil)
	if err != nil {
		return fmt.Errorf("unable to start OBS stream: %s", err.Error())
	}
	return c.JSON(fiber.Map{"result": "OK"})
}

func httpObsStop(c *fiber.Ctx) error {
	_, err := sendOBSRequestWait("StopStream", nil)
	if err != nil {
		return fmt.Errorf("unable to stop OBS stream: %s", err.Error())
	}
	return c.JSON(fiber.Map{"result": "OK"})
}

func httpScenes(c *fiber.Ctx) error {
	res, err := sendOBSRequestWait("GetSceneList", nil)
	if err != nil {
		return fmt.Errorf("unable to get scene list from OBS: %s", err.Error())
	}
	return c.JSON(res)
}

func httpSetScene(c *fiber.Ctx) error {
	data := make(map[string]interface{})
	data["sceneName"] = c.Query("scene")
	_, err := sendOBSRequest("SetCurrentProgramScene", data)
	if err != nil {
		return fmt.Errorf("unable to set current scene: %s", err.Error())
	}
	return c.JSON(fiber.Map{"result": "OK"})
}
