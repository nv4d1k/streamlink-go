package BiliBili

type StreamInfo struct {
	ProtocolName StreamPrototol
	FormatName   StreamFormat
	CodecName    StreamCodec
	CurrentQn    int
	AcceptQn     []int
	BaseURL      string
	URLHost      string
	URLExtra     string
}

type StreamPrototol string
type StreamFormat string
type StreamCodec string

func (p StreamPrototol) IntString() string {
	switch p {
	case "http_stream":
		return "0"
	case "http_hls":
		return "1"
	default:
		return "-1"
	}
}

func (f StreamFormat) IntString() string {
	switch f {
	case "flv":
		return "0"
	case "ts":
		return "1"
	case "fmp4":
		return "2"
	default:
		return "-1"
	}
}

func (c StreamCodec) IntString() string {
	switch c {
	case "avc":
		return "0"
	case "hevc":
		return "1"
	default:
		return "-1"
	}
}
