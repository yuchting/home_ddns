package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"home_ddns/config"
	"io/ioutil"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// DomainData
type DomainData struct {
	Domain string
	ID     string
}

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
		if strings.HasSuffix(v.Domain, do) {
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

func setCloudXNSHeader(request *http.Request, config config.HomeDDNSConfig) {
	dateStr := time.Now().Format(time.RFC1123Z)
	hashStr := config.CloudXNS_API_Key + request.URL.String() + dateStr + config.CloudXNS_API_Secret
	md5str := fmt.Sprintf("%x", md5.Sum([]byte(hashStr)))

	request.Header.Add("API-KEY", config.CloudXNS_API_Key)
	request.Header.Add("API-REQUEST-DATE", dateStr)
	request.Header.Add("API-HMAC", md5str)
}

func getCloudXNSDomainList(cfg config.HomeDDNSConfig) (domains []DomainData, err error) {
	const CloudXNSDomainListURL string = "https://www.cloudxns.net/api2/domain"
	request, err := http.NewRequest("GET", CloudXNSDomainListURL, nil)
	if err != nil {
		return nil, err
	}
	setCloudXNSHeader(request, cfg)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	// it show be return:
	//{
	//   "code":1,
	//   "message":"success",
	//   "total":"5",
	//   "data":[
	//      {
	//         "id":"1313",
	//         "domain":"chenqiao.com.",
	//         "status":"userlock",
	//         "take_over_status":"no",
	//         "level":"3",
	//         "create_time":"2014-09-17 09:45:46",
	//         "update_time":"2014-11-07 14:20:42",
	//         "ttl":"800"
	//      }
	//   ]
	//}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	jsonData := make(map[string]interface{})
	json.Unmarshal([]byte(content), &jsonData)

	if data, exist := jsonData["data"]; exist {
		if dataVal, ok := data.([]interface{}); ok {
			domains = make([]DomainData, 0, 5)
			for _, v := range dataVal {
				if domainData, ok := v.(map[string]interface{}); ok {
					d := DomainData{}
					d.Domain, _ = domainData["domain"].(string)
					d.ID, _ = domainData["id"].(string)

					domains = append(domains, d)
				}
			}
		} else {
			fmt.Printf("'data' is not exist, \n %s\n", content)
		}
	}

	return domains, nil
}

func main() {
	config := config.HomeDDNSConfig{}
	if err := config.Read("config.json"); err != nil {
		fmt.Println("error read config file, please create a json config file.")
		os.Exit(-1)
	}

	ip, err := getOwnIP()
	if err != nil {
		fmt.Println("error to get own IP ", err)
		os.Exit(-1)
	}

	fmt.Println("Got own IP: ", ip)

	domains, err := getCloudXNSDomainList(config)
	if err != nil {
		fmt.Println("error to get domain list ", err)
	}
	fmt.Println("Got own domains: ", domains)

	domainData := findDomain(domains, config.DDNS_Domain)
	if domainData == nil {
		os.Exit(-1)
	}

}
