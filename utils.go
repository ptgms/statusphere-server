package main

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"os"
	"runtime"
	"strings"
	"time"
)

func getCPUInfo(cpuModel string, count int) (CPUInfo, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return CPUInfo{}, err
	}

	returnCPU := CPUInfo{
		Count: count,
		Brand: strings.Trim(cpuModel, " "),
		Cores: 0,
		Mhz:   0,
	}

	for _, info := range cpuInfo {
		if info.ModelName != cpuModel {
			continue
		}
		returnCPU.Cores += int(info.Cores)
		returnCPU.Mhz = info.Mhz
	}

	return returnCPU, nil
}

func getSystemInfo() (*SystemInfo, error) {
	host, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	// get all unique CPU models
	cpuInfo, err := cpu.Info()
	if err != nil {
		return nil, err
	}

	cpuModels := make(map[string]bool)
	for _, info := range cpuInfo {
		cpuModels[info.ModelName] = true
	}

	var cpu []CPUInfo
	count := 1
	for model := range cpuModels {
		info, err := getCPUInfo(model, count)
		if err != nil {
			return nil, err
		}
		cpu = append(cpu, info)
		count++
	}
	// make uint64 var for RAM
	var ram uint64
	var swap uint64

	vmem, err := mem.VirtualMemory()
	if err == nil {
		ram = vmem.Total
		swap = vmem.SwapTotal
	}

	return &SystemInfo{
		Hostname: host,
		OS:       runtime.GOOS,
		Kernel:   runtime.GOARCH,
		Uptime:   int(time.Since(startTime).Seconds()),
		RAM:      ram,
		Swap:     swap,
		CPU:      cpu,
	}, nil
}
