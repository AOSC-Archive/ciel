package abstract

type Ciel interface {
	GetBasePath() string
	GetTree() Tree
	GetContainer() Container
}

type Container interface {
	GetBasePath() string
	DistDir() string
	GetCiel() Ciel
}

type Instance interface {
	MountPoint() string
	GetContainer() Container
}

type Tree interface {
	MountHandler(instance Instance, mount bool)
}
