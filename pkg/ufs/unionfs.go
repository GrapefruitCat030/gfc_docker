package ufs

type UnionFSer interface {
	WorkSpace() string

	CreateReadOnlyLayer(rootPath string) error
	CreateWriteLayer(rootPath string) error
	DeleteWriteLayer(rootPath string) error

	MountUFS(rootPath string) error
	UMountUFS(rootPath string) error

	MountVolume(rootPath, volumeMapPath string) error
	UMountVolume(rootPath, volumeMapPath string) error
}

func NewWorkSpace(rootPath, volumeMapPath string, unionFSer UnionFSer) error {
	if err := unionFSer.CreateReadOnlyLayer(rootPath); err != nil {
		return err
	}
	if err := unionFSer.CreateWriteLayer(rootPath); err != nil {
		return err
	}
	if err := unionFSer.MountUFS(rootPath); err != nil {
		return err
	}
	if volumeMapPath != "" {
		if err := unionFSer.MountVolume(rootPath, volumeMapPath); err != nil {
			return err
		}
	}
	return nil
}

func DeleteWorkSpace(rootPath, volumeMapPath string, unionFSer UnionFSer) error {
	if volumeMapPath != "" {
		if err := unionFSer.UMountVolume(rootPath, volumeMapPath); err != nil {
			return err
		}
	}
	if err := unionFSer.UMountUFS(rootPath); err != nil {
		return err
	}
	if err := unionFSer.DeleteWriteLayer(rootPath); err != nil {
		return err
	}
	return nil
}
