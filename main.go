package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

func findDomain(domains []DomainData, do string) (dd *DomainData) {
	fmt.Printf("finding %s in your domain list\n", do)
	dotIdx := strings.Index(do, ".")

	if dotIdx == -1 {
		fmt.Printf("%s is not domain!\n", do)
		return nil
	}

	do = do[(dotIdx + 1):]

	if !regexp.MustCompile(`^[\w\d]+\.[\w\d]+$`).Match([]byte(do)) {
		fmt.Printf("%s is not domain!\n", do)
		return nil
	}

	for _, v := range domains {
		if strings.HasSuffix(v.Domain, do+".") {
			return &v
		}
	}

	fmt.Printf("%s cannot be found in you own domains\n", do)
	return nil
}

func getOwnIP() (ip string, err error) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	const GetOwnIPWebsite string = "http://ip.cn"
	request, err := http.NewRequest("GET", GetOwnIPWebsite, nil)
	if err != nil {
		return "", err
	}

	request.Header.Add("User-Agent", "curl/7.29.0")

	response, err := client.Do(request)
	if err != nil {
		return "", err
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return "", err
	}

	reg := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)

	content := (string)(body)
	ownIP := reg.FindStringSubmatch(content)
	if ownIP == nil || len(ownIP) == 0 {
		return "", fmt.Errorf("cannot parse the ip address from the website body:\n" + content)
	}

	return ownIP[0], nil
}

func main() {

	const helpNote string = `
-c <json format config file>
or
--key <api key of CloudXNS> 
    you can apply from https://www.cloudxns.net
--secret <api secret of CloudXNS> 
    you can apply from https://www.cloudxns.net
--domain <domain name>
    'homeddns.example.com' or 'ddns.mydomain.com'
	`

	var key string
	var secret string
	var domain string
	var configFile string

	args := os.Args[1:]
	for i, v := range args {
		switch v {
		case "--key":
			key = args[i+1]
		case "--secret":
			secret = args[i+1]
		case "--domain":
			domain = args[i+1]
		case "-c":
			configFile = args[i+1]
		}
	}

	if configFile == "" {
		if key == "" || secret == "" || domain == "" {
			fmt.Println("miss required params! please check help note:\n", helpNote)
			return
		}
	}

	api := CloudXNSAPI{
		Config: HomeDDNSConfig{
			CloudXNS_API_Key:    key,
			CloudXNS_API_Secret: secret,
			DDNS_Domain:         domain,
		},
	}

	if configFile != "" {
		if err := api.Config.Read(configFile); err != nil {
			fmt.Println("error read config file, please check.\n", err)
			os.Exit(-1)
		}
	}

	ip, err := getOwnIP()
	if err != nil {
		fmt.Println("error to get own IP ", err)
		os.Exit(-1)
	}

	fmt.Println("Got own IP: ", ip)

	domains, err := api.getCloudXNSDomainList()
	if err != nil {
		fmt.Println("error to get domain list ", err)
	}
	fmt.Println("Got own domains: ", domains)

	domainData := findDomain(domains, api.Config.DDNS_Domain)
	if domainData == nil {
		os.Exit(-1)
	}

	records, err := api.getDomainRecords(*domainData)
	if err != nil {
		fmt.Println("error to get records ", err)
		os.Exit(-1)
	}

	hostName := api.Config.DDNS_Domain[0:strings.Index(api.Config.DDNS_Domain, ".")]
	foundHost := false
	for _, v := range records {
		if hostName == v.Host {
			foundHost = true

			if ip != v.Value {
				fmt.Printf("update exist host '%s.%s' A record as '%s'...\n", v.Host, domainData.Domain, ip)

				if err = api.updateDomainAAA(v, ip); err != nil {
					fmt.Println("error : ", err)
				} else {
					fmt.Println("Successfull!")
				}
			} else {
				fmt.Println("same IP value in record, don't need set again.")
			}

			break
		}
	}

	if !foundHost {
		fmt.Printf("add host record '%s'...\n", api.Config.DDNS_Domain)
		if err = api.addDomainAAA(*domainData, hostName, ip); err != nil {
			fmt.Println("error : ", err)
		} else {
			fmt.Println("Successfull!")
		}
	}
}
