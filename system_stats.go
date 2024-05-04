package main

import (
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
	"time"
)

func getSystemStats(showCPU bool, showRAM bool, showDisk bool, showProcs bool) (*SystemStats, error) {
	if !showCPU && !showRAM && !showDisk && !showProcs {
		return nil, fmt.Errorf("No stats requested")
	}

	// get timestamp to calculate how long execution took
	start := time.Now()
	memStats, err := getRAMStats(showRAM)
	if err != nil {
		return nil, err
	}
	println(fmt.Sprintf("RAM stats took %v", time.Since(start)))
	start = time.Now()

	cpuStats, err := getCPUStats(showCPU)
	if err != nil {
		return nil, err
	}
	println(fmt.Sprintf("CPU stats took %v", time.Since(start)))
	start = time.Now()

	diskStats, err := getDiskStats(showDisk)
	if err != nil {
		return nil, err
	}
	println(fmt.Sprintf("Disk stats took %v", time.Since(start)))
	start = time.Now()

	procs, err := getProcesses(showProcs)
	if err != nil {
		return nil, err
	}
	println(fmt.Sprintf("Process stats took %v", time.Since(start)))

	return &SystemStats{
		RAM:   memStats,
		CPU:   cpuStats,
		Disk:  diskStats,
		Procs: procs,
	}, nil
}

func getRAMStats(shown bool) (*RamStats, error) {
	if !shown {
		return nil, nil
	}
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return nil, err
	}

	return &RamStats{
		Total: vmem.Total,
		Free:  vmem.Free,
		Used:  vmem.Used,
	}, nil
}

func getCPUStats(shown bool) (*CpuStats, error) {
	if !shown {
		return nil, nil
	}

	return &CpuStats{
		//Total:  cpuPercent[0],
		PerCpu: cpuPercent,
	}, nil
}

func getDiskStats(shown bool) (*[]DiskStats, error) {
	if !shown {
		return nil, nil
	}
	diskPartitions, err := disk.Partitions(true)
	if err != nil {
		return nil, err
	}

	var diskStats []DiskStats
	for _, partition := range diskPartitions {
		diskStat, err := getDiskStatsForPath(partition.Mountpoint)
		if err != nil {
			continue
		}
		diskStats = append(diskStats, diskStat)
	}

	return &diskStats, nil
}

func getProcesses(shown bool) (*[]Processes, error) {
	if !shown {
		return nil, nil
	}

	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var processes = &[]Processes{}
	for _, p := range procs {
		name, err := p.Name()
		if err != nil {
			continue
		}

		pid := p.Pid
		mem, err := p.MemoryInfo()
		if err != nil {
			continue
		}

		cpu, err := p.CPUPercent()
		if err != nil {
			continue
		}

		proc := Processes{
			Pid:  int(pid),
			Name: name,
			Ram:  mem.RSS,
			Cpu:  cpu,
		}

		*processes = append(*processes, proc)
	}

	return processes, nil
}

func getDiskStatsForPath(diskPath string) (DiskStats, error) {
	diskStat, err := disk.Usage(diskPath)
	if err != nil {
		return DiskStats{}, err
	}

	if diskStat.Total == 0 {
		return DiskStats{}, fmt.Errorf("Disk not found")
	}

	return DiskStats{
		Total: diskStat.Total,
		Free:  diskStat.Free,
		Used:  diskStat.Used,
		Name:  diskStat.Path,
	}, nil
}
