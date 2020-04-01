package transcode

type OpusFrame struct {
	Seq     uint64
	Pos     uint64 // Granule position as number of samples at 48Khz sample rate.
	Samples uint16 // Number of 48Khz samples.
	Data    []byte // Opus encoded data.
}
