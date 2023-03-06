package service

import "mediasoup-signal-controller/common"

type Waiting interface {
	OnChannelMessage(msg common.ChannelMessage)
}
type Waiting_Response struct {
	waitId int
}
