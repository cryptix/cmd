package main

type Sysinfo_t struct {
	Uptime int64     // kern.boottime
	Loads  [3]uint64 // vm.loadavg

	Totalram  uint64
	Freeram   uint64
	Sharedram uint64
	Bufferram uint64

	Totalswap uint64
	Freeswap  uint64

	Procs uint16

	Totalhigh uint64
	Freehigh  uint64

	Unit uint32
}
