package routing

import (
	"fmt"
	"github.com/yamamoto-febc/jobq"
)

func Parallel(routes ...string) jobq.JobRouterFunc {
	return func(queue *jobq.Queue, option *jobq.Option, req jobq.JobRequestAPI) {
		for _, route := range routes {
			queue.PushRequest(route, req.GetPayload())
		}
	}
}

func PathThrough(dest string) jobq.JobRouterFunc {
	return func(queue *jobq.Queue, option *jobq.Option, req jobq.JobRequestAPI) {
		queue.PushRequest(dest, req.GetPayload())
	}
}

func Action(workerName string, f func(interface{}) jobq.JobAPI) jobq.JobRouterFunc {
	return func(queue *jobq.Queue, option *jobq.Option, req jobq.JobRequestAPI) {
		queue.PushJob(workerName, f(req.GetPayload()))
	}
}

var Goal jobq.JobRouterFunc = func(queue *jobq.Queue, option *jobq.Option, req jobq.JobRequestAPI) {
	queue.PushTrace(fmt.Sprintf("Route('%s') is finished.", req.GetName()))
}
