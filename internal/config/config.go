package config

import (
	"encoding/json"
	"io/ioutil"
)

func SetConfig(path string, config interface{}) error {
	configBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	} else {
		if err = json.Unmarshal(configBytes, config); err != nil {
			return err
		}
	}
	return nil
}
