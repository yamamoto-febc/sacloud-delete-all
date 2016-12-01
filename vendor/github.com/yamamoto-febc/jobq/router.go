package jobq

import (
	"fmt"
)

// Router ジョブキューに対するルーティング定義/処理を提供する
type Router struct {
	queue  *Queue
	option *Option
	routes map[string]JobRouterFunc
}

// NewRouter Routerの新規作成
func NewRouter(queue *Queue, option *Option) *Router {
	r := &Router{
		queue:  queue,
		option: option,
	}
	//r.buildRouteDefines()
	return r
}

// Routing ジョブキューへの処理リクエストを定義に従いルーティングし、適切なワーカーを呼び出す
func (r *Router) Routing(req JobRequestAPI) {

	payload := req.GetPayload()
	r.queue.PushTrace(fmt.Sprintf("request => '%s' payload => (%#v)", req.GetName(), payload))

	if route, ok := r.routes[req.GetName()]; ok {
		route(r.queue, r.option, req)
	} else {
		r.queue.PushWarn(fmt.Errorf("Route('%s') is not found.", req.GetName()))
	}
}
