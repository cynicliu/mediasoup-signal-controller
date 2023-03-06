package main

import (
	"encoding/json"
	"flag"
	"log"
	"mediasoup-signal-controller/conf"
	"mediasoup-signal-controller/service"
	"net/http"
	"os"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/cloudwebrtc/go-protoo/peer"
	"github.com/cloudwebrtc/go-protoo/server"
	"github.com/cloudwebrtc/go-protoo/transport"
)

var (
	ConfigPath       = flag.String("-c", "./conf/config.json", "config file location")
	g_server         *service.Server
	g_channel        *service.Channel
	g_payloadChannel *service.PayloadChannel
)

func handleProtooWebSocket(transport *transport.WebSocketTransport, request *http.Request) {

	vars := request.URL.Query()
	peerId := vars["peerId"][0]
	roomId := vars["roomId"][0]

	if len(peerId) == 0 || len(roomId) == 0 {
		logger.Errorf("Connection request without roomId and/or peerId")
		return
	}

	room := g_server.GetOrCreateRoom(roomId)
	if room == nil {
		logger.Errorf("room create faild from Room:%s", roomId)
		return
	}

	pr := room.CreatePeer(peerId, transport)
	if pr == nil {
		logger.Errorf("peer create faild from Room:%s", roomId)
		return
	}

	handleProtooRequest := func(request peer.Request, accept peer.RespondFunc, reject peer.RejectFunc) {
		logger.Infof("handleProtooRequest => %s", request.Method)

		room.HandleProtooRequest(pr, request, accept, reject)
	}

	handleNotification := func(notification map[string]interface{}) {
		logger.Infof("handleNotification => %s", notification["method"])

		method := notification["method"].(string)
		data := notification["data"].(map[string]interface{})

		//Forward notification to testRoom.
		room.Notify(pr, method, data)
	}

	handleClose := func(code int, err string) {
		logger.Infof("handleClose => peer (%s) [%d] %s", pr.ID(), code, err)

		room.HandleClose(pr, code, err)
	}

	_, _, _ = handleProtooRequest, handleNotification, handleClose

	for {
		select {
		case msg := <-pr.OnNotification:
			log.Println("OnNotification msg", msg)
			// handleNotification
		case msg := <-pr.OnRequest:
			handleProtooRequest(msg.Request, msg.Accept, msg.Reject)
			// log.Println(msg)
		case msg := <-pr.OnClose:
			handleClose(msg.Code, msg.Text)
		}
	}
}

func main() {
	protooServer := server.NewWebSocketServer(handleProtooWebSocket)

	file, _ := os.Open(*ConfigPath)
	defer file.Close()

	decoder := json.NewDecoder(file)

	var g_config conf.Config
	err := decoder.Decode(&g_config)

	if err != nil {
		panic(err)
	}

	logger.Infof("Handle socket:%s", g_config.Domain)

	//logger.Infof("%v", g_config)
	g_server = service.CreateNewServer(&g_config)

	err = g_server.RunMediasoupWorkers()
	if err != nil {
		logger.Errorf("RunMediasoupWorkers:%s", err.Error())
		panic(err)
	}

	config := server.DefaultConfig()
	config.Port = 4443
	config.CertFile = "./certs/cert.pem"
	config.KeyFile = "./certs/key.pem"
	protooServer.Bind(config)
}
