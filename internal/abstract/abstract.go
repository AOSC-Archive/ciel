package abstract

type Ciel interface {
	GetBasePath() string
}

type Container interface {
	GetBasePath() string
	DistDir() string
	GetCiel() Ciel
}
