package main

import (
	"home_ddns/config"
	"testing"
)

func TestGetOwnIp(t *testing.T) {
	t.Logf("own ip is : %+v \n", getOwnIP())
}

func TestGetCloudXNSDomainList(t *testing.T) {
	config := config.HomeDDNSConfig{}
	err := config.Read("./config.json")
	if err != nil {
		t.Error(err)
	}

	domains, err := getCloudXNSDomainList(config)
	if err != nil {
		t.Error(err)
	}

	t.Logf("domains : %+v \n", domains)
}
