package overlayfs

import (
	"os"
	"path"
)

func Create(layerPath string) error {
	var Layers = []string{
		"local",
		"diff",
	}
	if err := os.Mkdir(layerPath, 0755); err != nil {
		return err
	}
	for _, layer := range Layers {
		if err := os.Mkdir(path.Join(layerPath, layer), 0755); err != nil {
			return err
		}
	}
	return nil
}

func FromPath(basePath, layerPath string) *Instance {
	var Layers = []string{
		"local",
		"diff",
	}
	var layers = []string{basePath}
	for _, layer := range Layers {
		layers = append(layers, path.Join(layerPath, layer))
	}
	return &Instance{
		Layers: layers,
	}
}
