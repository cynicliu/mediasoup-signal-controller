package service

import (
	"encoding/json"
	"mediasoup-signal-controller/common"
	"net"
	"time"

	"github.com/cloudwebrtc/go-protoo/logger"
)

type PayloadChannel struct {
	common.UnixConnListener
	common.UnixDataListener
	puss         *common.UnixSocketServer
	cuss         *common.UnixSocketServer
	producerPath string
	consumerPath string

	sents    map[int]*common.SendMessage
	listener *Worker
}

type PayloadChannelHandler struct {
	common.UnixSocketHandler
}

func newPayloadChannelHandler(conn *net.UnixConn, chn *PayloadChannel) *PayloadChannelHandler {
	cnh := &PayloadChannelHandler{
		UnixSocketHandler: common.UnixSocketHandler{
			Conn:       conn,
			UdListener: chn,
			Running:    true,
		},
	}

	return cnh
}

func CreateNewPayloadChannel(producerPath string, consumerPath string) *PayloadChannel {
	channel := &PayloadChannel{
		producerPath: producerPath,
		consumerPath: consumerPath,
		listener:     nil,
	}

	channel.puss = common.NewUnixSocketServer(producerPath, channel)
	channel.cuss = common.NewUnixSocketServer(consumerPath, channel)

	channel.sents = make(map[int]*common.SendMessage)

	return channel
}

func (chn *PayloadChannel) SetListener(worker *Worker) { chn.listener = worker }

func (chn *PayloadChannel) Start() {
	chn.puss.StartServer()
	chn.cuss.StartServer()
}
func (chn *PayloadChannel) Remove(cnh *PayloadChannelHandler) {
	logger.Infof("Before remove handler size:%d, %d", chn.puss.Handlers.Len(), chn.cuss.Handlers.Len())

	chn.puss.Handlers.Remove(cnh.Pos)
	chn.cuss.Handlers.Remove(cnh.Pos)

	logger.Infof("Afer Remove handler size:%d, %d", chn.puss.Handlers.Len(), chn.cuss.Handlers.Len())
}

func (chn *PayloadChannel) Stop(value interface{}) {
	handler := value.(*PayloadChannelHandler)
	handler.Conn.Close()
}

func (chn *PayloadChannel) Send(value interface{}, data []byte) (int, error) {
	handler := value.(*PayloadChannelHandler)
	return handler.Conn.Write(data)
}

func (chn *PayloadChannel) HandleUnixConn(c *net.UnixConn, uss *common.UnixSocketServer) {
	logger.Errorf("Channel.HandleUnixConn: %p", c)
	handler := newPayloadChannelHandler(c, chn)
	handler.Pos = uss.Handlers.PushBack(handler)

	//c.SetReadDeadline(time.Now().Add(time.Millisecond * 100))
	go handler.Loop()
}

func (cnh *PayloadChannelHandler) Loop() {

	if cnh.Conn == nil {
		logger.Errorf("start a nil conn")
		return
	}

	cnh.Conn.SetReadDeadline(time.Now().Add(common.PongWait))

	for {
		if !cnh.Running {
			cnh.Conn.Close()
			cnh.Conn = nil
			return
		}

		buf := make([]byte, 2048)
		nlen, err := cnh.Conn.Read(buf)

		if err != nil {

			nerr, ok := err.(net.Error)
			if ok && nerr.Timeout() {
				//logger.Infof("payload channel continue")
				continue
			}

			logger.Errorf("socket errr: %s,%d", err.Error(), ok)

			cnh.UdListener.(*PayloadChannel).Remove(cnh)
			return
		} else if nlen == 0 {
			logger.Errorf("socket closed")

			cnh.UdListener.(*PayloadChannel).Remove(cnh)
			return
		}

		//data handle
		logger.Infof("Receive data:%d, %s", nlen, buf)
		data := append(cnh.Buffer, buf...)
		rlen := cnh.UdListener.RecvData(data, nlen)

		cnh.Buffer = data[(nlen - rlen):nlen]

		logger.Infof("data len:%d", len(cnh.Buffer))
		time.Sleep(time.Duration(10) * time.Second)
	}
}

func (chn *PayloadChannel) RecvData(buffer []byte, nsize int) int {
	logger.Infof("Receive data len: %d", nsize)
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

				logger.Infof("%d, str:%s, pos:%d", nlen, string(payload), pos)
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
		case 88:
		default:
			logger.Infof("RecvData end -1")
			return 0
		}
	}
}

func (chn *PayloadChannel) Close() {

}

func (chn *PayloadChannel) processMessage(msg common.ChannelMessage) {
	if msg.Id > 0 {
		sent := chn.sents[msg.Id]

		if sent == nil {
			logger.Errorf("received response does not match any sent request [id:%s]", msg.Id)
			return
		}

		if msg.Accepted && sent.Method != "method:dataProducer.getStats" && sent.Method != "method:transport.getStats" {
			logger.Debugf("request succeeded [method:%s, id:%s]", sent.Method, sent.Id)
			delete(chn.sents, msg.Id)
			return
		} else if len(msg.ErrorInfo) > 0 {
			logger.Warnf("request failed [method:%s, id:%s]: %s", sent.Method, sent.Id, msg.Reason)

			if msg.ErrorInfo == "TypeEror" {
				return
			} else {
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
		logger.Debugf("targetID:%s,event:%s", msg.TargetId, msg.Event)
		chn.listener.HandleMessage(msg, "PayloadChannel")
	} else {
		logger.Errorf("received message is not a response nor a notification")
	}
}
