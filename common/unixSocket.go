package common

import (
	"container/list"
	"net"
	"os"

	"github.com/cloudwebrtc/go-protoo/logger"
)

type UnixConnListener interface {
	HandleUnixConn(c *net.UnixConn, uss *UnixSocketServer)
	Stop(value interface{}) // value = UnixChannelHandler
	Send(value interface{}, data []byte) (int, error)
}

type UnixDataListener interface {
	RecvData(buffer []byte, len int) int
}

type UnixSocket struct {
	FileName string
}

type UnixSocketServer struct {
	UnixSocket
	listener     *net.UnixListener
	connListener UnixConnListener
	Handlers     *list.List //*UnixSocketHandler->ChannelHandler, PayloadHandler
}

type UnixSocketClient struct {
	UnixSocket
	UnixDataListener
	Handler *UnixSocketHandler
}

type UnixSocketHandlerBase interface {
	Send(buffer []byte) (int, error)
}
type UnixSocketHandler struct {
	Pos        *list.Element
	Conn       *net.UnixConn
	UdListener UnixDataListener
	Buffer     []byte
	Running    bool
}

func NewUnixSocketHandler(conn *net.UnixConn, listener UnixDataListener) *UnixSocketHandler {
	ush := &UnixSocketHandler{
		Conn:       conn,
		UdListener: listener,
		Running:    true,
	}

	return ush
}

func (ush *UnixSocketHandler) Loop() {

	if ush.Conn == nil {
		logger.Errorf("start a nil conn")
		return
	}
	for {
		if !ush.Running {
			ush.Conn.Close()
			ush.Conn = nil
			return
		}

		logger.Infof("Base Loop")

	}
}

func (ush *UnixSocketHandler) Send(buffer []byte) (n int, err error) {

	if !ush.Running {
		n = -1
		err = nil
		return
	}

	n, err = ush.Conn.Write(buffer)
	return
}

func (ush *UnixSocketHandler) Stop() {
	ush.Running = false
}

func NewUnixSocketServer(fileName string, connListener UnixConnListener) *UnixSocketServer {
	uss := &UnixSocketServer{
		UnixSocket: UnixSocket{
			FileName: fileName,
		},
		listener:     nil,
		connListener: connListener,
		Handlers:     list.New(),
	}

	return uss
}

func (uss *UnixSocketServer) createServer() error {
	os.Remove(uss.FileName)

	addr, err := net.ResolveUnixAddr("unix", uss.FileName)
	if err != nil {
		return err
	}

	uss.listener, err = net.ListenUnix("unix", addr)
	if err != nil {
		logger.Errorf("Unix socket listen failed:%s", err.Error())
		return err
	}

	logger.Infof("Unix socket listen on:%v", uss.listener.Addr())

	go func() {
		for {
			c, err := uss.listener.Accept()
			if err != nil {
				logger.Errorf("Unix socket accept exception: %s", err.Error())
				continue
			}
			logger.Infof("Accept socket :%p", c)

			uss.HandleUnixConn(c.(*net.UnixConn))
		}
	}()

	return nil
}

func (uss *UnixSocketServer) StartServer() error {
	return uss.createServer()
}

func (uss *UnixSocketServer) Stop() {

	for {
		if uss.Handlers.Len() <= 0 {
			break
		}

		el := uss.Handlers.Front()
		uss.connListener.Stop(el.Value)

		uss.Handlers.Remove(el)
	}
	if uss.listener != nil {
		uss.listener.Close()
	}
}

func (uss *UnixSocketServer) HandleUnixConn(c *net.UnixConn) {
	logger.Infof("handle socket :%p", c)
	uss.connListener.HandleUnixConn(c, uss)
}

func (uss *UnixSocketServer) Write(data []byte) (int, error) {
	// just sent to first client, because one worker one producer
	el := uss.Handlers.Front()
	logger.Debugf("Handlers len:%d", uss.Handlers.Len())
	return uss.connListener.Send(el.Value, data)
}

func NewUnixSocketClient(fileName string) *UnixSocketClient {
	usc := &UnixSocketClient{
		UnixSocket: UnixSocket{
			FileName: fileName,
		},
		Handler: nil,
	}

	return usc
}

func (usc *UnixSocketClient) Connect() error {
	addr, err := net.ResolveUnixAddr("unix", usc.FileName)
	if err != nil {
		logger.Errorf("Cannot resolve unix add: %s", err.Error())
	}

	conn, err := net.DialUnix("unix", nil, addr)
	if err != nil {
		logger.Errorf("Unix addr connect failed: %s, file:%s", err.Error(), usc.FileName)
		return err
	}

	logger.Infof("Client connect success")

	usc.Handler = NewUnixSocketHandler(conn, usc)
	go usc.Handler.Loop()

	return nil
}
