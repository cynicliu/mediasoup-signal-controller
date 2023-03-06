package main

import (
	"fmt"
	"mediasoup-signal-controller/common"
	"mediasoup-signal-controller/service"
	"time"

	"github.com/cloudwebrtc/go-protoo/logger"
)

type Test struct {
	common.UnixConnListener
	common.UnixDataListener
}

func newTest() *Test {

	test := &Test{}

	return test
}

func (test *Test) RecvData(buffer []byte, len int) {

}

func main() {

	// test := newTest()

	// uss := common.NewUnixSocketServer("./test.unix", &test.UnixConnListener)
	// err := uss.StartServer()
	// if err != nil {
	// 	logger.Errorf("Start UnixServer failed: %s", err.Error())
	// }

	// usc := common.NewUnixSocketClient("./test.unix")
	// err = usc.Connect()
	// if err != nil {
	// 	logger.Errorf("Connect UnixServer failed: %s", err.Error())
	// }

	datad := [...]byte{1, 2, 3, 4, 5, 6, 7}
	re := datad[5:7]

	fmt.Println(re)

	chn := service.CreateNewChannel("producer", "consumer")
	chn.Start()

	usc := common.NewUnixSocketClient("producer")
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

	usc.Handler.Conn.Close()

	time.Sleep(time.Duration(10) * time.Second)
}
