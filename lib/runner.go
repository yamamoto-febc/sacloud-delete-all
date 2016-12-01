package lib

import (
	"fmt"
	"github.com/yamamoto-febc/jobq"
	"sync"
)

// Run メイン処理
func Run(option *Option) error {

	currentOption = option
	wg = sync.WaitGroup{}
	wg.Add(17) // all resource

	// setup jobs environments
	jobQueue := jobq.NewJobQueue(option.JobQueueOption, routes)
	jobQueue.AddDefaultActionWorker("sacloud", 10)

	fmt.Println("Start.")

	// start jobs
	err := jobQueue.StartDispatch()

	if err != nil {
		return err
	}

	fmt.Println("Done.")

	return nil
}

var currentOption *Option
var wg sync.WaitGroup
