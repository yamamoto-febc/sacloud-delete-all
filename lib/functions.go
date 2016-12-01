package lib

import (
	"fmt"
	"github.com/sacloud/libsacloud/api"
	"github.com/sacloud/libsacloud/sacloud"
	"github.com/sacloud/sackerel/version"
	"github.com/yamamoto-febc/jobq"
	"time"
)

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

func findAndDelete(queue *jobq.Queue, option *jobq.Option, job jobq.JobAPI) {
	target := job.GetPayload().(string)
	err := doActionPerZone(currentOption, func(client *api.Client) error {
		apiWrapper := getSacloudAPIWrapper(client, target)
		resources, err := apiWrapper.findFunc()
		if err != nil {
			return fmt.Errorf("target[%s] : %s", target, err)
		}
		for _, id := range resources {
			if apiWrapper.isAvaiableFunc != nil {
				isPowerOn, err := apiWrapper.isAvaiableFunc(id)
				if err != nil {
					return fmt.Errorf("target[%s] : %s", target, err)
				}
				if isPowerOn {
					_, err := apiWrapper.powerOffFunc(id)
					if err != nil {
						return fmt.Errorf("target[%s] : %s", target, err)
					}
					err = apiWrapper.waitForPoweroffFunc(id, client.DefaultTimeoutDuration)
					if err != nil {
						return fmt.Errorf("target[%s] : %s", target, err)
					}
				}
			}
			err := apiWrapper.deleteFunc(id)
			if err != nil {
				return fmt.Errorf("target[%s] : %s", target, err)
			}
		}

		return nil
	})
	if err != nil {
		queue.StopByError(err)
	} else {
		queue.PushRequest(fmt.Sprintf("%s:done", target), nil)
		wg.Done()
	}
}

func getClient(o *Option, zone string) *api.Client {

	client := api.NewClient(o.AccessToken, o.AccessTokenSecret, zone)
	client.TraceMode = o.TraceMode
	client.UserAgent = fmt.Sprintf("sacloud-delete-all/%s", version.Version)

	return client

}

type sacloudAPIWrapper struct {
	findFunc            func() ([]int64, error)
	isAvaiableFunc      func(int64) (bool, error)
	powerOffFunc        func(int64) (bool, error)
	waitForPoweroffFunc func(int64, time.Duration) error
	deleteFunc          func(int64) error
}

func getSacloudAPIWrapper(client *api.Client, target string) *sacloudAPIWrapper {
	switch target {
	case "archive":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.Archive.Find),
			deleteFunc: func(id int64) error { _, err := client.Archive.Delete(id); return err },
		}
	case "bridge":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.Bridge.Find),
			deleteFunc: func(id int64) error { _, err := client.Bridge.Delete(id); return err },
		}
	case "cdrom":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.CDROM.Find),
			deleteFunc: func(id int64) error { _, err := client.CDROM.Delete(id); return err },
		}
	case "disk":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.Disk.Find),
			deleteFunc: func(id int64) error { _, err := client.Disk.Delete(id); return err },
		}
	case "icon":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.Icon.Find),
			deleteFunc: func(id int64) error { _, err := client.Icon.Delete(id); return err },
		}
	case "internet":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.Internet.Find),
			deleteFunc: func(id int64) error { _, err := client.Internet.Delete(id); return err },
		}
	case "license":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.License.Find),
			deleteFunc: func(id int64) error { _, err := client.License.Delete(id); return err },
		}
	case "note":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.Note.Find),
			deleteFunc: func(id int64) error { _, err := client.Note.Delete(id); return err },
		}
	case "packetfilter":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.PacketFilter.Find),
			deleteFunc: func(id int64) error { _, err := client.PacketFilter.Delete(id); return err },
		}
	case "sshkey":
		return &sacloudAPIWrapper{
			findFunc:   createFindFund(client.SSHKey.Find),
			deleteFunc: func(id int64) error { _, err := client.SSHKey.Delete(id); return err },
		}
	case "switch":
		return &sacloudAPIWrapper{
			findFunc: createFindFund(client.Switch.Find),
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
			findFunc: func() ([]int64, error) {
				result, err := client.AutoBackup.Find()
				return toIdArray(result.CommonServiceAutoBackupItems), err
			},
			deleteFunc: func(id int64) error { _, err := client.AutoBackup.Delete(id); return err },
		}
	case "gslb":
		return &sacloudAPIWrapper{
			findFunc: func() ([]int64, error) {
				result, err := client.GSLB.Find()
				return toIdArray(result.CommonServiceGSLBItems), err
			},
			deleteFunc: func(id int64) error { _, err := client.GSLB.Delete(id); return err },
		}
	case "dns":
		return &sacloudAPIWrapper{
			findFunc: func() ([]int64, error) {
				result, err := client.DNS.Find()
				return toIdArray(result.CommonServiceDNSItems), err
			},
			deleteFunc: func(id int64) error { _, err := client.DNS.Delete(id); return err },
		}
	case "simplemonitor":
		return &sacloudAPIWrapper{
			findFunc: func() ([]int64, error) {
				result, err := client.SimpleMonitor.Find()
				return toIdArray(result.SimpleMonitors), err
			},
			deleteFunc: func(id int64) error { _, err := client.SimpleMonitor.Delete(id); return err },
		}

	case "server":
		return &sacloudAPIWrapper{
			findFunc:            createFindFund(client.Server.Find),
			isAvaiableFunc:      client.Server.IsUp,
			powerOffFunc:        client.Server.Stop,
			waitForPoweroffFunc: client.Server.SleepUntilDown,
			deleteFunc:          func(id int64) error { _, err := client.Server.Delete(id); return err },
		}
	case "database":
		return &sacloudAPIWrapper{
			findFunc: func() ([]int64, error) {
				result, err := client.Database.Find()
				return toIdArray(result.Databases), err
			},
			isAvaiableFunc:      client.Database.IsUp,
			powerOffFunc:        client.Database.Stop,
			waitForPoweroffFunc: client.Database.SleepUntilDown,
			deleteFunc:          func(id int64) error { _, err := client.Database.Delete(id); return err },
		}
	case "loadbalancer":
		return &sacloudAPIWrapper{
			findFunc: func() ([]int64, error) {
				result, err := client.LoadBalancer.Find()
				return toIdArray(result.LoadBalancers), err
			},
			isAvaiableFunc:      client.LoadBalancer.IsUp,
			powerOffFunc:        client.LoadBalancer.Stop,
			waitForPoweroffFunc: client.LoadBalancer.SleepUntilDown,
			deleteFunc:          func(id int64) error { _, err := client.LoadBalancer.Delete(id); return err },
		}
	case "vpcrouter":
		return &sacloudAPIWrapper{
			findFunc: func() ([]int64, error) {
				result, err := client.VPCRouter.Find()
				return toIdArray(result.VPCRouters), err
			},
			isAvaiableFunc:      client.VPCRouter.IsUp,
			powerOffFunc:        client.VPCRouter.Stop,
			waitForPoweroffFunc: client.VPCRouter.SleepUntilDown,
			deleteFunc:          func(id int64) error { _, err := client.VPCRouter.Delete(id); return err },
		}

	}

	return nil
}

func createFindFund(f func() (*sacloud.SearchResponse, error)) func() ([]int64, error) {
	return func() ([]int64, error) {
		result, err := f()
		var res []int64
		if err != nil {
			return res, err
		}
		res = append(res, toIdArray(result.Archives)...)
		res = append(res, toIdArray(result.Bridges)...)
		res = append(res, toIdArray(result.CDROMs)...)
		res = append(res, toIdArray(result.Disks)...)
		res = append(res, toIdArray(result.Icons)...)
		res = append(res, toIdArray(result.Internet)...)
		res = append(res, toIdArray(result.Licenses)...)
		res = append(res, toIdArray(result.Notes)...)
		res = append(res, toIdArray(result.PacketFilters)...)
		res = append(res, toIdArray(result.Servers)...)
		res = append(res, toIdArray(result.SSHKeys)...)
		res = append(res, toIdArray(result.Switches)...)
		return res, err
	}
}

func toIdArray(arr interface{}) []int64 {
	var res []int64 = []int64{}
	switch sl := arr.(type) {
	case []sacloud.Archive:
		for _, s := range sl {
			if s.Scope != string(sacloud.ESCopeUser) {
				continue
			}
			res = append(res, s.GetID())

		}
		break
	case []sacloud.Bridge:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.CDROM:
		for _, s := range sl {
			if s.Scope != string(sacloud.ESCopeUser) {
				continue
			}
			res = append(res, s.GetID())
		}
		break
	case []sacloud.Disk:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.Icon:
		for _, s := range sl {
			if s.Scope != string(sacloud.ESCopeUser) {
				continue
			}
			res = append(res, s.GetID())
		}
		break
	case []sacloud.Internet:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.License:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.Note:
		for _, s := range sl {
			if s.Scope != string(sacloud.ESCopeUser) {
				continue
			}
			res = append(res, s.GetID())
		}
		break
	case []sacloud.PacketFilter:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.Server:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.SSHKey:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break

	case []sacloud.Switch:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.AutoBackup:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.DNS:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.GSLB:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.SimpleMonitor:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.Database:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break

	case []sacloud.LoadBalancer:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break
	case []sacloud.VPCRouter:
		for _, s := range sl {
			res = append(res, s.GetID())
		}
		break

	}

	return res
}
