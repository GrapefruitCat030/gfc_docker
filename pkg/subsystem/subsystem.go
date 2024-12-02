package subsystem

type ResourceConfig struct {
	MemLimit string
	CpuShare string
	CpuSet   string
}

type Subsystem interface {
	Name() string                                           // return the name of subsystem
	Set(cgroupPath string, resrcConf *ResourceConfig) error // set the resource limitation of the cgroup
	Apply(cgroupPath string, pid int) error                 // add a process to the cgroup
	Remove(cgroupPath string) error                         // remove the cgroup
}

var SubsystemIns = []Subsystem{
	&MemorySubsystem{},
}
