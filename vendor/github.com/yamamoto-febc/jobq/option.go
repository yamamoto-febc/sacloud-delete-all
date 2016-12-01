package jobq

type Option struct {
	TraceLog         bool
	InfoLog          bool
	WarnLog          bool
	ErrorLog         bool
	RequestQueueSize int
}

const defaultRequestQueueSize = 10

func NewOption() *Option {
	return &Option{
		RequestQueueSize: defaultLogBufferSize,
	}
}

func (o *Option) Validate() []error {
	return []error{}
}
