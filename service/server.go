package service

import (
	"mediasoup-signal-controller/conf"
	"os"

	"github.com/cloudwebrtc/go-protoo/logger"
)

type Server struct {
	Conf                   *conf.Config
	Rooms                  map[string]*Room
	workers                map[int]*Worker
	nextMediasoupWorkerIdx int
}

func CreateNewServer(cf *conf.Config) *Server {
	server := &Server{
		Conf:                   cf,
		nextMediasoupWorkerIdx: 0,
	}

	server.Rooms = make(map[string]*Room)
	server.workers = make(map[int]*Worker, 0)

	return server
}

func (svr *Server) RunMediasoupWorkers() error {

	numWorkers := svr.Conf.Mediasoup.NumWorkers

	os.Setenv("MEDIASOUP_VERSION", "3.6.32")
	for i := 0; i < numWorkers; i++ {

		worker := CreateNewWorker(svr, svr.Conf.Mediasoup.WorkerPath, svr.Conf.Mediasoup.WorkerSettings.LogLevel,
			svr.Conf.Mediasoup.WorkerSettings.RtcMinPort, svr.Conf.Mediasoup.WorkerSettings.RtcMaxPort,
			svr.Conf.Https.Tls.Cert, svr.Conf.Https.Tls.Key, i)
		if worker != nil {
			logger.Infof("CreateWorker success")
			svr.workers[worker.Pid] = worker
			worker.SetListener(svr)
		} else {
			logger.Errorf("Start worker %d failed", i)
		}
	}

	return nil
}

func (svr *Server) GetOrCreateRoom(roomId string) *Room {

	if roomId != "" && svr.Rooms[roomId] != nil {
		return svr.Rooms[roomId]
	}

	worker := svr.getMediasoupWorker()
	room := CreateNewRoom(svr.Conf, worker, roomId)
	svr.Rooms[roomId] = room
	return room
}

func (svr *Server) getMediasoupWorker() *Worker {

	var worker *Worker
	worker = nil

	i := 0
	for _, val := range svr.workers {
		if i == svr.nextMediasoupWorkerIdx {
			worker = val
		}
		i++
	}

	svr.nextMediasoupWorkerIdx++
	if svr.nextMediasoupWorkerIdx == len(svr.workers) {
		svr.nextMediasoupWorkerIdx = 0
	}

	return worker
}

func (svr *Server) OnWorkerExit(pid int, seq int) {
	svr.workers[pid] = nil
	// restart a woker
	logger.Infof("Restart worker after worker seq:%d = %d exit", seq, pid)
	worker := CreateNewWorker(svr, svr.Conf.Mediasoup.WorkerPath, svr.Conf.Mediasoup.WorkerSettings.LogLevel,
		svr.Conf.Mediasoup.WorkerSettings.RtcMinPort, svr.Conf.Mediasoup.WorkerSettings.RtcMaxPort,
		svr.Conf.Https.Tls.Cert, svr.Conf.Https.Tls.Key, seq)
	if worker != nil {
		logger.Infof("CreateWorker success")
		svr.workers[worker.Pid] = worker
		worker.SetListener(svr)
	} else {
		logger.Errorf("Restart worker failed")
	}
}

func (svr *Server) OnRoomClose(roomId string) {
	rom := svr.Rooms[roomId]
	rom.Close()
	svr.Rooms[roomId] = nil
}
