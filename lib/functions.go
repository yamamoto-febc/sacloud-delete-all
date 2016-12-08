package lib

import (
	"fmt"
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/yamamoto-febc/jobq"
	"github.com/yamamoto-febc/sacloud-delete-all/version"
	"strings"
	"sync"
	"time"
)

type ParallelJobPayload struct {
	RouteName string
	Targets   []string
}

func doActionPerZone(option *Option, sacloudAPIFunc func(*api.Client) error) error {
	for _, zone := range option.Zones {

		client := getClient(option, zone)
		// call API func per zone.
		err := sacloudAPIFunc(client)
		if err != nil {
			return err
		}
	}

	return nil
}

func FindAndDeleteJob(target string) func(interface{}) jobq.JobAPI {
	return func(p interface{}) jobq.JobAPI {
		return jobq.NewJob(fmt.Sprintf("FindAndDelete:%s", target), findAndDelete, target)
	}
}

func FindAndDeleteJobParallel(routeName string, targets ...string) func(interface{}) jobq.JobAPI {

	payload := ParallelJobPayload{
		RouteName: routeName,
		Targets:   targets,
	}
	targetName := strings.Join(targets, ",")

	return func(p interface{}) jobq.JobAPI {
		return jobq.NewJob(fmt.Sprintf("FindAndDeleteParallel:%s", targetName), findAndDeleteParallel, payload)
	}
}

func findAndDeleteParallel(queue *jobq.Queue, option *jobq.Option, job jobq.JobAPI) {
	targets := job.GetPayload().(ParallelJobPayload)
	var wg sync.WaitGroup
	wg.Add(len(targets.Targets))
	for _, target := range targets.Targets {
		go func(t string) {
			err := doFindAndDelete(queue, option, t)
			if err != nil {
				queue.StopByError(err)
			} else {
				queue.PushRequest(fmt.Sprintf("%s:done", target), nil)
				resourceWaitGroup.Done()
				wg.Done()
			}
		}(target)
	}
	wg.Wait()
	queue.PushRequest(fmt.Sprintf("%s:done", targets.RouteName), nil)
}

func findAndDelete(queue *jobq.Queue, option *jobq.Option, job jobq.JobAPI) {
	target := job.GetPayload().(string)
	err := doFindAndDelete(queue, option, target)
	if err != nil {
		queue.StopByError(err)
	} else {
		queue.PushRequest(fmt.Sprintf("%s:done", target), nil)
		resourceWaitGroup.Done()
	}
}

func doFindAndDelete(queue *jobq.Queue, option *jobq.Option, target string) error {
	return doActionPerZone(currentOption, func(client *api.Client) error {
		apiWrapper := getSacloudAPIWrapper(client, target)
		resources, err := apiWrapper.findFunc()
		if err != nil {
			return fmt.Errorf("target[%s](%s) : %s", target, client.Zone, err)
		}

		var wg sync.WaitGroup
		wg.Add(len(resources))

		for _, resource := range resources {

			go func(r sacloudResourceWrapper) {
				id := r.id
				name := r.name
				if apiWrapper.isAvaiableFunc != nil {
					isPowerOn, err := apiWrapper.isAvaiableFunc(id)
					if err != nil {
						queue.StopByError(fmt.Errorf("%-26s : resource(id:%d,name:%s) : %s", fmt.Sprintf("target[%s/%s]", target, client.Zone), id, name, err))
						return
					}
					if isPowerOn {
						_, err := apiWrapper.powerOffFunc(id)
						if err != nil {
							queue.StopByError(fmt.Errorf("%-26s : resource(id:%d,name:%s) : %s", fmt.Sprintf("target[%s/%s]", target, client.Zone), id, name, err))
							return
						}
						err = apiWrapper.waitForPoweroffFunc(id, client.DefaultTimeoutDuration)
						if err != nil {
							queue.StopByError(fmt.Errorf("%-26s : resource(id:%d,name:%s) : %s", fmt.Sprintf("target[%s/%s]", target, client.Zone), id, name, err))
							return
						}
					}
				}
				err := apiWrapper.deleteFunc(id)
				if err != nil {
					queue.StopByError(fmt.Errorf("%-26s : resource(id:%d,name:%s) : %s", fmt.Sprintf("target[%s/%s]", target, client.Zone), id, name, err))
					return
				}
				queue.PushInfo(fmt.Sprintf("%-26s : resource(id:%d,name:%s) deleted.", fmt.Sprintf("target[%s/%s]", target, client.Zone), id, name))
				wg.Done()
			}(resource)
		}

		wg.Wait()
		return nil
	})
}

func getClient(o *Option, zone string) *api.Client {

	client := api.NewClient(o.AccessToken, o.AccessTokenSecret, zone)
	client.TraceMode = o.TraceMode
	client.UserAgent = fmt.Sprintf("sacloud-delete-all/%s", version.Version)

	return client

}

type sacloudAPIWrapper struct {
	findFunc            func() ([]sacloudResourceWrapper, error)
	isAvaiableFunc      func(int64) (bool, error)
	powerOffFunc        func(int64) (bool, error)
	waitForPoweroffFunc func(int64, time.Duration) error
	deleteFunc          func(int64) error
}
type sacloudResourceWrapper struct {
	id   int64
	name string
}

func getSacloudAPIWrapper(client *api.Client, target string) *sacloudAPIWrapper {
	switch target {
	case "archive":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.Archive.Find),
			deleteFunc: func(id int64) error { _, err := client.Archive.Delete(id); return err },
		}
	case "bridge":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.Bridge.Find),
			deleteFunc: func(id int64) error { _, err := client.Bridge.Delete(id); return err },
		}
	case "cdrom":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.CDROM.Find),
			deleteFunc: func(id int64) error { _, err := client.CDROM.Delete(id); return err },
		}
	case "disk":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.Disk.Find),
			deleteFunc: func(id int64) error { _, err := client.Disk.Delete(id); return err },
		}
	case "icon":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.Icon.Find),
			deleteFunc: func(id int64) error { _, err := client.Icon.Delete(id); return err },
		}
	case "internet":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.Internet.Find),
			deleteFunc: func(id int64) error { _, err := client.Internet.Delete(id); return err },
		}
	case "license":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.License.Find),
			deleteFunc: func(id int64) error { _, err := client.License.Delete(id); return err },
		}
	case "note":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.Note.Find),
			deleteFunc: func(id int64) error { _, err := client.Note.Delete(id); return err },
		}
	case "packetfilter":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.PacketFilter.Find),
			deleteFunc: func(id int64) error { _, err := client.PacketFilter.Delete(id); return err },
		}
	case "sshkey":
		return &sacloudAPIWrapper{
			findFunc:   createFindFunc(client.SSHKey.Find),
			deleteFunc: func(id int64) error { _, err := client.SSHKey.Delete(id); return err },
		}
	case "switch":
		return &sacloudAPIWrapper{
			findFunc: createFindFunc(client.Switch.Find),
			deleteFunc: func(id int64) error {
				// ブリッジより先に実行
				// もしブリッジ接続があれば切断
				sw, err := client.Switch.Read(id)
				if err != nil {
					return err
				}
				if sw.Bridge != nil {
					_, err := client.Switch.DisconnectFromBridge(id)
					if err != nil {
						return err
					}
				}
				_, err = client.Switch.Delete(id)
				return err
			},
		}
	case "autobackup":
		return &sacloudAPIWrapper{
			findFunc: func() ([]sacloudResourceWrapper, error) {
				result, err := client.AutoBackup.Find()
				return toResourceList(result.CommonServiceAutoBackupItems), err
			},
			deleteFunc: func(id int64) error { _, err := client.AutoBackup.Delete(id); return err },
		}
	case "gslb":
		return &sacloudAPIWrapper{
			findFunc: func() ([]sacloudResourceWrapper, error) {
				result, err := client.GSLB.Find()
				return toResourceList(result.CommonServiceGSLBItems), err
			},
			deleteFunc: func(id int64) error { _, err := client.GSLB.Delete(id); return err },
		}
	case "dns":
		return &sacloudAPIWrapper{
			findFunc: func() ([]sacloudResourceWrapper, error) {
				result, err := client.DNS.Find()
				return toResourceList(result.CommonServiceDNSItems), err
			},
			deleteFunc: func(id int64) error { _, err := client.DNS.Delete(id); return err },
		}
	case "simplemonitor":
		return &sacloudAPIWrapper{
			findFunc: func() ([]sacloudResourceWrapper, error) {
				result, err := client.SimpleMonitor.Find()
				return toResourceList(result.SimpleMonitors), err
			},
			deleteFunc: func(id int64) error { _, err := client.SimpleMonitor.Delete(id); return err },
		}

	case "server":
		return &sacloudAPIWrapper{
			findFunc:            createFindFunc(client.Server.Find),
			isAvaiableFunc:      client.Server.IsUp,
			powerOffFunc:        client.Server.Stop,
			waitForPoweroffFunc: client.Server.SleepUntilDown,
			deleteFunc:          func(id int64) error { _, err := client.Server.Delete(id); return err },
		}
	case "database":
		return &sacloudAPIWrapper{
			findFunc: func() ([]sacloudResourceWrapper, error) {
				result, err := client.Database.Find()
				return toResourceList(result.Databases), err
			},
			isAvaiableFunc:      client.Database.IsUp,
			powerOffFunc:        client.Database.Stop,
			waitForPoweroffFunc: client.Database.SleepUntilDown,
			deleteFunc:          func(id int64) error { _, err := client.Database.Delete(id); return err },
		}
	case "loadbalancer":
		return &sacloudAPIWrapper{
			findFunc: func() ([]sacloudResourceWrapper, error) {
				result, err := client.LoadBalancer.Find()
				return toResourceList(result.LoadBalancers), err
			},
			isAvaiableFunc:      client.LoadBalancer.IsUp,
			powerOffFunc:        client.LoadBalancer.Stop,
			waitForPoweroffFunc: client.LoadBalancer.SleepUntilDown,
			deleteFunc:          func(id int64) error { _, err := client.LoadBalancer.Delete(id); return err },
		}
	case "vpcrouter":
		return &sacloudAPIWrapper{
			findFunc: func() ([]sacloudResourceWrapper, error) {
				result, err := client.VPCRouter.Find()
				return toResourceList(result.VPCRouters), err
			},
			isAvaiableFunc:      client.VPCRouter.IsUp,
			powerOffFunc:        client.VPCRouter.Stop,
			waitForPoweroffFunc: client.VPCRouter.SleepUntilDown,
			deleteFunc:          func(id int64) error { _, err := client.VPCRouter.Delete(id); return err },
		}

	}

	return nil
}

func createFindFunc(f func() (*sacloud.SearchResponse, error)) func() ([]sacloudResourceWrapper, error) {
	return func() ([]sacloudResourceWrapper, error) {
		result, err := f()
		var res []sacloudResourceWrapper
		if err != nil {
			return res, err
		}
		res = append(res, toResourceList(result.Archives)...)
		res = append(res, toResourceList(result.Bridges)...)
		res = append(res, toResourceList(result.CDROMs)...)
		res = append(res, toResourceList(result.Disks)...)
		res = append(res, toResourceList(result.Icons)...)
		res = append(res, toResourceList(result.Internet)...)
		res = append(res, toResourceList(result.Licenses)...)
		res = append(res, toResourceList(result.Notes)...)
		res = append(res, toResourceList(result.PacketFilters)...)
		res = append(res, toResourceList(result.Servers)...)
		res = append(res, toResourceList(result.SSHKeys)...)
		res = append(res, toResourceList(result.Switches)...)
		return res, err
	}
}

func toResourceList(arr interface{}) []sacloudResourceWrapper {
	var res []sacloudResourceWrapper = []sacloudResourceWrapper{}
	switch sl := arr.(type) {
	case []sacloud.Archive:
		for _, s := range sl {
			if s.Scope != string(sacloud.ESCopeUser) {
				continue
			}
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})

		}
		break
	case []sacloud.Bridge:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.CDROM:
		for _, s := range sl {
			if s.Scope != string(sacloud.ESCopeUser) {
				continue
			}
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.Disk:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.Icon:
		for _, s := range sl {
			if s.Scope != string(sacloud.ESCopeUser) {
				continue
			}
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.Internet:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.License:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.Note:
		for _, s := range sl {
			if s.Scope != string(sacloud.ESCopeUser) {
				continue
			}
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.PacketFilter:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.Server:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.SSHKey:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break

	case []sacloud.Switch:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.AutoBackup:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.DNS:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.GSLB:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.SimpleMonitor:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.Database:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break

	case []sacloud.LoadBalancer:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break
	case []sacloud.VPCRouter:
		for _, s := range sl {
			res = append(res, sacloudResourceWrapper{id: s.GetID(), name: s.Name})
		}
		break

	}

	return res
}
