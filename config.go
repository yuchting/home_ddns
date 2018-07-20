package main

import (
	"encoding/json"
	"io/ioutil"
	"regexp"
)

// HomeDDNSConfig Config for CloudXNS settings
type HomeDDNSConfig struct {
	CloudXNS_API_Key    string
	CloudXNS_API_Secret string
	DDNS_Domain         string
}

func (config *HomeDDNSConfig) Read(filepath string) error {
	dat, err := ioutil.ReadFile(filepath)
	if err != nil {
		return err
	}

	reg, err := regexp.Compile("//.*")
	if err != nil {
		return err
	}

	content := reg.ReplaceAllString(string(dat), "")

	if err := json.Unmarshal([]byte(content), &config); err != nil {
		return err
	}
	return nil
}
