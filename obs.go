package main

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"time"

	"github.com/charmbracelet/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

var gOBSWsc *websocket.Conn
var gOBSRequests = make(map[string]chan map[string]interface{})

func generateAuthentication(challenge, salt string) string {
	secretHash := sha256.Sum256([]byte(viper.GetString("obs.password") + salt))
	secretHashBase64 := base64.StdEncoding.EncodeToString(secretHash[:])
	authenticationHash := sha256.Sum256([]byte(secretHashBase64 + challenge))
	return base64.StdEncoding.EncodeToString(authenticationHash[:])
}

type OBSMessage struct {
	OpCode int         `json:"op"`
	Data   interface{} `json:"d"`
}

func (msg OBSMessage) get(key string) interface{} {
	return msg.Data.(map[string]interface{})[key]
}

func newOBSMessage(opcode int, data interface{}) OBSMessage {
	return OBSMessage{
		OpCode: opcode,
		Data:   data,
	}
}

type OBSIdentify struct {
	RpcVersion         int    `json:"rpcVersion"`
	Authentication     string `json:"authentication,omitempty"`
	EventSubscriptions int    `json:"eventSubscriptions,omitempty"`
}

type OBSRequest struct {
	RequestType string      `json:"requestType"`
	RequestID   string      `json:"requestId"`
	RequestData interface{} `json:"requestData,omitempty"`
}

func newOBSRequestMessage(requestType string, data interface{}) (string, OBSMessage) {
	rid := uuid.NewString()
	return rid, newOBSMessage(6, OBSRequest{
		RequestType: requestType,
		RequestID:   rid,
		RequestData: data,
	})
}

func sendOBSRequest(requestType string, data interface{}) (string, error) {
	rid, msg := newOBSRequestMessage(requestType, data)
	return rid, gOBSWsc.WriteJSON(msg)
}

func sendOBSRequestWait(requestType string, data interface{}) (map[string]interface{}, error) {
	if gOBSWsc == nil {
		return nil, errors.New("websocket is not ininitialized or connected")
	}

	rid, msg := newOBSRequestMessage(requestType, data)

	c := make(chan map[string]interface{})
	gOBSRequests[rid] = c

	err := gOBSWsc.WriteJSON(msg)
	if err != nil {
		return nil, err
	}

	return <-c, nil
}

func loopOBS() {
	firstWait := false

	for {
		gOBSWsc = nil

		if firstWait {
			time.Sleep(1 * time.Second)
		}
		firstWait = true

		log.Info("Connecting to OBS websocket..")

		var err error
		gOBSWsc, _, err = websocket.DefaultDialer.Dial(viper.GetString("obs.address"), nil)
		if err != nil {
			log.Error("Unable to open websocket to OBS", "err", err.Error())
			continue
		}

		for {
			var msg OBSMessage
			err = gOBSWsc.ReadJSON(&msg)
			if err != nil {
				log.Error("Unable to read message from websocket", "err", err.Error())
				break
			}

			switch msg.OpCode {
			case 0: // Hello
				mapAuth := msg.get("authentication").(map[string]interface{})
				challenge := mapAuth["challenge"].(string)
				salt := mapAuth["salt"].(string)

				gOBSWsc.WriteJSON(newOBSMessage(1, OBSIdentify{
					RpcVersion:     1,
					Authentication: generateAuthentication(challenge, salt),
				}))

			case 2: // Identified
				log.Info("Identified with OBS!")
				//sendOBSRequest("GetSceneList", nil)

			case 5: // Event
				//typeName := msg.get("eventType").(string)
				//log.Trace("OBS event", "typeName", typeName)

			case 7: // RequestResponse
				//rtype := msg.get("requestType").(string)
				//log.Tracef("RequestResponse: %s: %v", rtype, msg)

				rid := msg.get("requestId").(string)
				responseData := msg.get("responseData")
				c, ok := gOBSRequests[rid]
				if ok {
					if responseData == nil {
						c <- nil
					} else {
						c <- responseData.(map[string]interface{})
					}
					delete(gOBSRequests, rid)
				}

			default:
				log.Info("Unhandled op: %d", msg.OpCode)
			}
		}
	}
}
