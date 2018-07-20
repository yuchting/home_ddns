package main

import (
	"fmt"
	"testing"
)

func TestGetOwnIp(t *testing.T) {
	ip, err := getOwnIP()
	if err != nil {
		t.Error(err)
	}

	fmt.Printf("own ip is : %+v \n", ip)
}

func getList(t *testing.T) (dd []DomainData, api CloudXNSAPI) {
	api = CloudXNSAPI{
		Config: HomeDDNSConfig{},
	}

	err := api.Config.Read("./config.json")
	if err != nil {
		t.Error(err)
	}

	dd, err = api.getCloudXNSDomainList()
	if err != nil {
		t.Error(err)
	}

	return dd, api
}

func TestGetCloudXNSDomainList(t *testing.T) {
	domains, _ := getList(t)
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

func TestGetDomainRecords(t *testing.T) {
	domains, api := getList(t)
	for _, v := range domains {
		records, err := api.getDomainRecords(v)
		if err != nil {
			t.Error(err)
			break
		}

		fmt.Printf("get %s records:\n", v.Domain)

		for _, re := range records {
			fmt.Println(re.Host)
		}
	}
}
