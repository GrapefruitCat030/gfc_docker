package cgroup

import (
	gfc_subsys "github.com/GrapefruitCat030/gfc_docker/pkg/subsystem"
)

type CgroupManager struct {
	Path     string
	Resource *gfc_subsys.ResourceConfig
}

func NewCgroupManager(path string) *CgroupManager {
	return &CgroupManager{
		Path: path,
	}
}

func (m *CgroupManager) Set() error {
	for _, subSysIns := range gfc_subsys.SubsystemIns {
		if err := subSysIns.Set(m.Path, m.Resource); err != nil {
			return err
		}
	}
	return nil
}

func (m *CgroupManager) Apply(pid int) error {
	for _, subSysIns := range gfc_subsys.SubsystemIns {
		if err := subSysIns.Apply(m.Path, pid); err != nil {
			return err
		}
	}
	return nil
}

func (m *CgroupManager) Remove() error {
	for _, subSysIns := range gfc_subsys.SubsystemIns {
		if err := subSysIns.Remove(m.Path); err != nil {
			return err
		}
	}
	return nil
}
