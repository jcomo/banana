package banana

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Site struct {
	Title       string                 `yaml:"title"`
	Description string                 `yaml:"description"`
	Author      string                 `yaml:"author"`
	Vars        map[string]interface{} `yaml:"vars"`
}

type Config struct {
	Site Site `yaml:"site"`
}

func ReadConfig(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	defer f.Close()

	cfg := new(Config)
	err = yaml.NewDecoder(f).Decode(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
