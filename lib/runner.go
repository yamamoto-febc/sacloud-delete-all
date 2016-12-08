package lib

import (
	"fmt"
	"github.com/yamamoto-febc/jobq"
	"sync"
)

// Run メイン処理
func Run(option *Option) error {

	currentOption = option
	resourceWaitGroup = sync.WaitGroup{}
	resourceWaitGroup.Add(19) // all resource

	// setup jobs environments
	jobQueue := jobq.NewJobQueue(option.JobQueueOption, routes)
	jobQueue.AddDefaultActionWorker("sacloud", 10)

	if !option.ForceMode {
		input := ""
		fmt.Print("Do you really want to destroy all?[Y/n]")
		fmt.Scanln(&input)
		if input != "Y" {
			return nil
		}
	}

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
var resourceWaitGroup sync.WaitGroup
