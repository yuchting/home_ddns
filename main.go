package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
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

func ipDiff(cfg HomeDDNSConfig, ip string) {
	if cfg.IP_Path != "" {

		ipdata, err := ioutil.ReadFile(cfg.IP_Path)
		if err != nil && !os.IsNotExist(err) {
			fmt.Printf("read ip '%s'file error!\n %+v", cfg.IP_Path, err)
			return
		}

		formerIP := string(ipdata)

		if formerIP != ip {
			fmt.Printf("current ip <%s> is different from former<%s>, ", ip, formerIP)
			if cfg.IP_Diff_Shell != "" {
				fmt.Printf("execute '%s'\n", cfg.IP_Diff_Shell)
				cmd := exec.Command(cfg.IP_Diff_Shell)
				out, err := cmd.Output()
				if err != nil {
					fmt.Println(err)
				}
				fmt.Println(string(out))
			}

			ioutil.WriteFile(cfg.IP_Path, []byte(ip), 0666)
		}
	}
}

func main() {
	var jsonPath string
	var config = HomeDDNSConfig{}

	flag.StringVar(&jsonPath, "c", "", "json format config file")
	flag.StringVar(&config.CloudXNS_API_Key, "key", "", "[api key of CloudXNS], you can apply from https://www.cloudxns.net")
	flag.StringVar(&config.CloudXNS_API_Secret, "secret", "", "[api secret of CloudXNS], you can apply from https://www.cloudxns.net")
	flag.StringVar(&config.DDNS_Domain, "domain", "", "[domain name], write current ip address to a file")
	flag.StringVar(&config.IP_Path, "ip", "", "[filepath], write current ip address to a file")
	flag.StringVar(&config.IP_Diff_Shell, "ip-diff-sh", "", "[sh filepath/command], if the ip address is different with former ip which was writen into file, this shell/command will be called")

	flag.Parse()

	if jsonPath == "" {
		if config.CloudXNS_API_Key == "" || config.CloudXNS_API_Secret == "" || config.DDNS_Domain == "" {
			fmt.Println("miss required params! please check help note:")
			flag.PrintDefaults()
			os.Exit(-1)
		}
	}

	if jsonPath != "" {
		if err := config.Read(jsonPath); err != nil {
			fmt.Println("error read config file, please check.\n", err)
			os.Exit(-1)
		}
	}

	api := CloudXNSAPI{
		Config: config,
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
					fmt.Println("updated!")
				}
			} else {
				fmt.Println("same IP value in record, don't need set again.")
			}

			ipDiff(api.Config, ip)

			break
		}
	}

	if !foundHost {
		fmt.Printf("add host record '%s'...\n", api.Config.DDNS_Domain)
		if err = api.addDomainAAA(*domainData, hostName, ip); err != nil {
			fmt.Println("error : ", err)
		} else {
			ipDiff(api.Config, ip)
			fmt.Println("added!")
		}
	}
}
