package jobq

import "fmt"

// Queue ジョブキュー
type Queue struct {
	Request chan JobRequestAPI
	Job     map[string]chan JobAPI
	Logs    *LogQueue
	Quit    chan error
}

// LogQueue ログ出力用キュー
type LogQueue struct {
	Info  chan string
	Trace chan string
	Warn  chan error
	Error chan error
}

var defaultLogBufferSize = 10

// NewQueue ジョブキューの新規作成
func NewQueue(workBufSize int) *Queue {
	return &Queue{
		Request: make(chan JobRequestAPI, workBufSize),
		Job:     map[string]chan JobAPI{},
		Logs: &LogQueue{
			Info:  make(chan string, defaultLogBufferSize),
			Trace: make(chan string, defaultLogBufferSize),
			Warn:  make(chan error, defaultLogBufferSize),
			Error: make(chan error, defaultLogBufferSize),
		},
		Quit: make(chan error),
	}
}

func (q *Queue) CreateJobQueue(name string, bufSize int) error {
	if _, ok := q.Job[name]; ok {
		return fmt.Errorf("JobQueue[%s] is already exists.", name)
	}
	q.Job[name] = make(chan JobAPI, bufSize)
	return nil
}

func (q *Queue) RemoveJobQueue(name string) error {
	if _, ok := q.Job[name]; ok {
		delete(q.Job, name)
	}
	return fmt.Errorf("JobQueue[%s] is not exists.", name)
}

func (q *Queue) ClearJobQueue() {
	q.Job = map[string]chan JobAPI{}
}

// PushRequest push new request to job-routing queue
func (q *Queue) PushRequest(requestName string, payload interface{}) {
	q.Request <- &jobRequest{
		name:    requestName,
		payload: payload,
	}
}

//---------------------------------------------------------
// Push jobs
//---------------------------------------------------------

// PushJob push job to job queue
func (q *Queue) PushJob(queueName string, work JobAPI) {
	q.Job[queueName] <- work
}

//---------------------------------------------------------
// Stop
//---------------------------------------------------------

// Stop push stop request to queue
func (q *Queue) Stop() {
	q.Quit <- nil
}

// StopByError push stop request wth error to queue
func (q *Queue) StopByError(err error) {
	q.Quit <- err
}

//---------------------------------------------------------
// Logging functions
//---------------------------------------------------------

// PushTrace push message to trace-log queue
func (q *Queue) PushTrace(msg string) {
	q.Logs.Trace <- msg
}

// PushInfo push message to info-log queue
func (q *Queue) PushInfo(msg string) {
	q.Logs.Info <- msg
}

// PushWarn push message to warn-log queue
func (q *Queue) PushWarn(err error) {
	q.Logs.Warn <- err
}

// PushError push error to error queue
func (q *Queue) PushError(err error) {
	q.Logs.Error <- err
}
