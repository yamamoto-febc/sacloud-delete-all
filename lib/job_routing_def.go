package lib

import (
	"github.com/yamamoto-febc/jobq"
	"github.com/yamamoto-febc/jobq/routing"
)

var routes map[string]jobq.JobRouterFunc = map[string]jobq.JobRouterFunc{
	"init": routing.Parallel(
		"parallel-delete",
		"serialize-delete",
		"wait-for-delete",
	),
	"parallel-delete": routing.Parallel(
		"archive",
		"autobackup",
		"cdrom",
		"packetfilter",
		"gslb",
		"dns",
		"simplemonitor",
		"license",
		"sshkey",
		"note",
		"icon",
	),
	"serialize-delete": routing.PathThrough("appliance"),

	"appliance":         routing.Action("sacloud", FindAndDeleteJobParallel("appliance", "loadbalancer", "vpcrouter", "database")),
	"appliance:done":    routing.PathThrough("server"),
	"loadbalancer:done": routing.Goal,
	"vpcrouter:done":    routing.Goal,
	"database:done":     routing.Goal,

	"server":      routing.Action("sacloud", FindAndDeleteJob("server")),
	"server:done": routing.PathThrough("after-server"),

	"after-server": routing.Parallel("internet", "disk"),

	"internet":      routing.Action("sacloud", FindAndDeleteJob("internet")),
	"internet:done": routing.PathThrough("switch"),
	"switch":        routing.Action("sacloud", FindAndDeleteJob("switch")),
	"switch:done":   routing.PathThrough("bridge"),

	"archive":            routing.Action("sacloud", FindAndDeleteJob("archive")),
	"archive:done":       routing.Goal,
	"autobackup":         routing.Action("sacloud", FindAndDeleteJob("autobackup")),
	"autobackup:done":    routing.Goal,
	"bridge":             routing.Action("sacloud", FindAndDeleteJob("bridge")),
	"bridge:done":        routing.Goal,
	"cdrom":              routing.Action("sacloud", FindAndDeleteJob("cdrom")),
	"cdrom:done":         routing.Goal,
	"disk":               routing.Action("sacloud", FindAndDeleteJob("disk")),
	"disk:done":          routing.Goal,
	"packetfilter":       routing.Action("sacloud", FindAndDeleteJob("packetfilter")),
	"packetfilter:done":  routing.Goal,
	"gslb":               routing.Action("sacloud", FindAndDeleteJob("gslb")),
	"gslb:done":          routing.Goal,
	"dns":                routing.Action("sacloud", FindAndDeleteJob("dns")),
	"dns:done":           routing.Goal,
	"simplemonitor":      routing.Action("sacloud", FindAndDeleteJob("simplemonitor")),
	"simplemonitor:done": routing.Goal,
	"license":            routing.Action("sacloud", FindAndDeleteJob("license")),
	"license:done":       routing.Goal,
	"sshkey":             routing.Action("sacloud", FindAndDeleteJob("sshkey")),
	"sshkey:done":        routing.Goal,
	"note":               routing.Action("sacloud", FindAndDeleteJob("note")),
	"note:done":          routing.Goal,
	"icon":               routing.Action("sacloud", FindAndDeleteJob("icon")),
	"icon:done":          routing.Goal,

	"wait-for-delete": func(queue *jobq.Queue, option *jobq.Option, req jobq.JobRequestAPI) {
		resourceWaitGroup.Wait()
		queue.Stop()
	},
}
