package main

import (
	"github.com/gorilla/websocket"
	"github.com/hpcloud/tail"
	"os"
	"time"
)

type SystemStats struct {
	RAM   *RamStats    `json:"ram,omitempty"`
	CPU   *CpuStats    `json:"cpu,omitempty"`
	Disk  *[]DiskStats `json:"disk,omitempty"`
	Procs *[]Processes `json:"procs,omitempty"`
}

type RamStats struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
	Used  uint64 `json:"used"`
}

type Processes struct {
	Pid      int     `json:"pid"`
	Name     string  `json:"name"`
	Ram      uint64  `json:"ram"`
	Cpu      float64 `json:"cpu"`
	Killable bool    `json:"killable"`
}

type CpuStats struct {
	//Total  float64   `json:"total"`
	PerCpu []float64 `json:"per_cpu"`
}

type DiskStats struct {
	Total uint64 `json:"total"`
	Free  uint64 `json:"free"`
	Used  uint64 `json:"used"`
	Name  string `json:"name"`
}

type SystemInfo struct {
	Hostname string    `json:"hostname"`
	OS       string    `json:"os"`
	Kernel   string    `json:"kernel"`
	Uptime   int       `json:"uptime"`
	RAM      uint64    `json:"ram"`
	Swap     uint64    `json:"swap"`
	CPU      []CPUInfo `json:"cpu"`
}

type CPUInfo struct {
	Count int     `json:"count"`
	Brand string  `json:"brand"`
	Cores int     `json:"cores"`
	Mhz   float64 `json:"mhz"`
}

type TokenStatus struct {
	Valid bool      `json:"valid"`
	Exp   time.Time `json:"exp"`
}

type DirResponse struct {
	Directory string      `json:"directory"`
	Dirs      []Directory `json:"dirs"`
}

type Directory struct {
	Path  string `json:"path"`
	IsDir bool   `json:"is_dir"`
	Size  int64  `json:"size"`
}

type logConnection struct {
	conn *websocket.Conn
	file *os.File
	tail *tail.Tail
}
