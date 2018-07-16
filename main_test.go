package main

import (
	"home_ddns/config"
	"testing"
)

func TestgetCloudXNSDomainList(t *testing.T) {
	config := config.HomeDDNSConfig{}
	err := config.Read("./config.json")
	if err != nil {
		t.Error(err)
	}

	t.Logf("domain body is : %+v \n", main.getCloudXNSDomainList(config))

}
