package main

import (
	"log"
	"time"

	geoipgo "geoip-go"
)

type CidrInfo struct {
	CountryCode string
}

type CidrList = geoipgo.CidrList[geoipgo.StringCidr, CidrInfo]
type IPDataSource = geoipgo.IPDataSource[geoipgo.StringCidr, CidrInfo]
type IPManager = geoipgo.IPManager[geoipgo.StringCidr, CidrInfo]

type MockIPDataSource struct{}

func (m *MockIPDataSource) Fetch() (CidrList, error) {
	return CidrList{
		{Cidr: geoipgo.StringCidr("192.168.0.0/16"), Info: CidrInfo{"CN"}},
		{Cidr: geoipgo.StringCidr("10.0.0.0/8"), Info: CidrInfo{"US"}},
		{Cidr: geoipgo.StringCidr("172.16.0.0/12"), Info: CidrInfo{"JP"}},
		{Cidr: geoipgo.StringCidr("8.8.8.0/24"), Info: CidrInfo{"US"}},
	}, nil
}

type IPUpdater struct {
	manager    IPManager
	dataSource IPDataSource
	interval   time.Duration
	updateTime time.Time
}

func NewIPUpdater(
	manager IPManager,
	dataSource IPDataSource,
	updateTime time.Time,
	interval time.Duration,
) *IPUpdater {
	return &IPUpdater{
		manager:    manager,
		dataSource: dataSource,
		updateTime: updateTime,
		interval:   interval,
	}
}

func (u *IPUpdater) Start() {
	go func() {
		for {
			// 计算下次更新时间
			now := time.Now()
			nextUpdate := time.Date(
				now.Year(), now.Month(), now.Day(),
				u.updateTime.Hour(), u.updateTime.Minute(),
				u.updateTime.Second(), 0, time.Local,
			)

			// 如果当前时间已经超过今天的更新时间，调整到明天
			if now.After(nextUpdate) {
				nextUpdate = nextUpdate.AddDate(0, 0, 1)
			}

			// 等待到更新时间
			duration := nextUpdate.Sub(now)
			log.Printf("下次更新时间: %v", nextUpdate)

			time.Sleep(duration)

			// 获取并更新数据
			ipData, err := u.dataSource.Fetch()
			if err != nil {
				log.Printf("更新失败: %v", err)
				continue
			}

			if err := u.manager.Update(ipData); err != nil {
				log.Printf("更新失败: %v", err)
				continue
			}

			log.Println("IP 数据已更新")

			// 等待间隔周期
			time.Sleep(u.interval)
		}
	}()
}

func main() {
	// 使用Mock模拟数据源初始化 manager
	dataSource := &MockIPDataSource{}
	manager := geoipgo.NewSyncIPManager(dataSource)

	// 创建更新器，每天凌晨 2 点更新，间隔24小时
	updateTime := time.Date(0, 0, 0, 2, 0, 0, 0, time.Local)
	updater := NewIPUpdater(manager, dataSource, updateTime, 24*time.Hour)
	updater.Start()

	// 测试查询
	testIPs := []string{"192.168.1.1", "10.1.1.1", "8.8.8.8", "172.16.5.5", "1.2.3.4"}
	for {
		for _, ip := range testIPs {
			cidrInfo := manager.Query(ip)
			if cidrInfo != nil {
				println("IP:", ip, "CountryCode:", cidrInfo.CountryCode)
			} else {
				println("IP:", ip, "CountryCode: 未知")
			}
		}
		// 每隔5秒查询一次
		time.Sleep(5 * time.Second)
	}
}
