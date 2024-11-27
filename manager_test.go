package geoipgo

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

type MultiFileIPDataSource struct {
	directory string
}

type StringCidrList = CidrList[StringCidr, string]
type StringCidrItem = CidrItem[StringCidr, string]

func (s *MultiFileIPDataSource) Fetch() (StringCidrList, error) {
	var ipData StringCidrList

	err := filepath.Walk(s.directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() || !strings.HasSuffix(path, ".txt") {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		countryCode := extractCountryCode(filepath.Base(path))

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			cidr := strings.TrimSpace(scanner.Text())
			if cidr != "" && !strings.HasPrefix(cidr, "#") {
				ipData = append(ipData, StringCidrItem{
					Cidr: StringCidr(cidr),
					Info: countryCode,
				})
			}
		}

		if err := scanner.Err(); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return ipData, nil
}

func extractCountryCode(filename string) string {
	parts := strings.Split(filename, "-")
	if len(parts) > 0 {
		return strings.ToUpper(strings.TrimSuffix(parts[0], ".txt"))
	}
	return "UNKNOWN"
}

func BenchmarkIPManagerLoad(b *testing.B) {
	dataSourcePath := "./test"

	dataSource := &MultiFileIPDataSource{directory: dataSourcePath}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		manager := NewSyncIPManager(dataSource)
		_ = manager
	}
}

func BenchmarkIPManagerQuery(b *testing.B) {
	dataSourcePath := "./test"

	dataSource := &MultiFileIPDataSource{directory: dataSourcePath}
	manager := NewSyncIPManager(dataSource)

	testIPs := []string{
		"1.1.1.1",
		"8.8.8.8",
		"114.114.114.114",
		"223.5.5.5",
		"220.181.38.148",
		"192.168.1.1",
		"10.0.0.1",
		"172.16.0.1",
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			for _, ip := range testIPs {
				result := manager.Query(ip)
				_ = result
			}
		}
	})
}

func BenchmarkConcurrentQueries(b *testing.B) {
	dataSourcePath := "./test"

	dataSource := &MultiFileIPDataSource{directory: dataSourcePath}
	manager := NewSyncIPManager(dataSource)

	var testIPs []string
	for i := 0; i < 10000; i++ {
		testIPs = append(testIPs, fmt.Sprintf("%d.%d.%d.%d",
			1+(i%254),
			1+((i/254)%254),
			1+((i/(254*254))%254),
			1+(i%254)))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			var wg sync.WaitGroup
			for _, ip := range testIPs {
				wg.Add(1)
				go func(testIP string) {
					defer wg.Done()
					result := manager.Query(testIP)
					_ = result
				}(ip)
			}
			wg.Wait()
		}
	})
}

func TestProfileIPManager(t *testing.T) {
	dataSourcePath := "./test"

	dataSource := &MultiFileIPDataSource{directory: dataSourcePath}
	manager := NewSyncIPManager(dataSource)

	// 预热
	for i := 0; i < 1000; i++ {
		ip := fmt.Sprintf("1.2.%d.%d", i%256, i%256)
		manager.Query(ip)
	}

	start := time.Now()
	iterations := 100000
	for i := 0; i < iterations; i++ {
		ip := fmt.Sprintf("1.2.%d.%d", i%256, i%256)
		manager.Query(ip)
	}
	duration := time.Since(start)

	t.Logf("查询性能:\n")
	t.Logf("总迭代次数: %d\n", iterations)
	t.Logf("总耗时: %v\n", duration)
	t.Logf("每次查询平均耗时: %v\n", duration/time.Duration(iterations))
}
