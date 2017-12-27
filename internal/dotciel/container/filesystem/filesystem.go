package filesystem

type FileSystem interface {
	Mount(readOnly bool) error
	Unmount() error

	Rollback() error
	Merge() error
}
