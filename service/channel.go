package service

import (
	"encoding/json"
	"mediasoup-signal-controller/common"
	"net"
	"time"

	"github.com/cloudwebrtc/go-protoo/logger"
)

type RespondFunc func(data interface{})
type AcceptFunc func(data common.ChannelMessage)
type RejectFunc func(errorCode int, errorReason string)

type Transcation struct {
	id     int
	accept AcceptFunc
	reject RejectFunc
	close  func()
	//resultChan chan ResultFuture
}

type ChannelSendMessage struct {
	common.SendMessage

	accept AcceptFunc
	reject RejectFunc

	channel *Channel
	router  *Router
}

func createChannelSendMessage(id int, method string, channel *Channel, router *Router, accept AcceptFunc, reject RejectFunc) *ChannelSendMessage {
	sent := &ChannelSendMessage{
		SendMessage: common.SendMessage{
			Id:     id,
			Method: method,
		},
		channel: channel,
		router:  router,
		accept:  accept,
		reject:  reject,
	}

	// add time out timer

	return sent
}
func (csm *ChannelSendMessage) OnTimeout() {
	delete(csm.channel.sents, csm.Id)
}

type Channel struct {
	common.UnixConnListener
	common.UnixDataListener
	puss         *common.UnixSocketServer // myself is a consumer
	cuss         *common.UnixSocketServer // remote is a consumer
	producerPath string
	consumerPath string

	sents            map[int]*ChannelSendMessage
	producerToRouter map[string]*Router
	tranportToRouter map[string]*Router
	routerIdToRouter map[string]*Router
	listener         *Worker

	listeners map[string]interface{}

	nextId int
}

type ChannelHandler struct {
	common.UnixSocketHandler

	path string
}

func newChannelHandler(conn *net.UnixConn, chn *Channel, path string) *ChannelHandler {
	cnh := &ChannelHandler{
		UnixSocketHandler: common.UnixSocketHandler{
			Conn:       conn,
			UdListener: chn,
			Running:    true,
		},
	}

	cnh.path = path

	return cnh
}

func CreateNewChannel(producerPath string, consumerPath string) *Channel {
	channel := &Channel{
		producerPath: producerPath,
		consumerPath: consumerPath,
		listener:     nil,
	}

	channel.puss = common.NewUnixSocketServer(producerPath, channel)
	channel.cuss = common.NewUnixSocketServer(consumerPath, channel)

	channel.sents = make(map[int]*ChannelSendMessage)
	channel.producerToRouter = make(map[string]*Router)
	channel.tranportToRouter = make(map[string]*Router)
	channel.routerIdToRouter = make(map[string]*Router)

	channel.listeners = make(map[string]interface{})

	return channel
}

func (chn *Channel) SetListener(worker *Worker) { chn.listener = worker }

func (chn *Channel) Start() {
	chn.puss.StartServer()
	chn.cuss.StartServer()
}
func (chn *Channel) Remove(cnh *ChannelHandler) {
	logger.Infof("Before remove handler size:%d, %d", chn.puss.Handlers.Len(), chn.cuss.Handlers.Len())

	chn.puss.Handlers.Remove(cnh.Pos)
	chn.cuss.Handlers.Remove(cnh.Pos)

	logger.Infof("Afer Remove handler size:%d, %d", chn.puss.Handlers.Len(), chn.cuss.Handlers.Len())
}

func (chn *Channel) Stop(value interface{}) {
	handler := value.(*ChannelHandler)
	handler.Stop()
}

func (chn *Channel) Send(value interface{}, data []byte) (int, error) {
	logger.Debugf("Channel Send: %s", string(data))
	handler := value.(*ChannelHandler)
	return handler.Conn.Write(data)
}

func (chn *Channel) HandleUnixConn(c *net.UnixConn, uss *common.UnixSocketServer) {
	logger.Errorf("Channel.HandleUnixConn: %s", uss.FileName)
	handler := newChannelHandler(c, chn, uss.FileName)
	handler.Pos = uss.Handlers.PushBack(handler)

	go handler.Loop()
}

func (cnh *ChannelHandler) Loop() {

	if cnh.Conn == nil {
		logger.Errorf("start a nil conn")
		return
	}

	cnh.Conn.SetReadDeadline(time.Now().Add(common.PongWait))
	for {
		if !cnh.Running {
			logger.Infof("----------------ChannelHandler exit")
			cnh.Conn.Close()
			cnh.Conn = nil
			return
		}

		buf := make([]byte, 2048)
		nlen, err := cnh.Conn.Read(buf)

		if nlen != 0 {
			//logger.Infof("channel:%d, %s", nlen, cnh.path)
		}

		if err != nil {
			if nerr, ok := err.(net.Error); ok && nerr.Timeout() {
				//logger.Infof("channel continue")
				continue
			}

			logger.Errorf("socket errr: %s", err.Error())

			cnh.UdListener.(*Channel).Remove(cnh)
			return
		} else if nlen == 0 {
			logger.Errorf("xxxxxsocket closed")

			cnh.UdListener.(*Channel).Remove(cnh)
			return
		}

		//data handle
		//logger.Infof("=============Receive data:%d, %s", nlen, buf)
		data := append(cnh.Buffer, buf...)
		rlen := cnh.UdListener.RecvData(data, nlen)

		cnh.Buffer = data[rlen:nlen]

		//logger.Infof("data len:%d", len(cnh.Buffer))
		//time.Sleep(time.Duration(10) * time.Second)
	}
}

func (cnh *ChannelHandler) Stop() {
	cnh.Running = false
	cnh.Conn.Close()
}

func (chn *Channel) RecvData(buffer []byte, nsize int) int {
	//logger.Infof("Receive data len: %d", nsize)
	pos := 0
	for {
		payload, nlen := common.NsPayload(buffer, pos)

		switch payload[0] {
		case 123: // 123 = {'s ascii
			if payload != nil {
				var cm common.ChannelMessage
				err := json.Unmarshal(payload, &cm)
				if err != nil {
					logger.Errorf("%s", err.Error())
				}
				chn.processMessage(cm)

				pos += common.NsWriteLength(nlen)

				//logger.Infof("%d, str:%s, pos:%d", nlen, string(payload), pos)
				if pos >= nsize {
					return pos
				}
			} else if nlen == -1 {
				logger.Infof("RecvData end -1")
				return 0
			} else {
				logger.Infof("RecvData end:%d", nsize-pos)
				return nsize - pos
			}
			break
		case 68:
		case 87:
		case 69:
			logger.Errorf("%s", string(payload))
			return nsize
		case 88:
		default:
			logger.Infof("RecvData end -1")
			return 0
		}
	}
}

func (chn *Channel) Request(method string, internal interface{}, reqData interface{}, router *Router, accept AcceptFunc, reject RejectFunc) (int, error) {

	if chn.nextId < 4294967295 {
		chn.nextId++
	} else {
		chn.nextId = 1
	}

	request := common.Request_t{
		Id:       chn.nextId,
		Method:   method,
		Internal: internal,
		Data:     reqData,
	}

	data, _ := json.Marshal(request)
	//logger.Debugf("Channel Request:len:%d, data:%s", len(data), string(data))

	ns, _ := common.NsWrite(data, 0, len(data))
	//logger.Debugf("Channel Request nlen:%d,nstring:%s to %s", slen, string(ns), chn.cuss.FileName)
	sent, err := chn.cuss.Write(ns)

	if err != nil {
		logger.Errorf("Channel %p sent failed: %s", chn, err.Error())
		return -1, err
	}

	logger.Debugf("Channel Request sent: %d", sent)

	chn.sents[request.Id] = createChannelSendMessage(request.Id, method, chn, router, accept, reject)

	return request.Id, err
}

func (chn *Channel) AddRouter(id string, router *Router) {
	chn.routerIdToRouter[id] = router
}

func (chn *Channel) RemoveRouter(id string) {
	chn.routerIdToRouter[id] = nil
}

func (chn *Channel) AddProducer(id string, router *Router) {
	logger.Debugf("=============add producer:%s==========", id)
	chn.producerToRouter[id] = router
}

func (chn *Channel) RemoveProducer(id string) {
	chn.producerToRouter[id] = nil
}

func (chn *Channel) AddTransport(id string, router *Router) {
	chn.tranportToRouter[id] = router
}

func (chn *Channel) RemoveTransport(id string) {
	chn.tranportToRouter[id] = nil
}

func (chn *Channel) AddListener(id string, listener interface{}) {
	chn.listeners[id] = listener
}

func (chn *Channel) RemoveListener(id string) {
	delete(chn.listeners, id)
}

func (chn *Channel) Close() {

}

// else if len(msg.Event) > 0 && msg.Event == "score" {
// 	var scoreDatas []common.ScoreData
// 	json.Unmarshal(msg.Data, &scoreDatas)

// 	if chn.routerIdToRouter[msg.TargetId] != nil {
// 		logger.Debugf("====================router score:%+v", scoreDatas)
// 		//chn.routerIdToRouter[msg.TargetId].rom.NotifyAll("producerScore", msg.TargetId, msg.Data)
// 	} else if chn.producerToRouter[msg.TargetId] != nil {
// 		//logger.Debugf("====================producer score:%+v", scoreDatas)
// 		//chn.producerToRouter[msg.TargetId].Notify(chn.producerToRouter[msg.TargetId].rom, msg.TargetId, msg.Data)
// 	} else if chn.tranportToRouter[msg.TargetId] != nil {
// 		logger.Debugf("====================transport score:%+v", scoreDatas)
// 		//chn.tranportToRouter[msg.TargetId].rom.Notify()
// 	}
// 	//} else if len(msg.Event) > 0 && msg.Event == "producerpause" {
// }
func (chn *Channel) processMessage(msg common.ChannelMessage) {
	//logger.Infof("enter channel.processMessage")
	if msg.Id > 0 {
		sent := chn.sents[msg.Id]

		if sent == nil {
			logger.Errorf("received response does not match any sent request [id:%d]", msg.Id)
			return
		}

		if msg.Accepted && sent.Method != "method:dataProducer.getStats" && sent.Method != "method:transport.getStats" {
			logger.Debugf("request succeeded [method:%s, id:%d]", sent.Method, sent.Id)

			if chn.sents[msg.Id].accept != nil {
				chn.sents[msg.Id].accept(msg)
			}

			delete(chn.sents, msg.Id)
			return
		} else if len(msg.ErrorInfo) > 0 {
			logger.Warnf("request failed [method:%s, id:%d]: %s", sent.Method, sent.Id, msg.Reason)

			if chn.sents[msg.Id].reject != nil {
				chn.sents[msg.Id].reject(400, msg.ErrorInfo)
			}

			delete(chn.sents, msg.Id)
		}
	} else if len(msg.Event) > 0 {
		logger.Debugf("targetID:%s,event:%s,data:%s", msg.TargetId, msg.Event, string(msg.Data))
		object := chn.listeners[msg.TargetId]
		if object != nil {
			consumer, ok := object.(*Consumer)
			if ok {
				consumer.HandleNotification(msg.TargetId, msg)
				return
			}

			producer, ok := object.(*Producer)
			if ok {
				producer.HandleNotification(msg.TargetId, msg)
				return
			}

			router, ok := object.(*Router)
			if ok {
				router.HandleNotification(msg.TargetId, msg)
				return
			}
		}
	} else if len(msg.TargetId) > 0 && len(msg.Event) > 0 {
		// Due to how Promises work, it may happen that we receive a response
		// from the worker followed by a notification from the worker. If we
		// emit the notification immediately it may reach its target **before**
		// the response, destroying the ordered delivery. So we must wait a bit
		// here.
		// See https://github.com/versatica/mediasoup/issues/510
		//setImmediate(() => this.emit(msg.targetId, msg.event, msg.data));
		logger.Debugf("targetID:%s,event:%s, %s", msg.TargetId, msg.Event, string(msg.Data))
		chn.listener.HandleMessage(msg, "Channel")
	} else {
		logger.Errorf("received message is not a response nor a notification")
	}
}

func (chn *Channel) OnTimeout() {

}
