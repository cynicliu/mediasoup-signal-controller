package service

import (
	"fmt"
	"mediasoup-signal-controller/common"
	"mediasoup-signal-controller/conf"
	"os"
	"os/exec"
	"sync"

	"github.com/cloudwebrtc/go-protoo/logger"
	"github.com/go-basic/uuid"
)

type Worker struct {
	Waiting_Response
	Pid            int
	seq            int
	cf             *conf.Config
	channel        *Channel
	payloadChannel *PayloadChannel
	closed         bool
	routers        []*Router

	cmd    *exec.Cmd
	server *Server
}

func CreateNewWorker(server *Server, workerBin string, logLevel string, rtcMinPort int, rtcMaxPort int, cert string, key string, seq int) *Worker {
	worker := &Worker{
		seq:            seq,
		server:         server,
		cmd:            nil,
		channel:        nil,
		payloadChannel: nil,
		Waiting_Response: Waiting_Response{
			waitId: -1,
		},
	}

	worker.cf = server.Conf
	channelP := fmt.Sprintf("%s/channelProducer_%d", worker.cf.Mediasoup.UnixPath, seq)
	channelC := fmt.Sprintf("%s/channelConsumer_%d", worker.cf.Mediasoup.UnixPath, seq)
	channelPP := fmt.Sprintf("%s/channelPayloadProducer_%d", worker.cf.Mediasoup.UnixPath, seq)
	channelPC := fmt.Sprintf("%s/channelPayloadConsumer_%d", worker.cf.Mediasoup.UnixPath, seq)
	worker.channel = CreateNewChannel(channelP, channelC)
	worker.channel.SetListener(worker)
	worker.payloadChannel = CreateNewPayloadChannel(channelPP, channelPC)
	worker.payloadChannel.SetListener(worker)

	worker.channel.Start()
	worker.payloadChannel.Start()

	parameters := make([]string, 0)
	parameters = append(parameters, fmt.Sprintf("--logLevel=%s", logLevel))
	parameters = append(parameters, fmt.Sprintf("--rtcMinPort=%d", rtcMinPort))
	parameters = append(parameters, fmt.Sprintf("--rtcMaxPort=%d", rtcMaxPort))
	parameters = append(parameters, fmt.Sprintf("--dtlsCertificateFile=%s", cert))
	parameters = append(parameters, fmt.Sprintf("--dtlsPrivateKeyFile=%s", key))
	parameters = append(parameters, fmt.Sprintf("--seq=%d", seq))
	parameters = append(parameters, fmt.Sprintf("--channelProducer=%s", channelP))
	parameters = append(parameters, fmt.Sprintf("--channelConsumer=%s", channelC))
	parameters = append(parameters, fmt.Sprintf("--channelPayloadProducer=%s", channelPP))
	parameters = append(parameters, fmt.Sprintf("--channelPayloadConsumer=%s", channelPC))

	logger.Infof("start worker %s", workerBin)

	worker.cmd = exec.Command(workerBin, parameters...)
	logger.Infof("exec args: %v", worker.cmd.Args)
	//pp, err := worker._cmd.CombinedOutput()

	logger.Infof("MEDIASOUP_VERSION:%s", os.Getenv("MEDIASOUP_VERSION"))

	err := worker.cmd.Start()

	if err != nil {
		logger.Errorf("Mediasoup worker start failed:%s", err.Error())
		return nil
	}
	worker.Pid = worker.cmd.Process.Pid
	logger.Debugf("process id:%d", worker.Pid)
	go worker.ExitMonitor()
	return worker
}

func (worker *Worker) SetListener(server *Server) {
	worker.server = server
}

func (worker *Worker) CreateRouter(rom *Room) *Router {

	internal := &common.Internal_t{
		RouterId: uuid.New(),
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		worker.channel.Request("worker.createRouter", internal, nil, nil, nil, nil)
		wg.Done()
	}()

	router := CreateNewRouter(rom, worker.cf.Mediasoup.RouterOptions, internal, nil, worker.channel, worker.payloadChannel, nil)
	worker.routers = append(worker.routers, router)

	router.channel.AddListener(router.internal.RouterId, router)
	return router
}

func (worker *Worker) OnRouterClose(router *Router) {
	// todo
}

func (worker *Worker) HandleMessage(msg common.ChannelMessage, channelType string) {

	if len(msg.Event) > 0 && msg.Event == "running" {
		logger.Debugf("worker process running [pid:%s], %s, %p", msg.TargetId, channelType, worker)
	}
}

func (worker *Worker) ExitMonitor() {

	worker.cmd.Wait()
	logger.Errorf("worker %d exited", worker.Pid)
	worker.server.OnWorkerExit(worker.Pid, worker.seq)
}

func (worker *Worker) OnChannelMessage(msg common.ChannelMessage) {
	logger.Debugf("worker receive accepted message:id=%d, method:%t", msg.Id, msg.Accepted)

	if msg.Id == worker.waitId {
		worker.waitId = -1
	}
}
