package main

import (
	"fmt"
	"home_ddns/config"
	"testing"
)

func TestGetOwnIp(t *testing.T) {
	ip, err := getOwnIP()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("own ip is : %+v \n", ip)
}

func getList(t *testing.T) (dd []DomainData) {
	config := config.HomeDDNSConfig{}
	err := config.Read("./config.json")
	if err != nil {
		t.Error(err)
	}

	dd, err = getCloudXNSDomainList(config)
	if err != nil {
		t.Error(err)
	}

	return
}

func TestGetCloudXNSDomainList(t *testing.T) {
	domains := getList(t)
	fmt.Printf("domains : %+v \n", domains)
}

func TestFindDomains(t *testing.T) {
	domains := make([]DomainData, 3, 10)
	domains[0].ID = "1"
	domains[0].Domain = "aaa.com"

	domains[1].ID = "2"
	domains[1].Domain = "bbb.com"

	domains[2].ID = "3"
	domains[2].Domain = "ccc.com"

	if findDomain(domains, "test.aaa.com") == nil {
		t.Error(fmt.Errorf("test error"))
	}

	if findDomain(domains, "aaa.com") != nil {
		t.Error(fmt.Errorf("test error"))
	}

	if findDomain(domains, "test.test.test.aaa.com") != nil {
		t.Error(fmt.Errorf("test error"))
	}
}
