package taxonomy

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Field struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Asset struct {
	Kind   string  `yaml:"kind"`
	Fields []Field `yaml:"fields"`
}

type Taxonomy struct {
	Assets []Asset `yaml:"assets"`
}

func Load(path string) (*Taxonomy, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var t Taxonomy
	if err := yaml.Unmarshal(data, &t); err != nil {
		return nil, err
	}
	return &t, nil
}
