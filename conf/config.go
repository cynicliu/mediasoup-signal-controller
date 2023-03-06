package conf

type Tls_t struct {
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

type Https_t struct {
	ListenIp   string `json:"listenIp"`
	ListenPort int    `json:"listenPort"`
	Tls        Tls_t  `json:"tls"`
}

type WorkerSettings_t struct {
	LogLevel   string `json:"logLevel"`
	RtcMinPort int    `json:"rtcMinPort"`
	RtcMaxPort int    `json:"rtcMaxPort"`
}

type MediaCodecs_Param_t struct {
	X_google_start_bitrate  int    `json:"x-google-start-bitrate,omitempty"`
	Packetization_mode      int    `json:"packetization_mode,omitempty"`
	Profile_level_id        string `json:"profile-level-id"`
	Profile_id              int    `json:"profile-id,omitempty"`
	Level_asymmetry_allowed int    `json:"level_asymmetry_allowed,omitempty"`
}

type MediaCodec_t struct {
	Kind                 string              `json:"kind"`
	MimeType             string              `json:"mimeType"`
	ClockRate            int                 `json:"clockRate"`
	Channels             int                 `json:"channels"`
	PreferredPayloadType int                 `json:"preferredPayloadType"`
	Parameters           MediaCodecs_Param_t `json:"parameters"`
}

type ListenIp_t struct {
	Ip          string `json:"ip"`
	AnnouncedIp string `json:"announcedIp"`
}

type WebRtcTransportOptions_t struct {
	ListenIps                       []ListenIp_t `json:"listenIps"`
	InitialAvailableOutgoingBitrate int          `json:"initialAvailableOutgoingBitrate"`
	MinimumAvailableOutgoingBitrate int          `json:"minimumAvailableOutgoingBitrate"`
	MaxSctpMessageSize              int          `json:"maxSctpMessageSize"`
	MaxIncomingBitrate              int          `json:"maxIncomingBitrate"`
}

type PlainTransportOptions_t struct {
	ListenIp           ListenIp_t `json:"listenIp"`
	MaxSctpMessageSize int        `json:"maxSctpMessageSize"`
}

type RouterOptions_t struct {
	MediaCodecs            []MediaCodec_t           `json:"mediaCodecs"`
	WebRtcTransportOptions WebRtcTransportOptions_t `json:"webRtcTransportOptions"`
}
type MediaSoup_t struct {
	NumWorkers int    `json:"numWorkers"`
	WorkerPath string `json:"workerPath"`
	UnixPath   string `json:"unixPath"`

	WorkerSettings WorkerSettings_t `json:"workerSettings"`
	RouterOptions  RouterOptions_t  `json:"routerOptions"`

	WebRtcTransportOptions WebRtcTransportOptions_t `json:"webRtcTransportOptions"`
	PlainTransportOptions  PlainTransportOptions_t  `json:"plainTransportOptions"`
}

type Config struct {
	Domain    string      `json:"domain"`
	Https     Https_t     `json:"https"`
	Mediasoup MediaSoup_t `json:"mediasoup"`
}
