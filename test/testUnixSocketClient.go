package main

import (
	"mediasoup-signal-controller/common"
	"time"

	"github.com/cloudwebrtc/go-protoo/logger"
)

func main() {

	usc := common.NewUnixSocketClient("/export/webrtc/channelProducer")
	err := usc.Connect()
	if err != nil {
		logger.Errorf("Connect UnixServer failed: %s", err.Error())
	}

	var data string
	data = "fdafdafadsf"
	len, err := usc.Handler.Conn.Write([]byte(data))
	if err != nil {
		logger.Errorf("write failed:%", err.Error())
	} else {
		logger.Infof("send data:%d", len)
	}

	//usc.Handler.Conn.Close()

	time.Sleep(time.Duration(10) * time.Second)
}
