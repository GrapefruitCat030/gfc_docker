package ufs

type UnionFSer interface {
	WorkSpace() string
	CreateReadOnlyLayer(rootPath string) error
	CreateWriteLayer(rootPath string) error
	MountUFS(rootPath string) error
	DeleteWriteLayer(rootPath string) error
	UMountUFS(rootPath string) error
}

func NewWorkSpace(rootPath string, unionFSer UnionFSer) error {
	if err := unionFSer.CreateReadOnlyLayer(rootPath); err != nil {
		return err
	}
	if err := unionFSer.CreateWriteLayer(rootPath); err != nil {
		return err
	}
	if err := unionFSer.MountUFS(rootPath); err != nil {
		return err
	}
	return nil
}

func DeleteWorkSpace(rootPath string, unionFSer UnionFSer) error {
	if err := unionFSer.UMountUFS(rootPath); err != nil {
		return err
	}
	if err := unionFSer.DeleteWriteLayer(rootPath); err != nil {
		return err
	}
	return nil
}
