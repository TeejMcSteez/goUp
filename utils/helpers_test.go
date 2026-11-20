package utils_test

import (
	"goUp/utils"
)

// loadConfig reads path through a fresh ConfigStore, the same way main does.
// Stand-in for the removed utils.LoadConfig.
func loadConfig(path string) (*utils.Config, error) {
	s := utils.NewConfigStore(path)
	if err := s.LoadFromFile(); err != nil {
		return nil, err
	}
	return s.Get(), nil
}

// apply runs a conf.go mutator on conf in place, then persists the result to
// conf.ConfigPath through a ConfigStore, the way handlers do via store.Update.
// Lets tests keep the "mutate, then reload from disk and check" shape.
func apply(conf *utils.Config, fn func(*utils.Config) (*utils.Config, error)) error {
	if _, err := fn(conf); err != nil {
		return err
	}
	s := utils.NewConfigStore(conf.ConfigPath)
	if err := s.LoadFromFile(); err != nil {
		return err
	}
	return s.Update(func(c *utils.Config) error {
		*c = *conf.Clone()
		return nil
	})
}
