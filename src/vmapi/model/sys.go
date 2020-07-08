package model

type SYS struct {
	MEM MEMINFO `json:"mem"`
	DISK PARTION `json:"disk"`
	CPU CPUINFO `json:"cpu"`
}

type MEMINFO struct {
	AVAILABLE uint64 `json:"available"`
	USED uint64 `json:"used"`
	CACHED uint64 `json:"cached"`
	PERCENT float64 `json:"percent"`
	FREE uint64 `json:"free"`
	INACTIVE uint64 `json:"inactive"`
	ACTIVE uint64 `json:"active"`
	SHARED uint64 `json:"shared"`
	TOTAL uint64 `json:"total"`
	SLAB uint64 `json:"slab"`
	BUFFERS uint64 `json:"buffers"`
}

type PARTION struct {
	PAR DISKINFO `json:"/ddhome"`
}

type DISKINFO struct {
	TOTAL uint64 `json:"total"`
	PERCENT uint64 `json:"percent"`
	FREE uint64 `json:"free"`
	USED uint64 `json:"used"`
}

type CPUINFO struct {
	COUNT int `json:"count"`
	CPUPERCENT float64 `json:"cpu_percent"`
}