package transcode

type Transcode interface {
	Write(packet []byte) error

	ReadFrame() ([]int16, error)
}
