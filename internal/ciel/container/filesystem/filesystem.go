package filesystem

type FileSystem interface {
	MountLocal() error
	Mount(readOnly bool) error
	Unmount() error

	Rollback() error
	Merge() error
}
