package jobq

import (
	"fmt"
	"log"
	"os"
	"time"
)

// JobQueue ジョブキューに対するディスパッチ処理を担当する
type JobQueue struct {
	option          *Option
	queue           *Queue
	workers         map[string]func(chan JobAPI)
	jobRoutingTable map[string]JobRouterFunc
}

// NewJobQueue Dispatcherの新規作成
func NewJobQueue(option *Option, routingTable map[string]JobRouterFunc) *JobQueue {
	return &JobQueue{
		option:          option,
		queue:           NewQueue(option.RequestQueueSize),
		workers:         map[string]func(chan JobAPI){},
		jobRoutingTable: routingTable,
	}
}

// AddWorker ジョブキューをディスパッチするワーカーの追加
func (d *JobQueue) AddWorker(workerName string, queueSize int, action func(chan JobAPI)) error {
	if _, ok := d.workers[workerName]; ok {
		return fmt.Errorf("JobDispatch worker[%s] is already exists.", workerName)
	}
	d.workers[workerName] = action
	d.queue.CreateJobQueue(workerName, queueSize)
	return nil
}

// RemoveWorker ジョブキューをディスパッチするワーカーの削除
func (d *JobQueue) RemoveWorker(workerName string) error {
	if _, ok := d.workers[workerName]; ok {
		delete(d.workers, workerName)
		return d.queue.RemoveJobQueue(workerName)

	}
	return fmt.Errorf("JobDispatch worker[%s] is not exists.", workerName)
}

// ClearWorkers ジョブキューをディスパッチするワーカーをクリア
func (d *JobQueue) ClearWorkers() {
	d.workers = map[string]func(chan JobAPI){}
	d.queue.ClearJobQueue()
}

// Dispatch ジョブキューに対する各種ワーカーの登録やルーティング処理の登録、ディスパッチを行う
func (d *JobQueue) StartDispatch() error {

	// エラー時処理 登録
	d.dispatchMessageAction()

	// ジョブのルーティング 登録
	d.dispatchJobRequests()

	// API呼び出し 登録
	d.dispatchJobs()

	// 初期化リクエスト登録
	d.queue.PushRequest("init", nil)

	// 終了待機
	err := <-d.queue.Quit
	return err
}

func (d *JobQueue) dispatchMessageAction() {

	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(os.Stdout)
	out := log.Printf

	go func() {
		for {
			select {
			case msg := <-d.queue.Logs.Trace:
				if d.option.TraceLog {
					go out("[TRACE] %s\n", msg)
				}
			case msg := <-d.queue.Logs.Info:
				if d.option.InfoLog {
					go out("[INFO]  %s\n", msg)
				}
			case err := <-d.queue.Logs.Warn:
				if d.option.WarnLog {
					go out("[WARN]  %s\n", err)
				}
			case err := <-d.queue.Logs.Error:
				if d.option.ErrorLog {
					go out("[ERROR] %s\n", err)
				}
			}
		}
	}()
}

func (d *JobQueue) dispatchJobRequests() {

	router := NewRouter(d.queue, d.option)
	router.routes = d.jobRoutingTable
	go func() {
		for {
			select {
			case req := <-d.queue.Request:
				go router.Routing(req)
			}
		}
	}()
}

func (d *JobQueue) dispatchJobs() {

	for queueName, action := range d.workers {
		q := d.queue.Job[queueName]
		go func() {
			for {
				action(q)
			}
		}()
	}

}

func (d *JobQueue) AddDefaultActionWorker(workerName string, queueSize int) {
	d.AddWorker(workerName, queueSize, func(q chan JobAPI) {
		job := <-q
		go job.Start(d.queue, d.option)

	})
}

func (d *JobQueue) AddSerializedActionWorker(workerName string) {
	d.AddSerializedWithIntervalActionWorker(workerName, 0*time.Second)
}

func (d *JobQueue) AddSerializedWithIntervalActionWorker(workerName string, interval time.Duration) {
	d.AddWorker(workerName, 1, func(q chan JobAPI) {
		job := <-q
		time.Sleep(interval)
		go job.Start(d.queue, d.option)

	})
}
