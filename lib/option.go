package lib

import (
	"fmt"
	"github.com/yamamoto-febc/jobq"
)

var SakuraCloudDefaultZones = []string{"tk1v", "is1a", "is1b", "tk1a"}

type Option struct {
	AccessToken       string
	AccessTokenSecret string
	Zones             []string
	TraceMode         bool
	ForceMode         bool
	JobQueueOption    *jobq.Option
}

func NewOption() *Option {
	return &Option{
		Zones:          SakuraCloudDefaultZones,
		JobQueueOption: jobq.NewOption(),
	}
}

func (o *Option) Validate() []error {
	var errors []error
	if o.AccessToken == "" {
		errors = append(errors, fmt.Errorf("[%s] is required", "token"))
	}
	if o.AccessTokenSecret == "" {
		errors = append(errors, fmt.Errorf("[%s] is required", "secret"))
	}

	return errors
}
