package main

import (
	"github.com/shirou/gopsutil/v3/cpu"
	"os"
	"runtime"
)

func getCPUInfo(cpuModel string) (CPUInfo, error) {
	cpuInfo, err := cpu.Info()
	if err != nil {
		return CPUInfo{}, err
	}

	returnCPU := CPUInfo{
		Brand: cpuModel,
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
	for model := range cpuModels {
		info, err := getCPUInfo(model)
		if err != nil {
			return nil, err
		}
		cpu = append(cpu, info)
	}

	return &SystemInfo{
		Hostname: host,
		OS:       runtime.GOOS,
		Kernel:   runtime.GOARCH,
		CPU:      cpu,
	}, nil
}
