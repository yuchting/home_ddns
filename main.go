package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"home_ddns/config"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

func getOwnIP() (ip string) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	const GetOwnIPWebsite string = "http://ip.cn"
	request, err := http.NewRequest("GET", GetOwnIPWebsite, nil)
	if err != nil {
		fmt.Println("request "+GetOwnIPWebsite+" failed!\n ", err)
		return
	}

	request.Header.Add("User-Agent", "curl/7.29.0")

	response, err := client.Do(request)
	if err != nil {
		fmt.Println("request "+GetOwnIPWebsite+" failed!\n ", err)
		return
	}

	defer response.Body.Close()
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("read body error !\n ", err)
		return
	}

	reg := regexp.MustCompile(`(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})`)

	content := (string)(body)
	ownIP := reg.FindStringSubmatch(content)
	if ownIP == nil || len(ownIP) == 0 {
		fmt.Printf("cannot parse the ip address from the website body:\n" + content)
		return
	}

	return ownIP[0]
}

func setCloudXNSHeader(request *http.Request, config config.HomeDDNSConfig) {
	dateStr := time.Now().Format(time.RFC1123Z)
	hashStr := config.CloudXNS_API_Key + request.URL.String() + dateStr + config.CloudXNS_API_Secret
	md5str := fmt.Sprintf("%x", md5.Sum([]byte(hashStr)))

	request.Header.Add("API-KEY", config.CloudXNS_API_Key)
	request.Header.Add("API-REQUEST-DATE", dateStr)
	request.Header.Add("API-HMAC", md5str)
}

func getCloudXNSDomainList(cfg config.HomeDDNSConfig) (domains []string, err error) {
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
		if dataVal, ok := data.([]map[string]interface{}); ok {
			for i, v := range dataVal {
				if _, exist := v["domain"]; exist {
					if domain, ok := v["domain"].(string); ok {
						domains[i] = domain
					} else {
						fmt.Printf("this 'domain' is not string\n %s\n", content)
					}
				} else {
					fmt.Printf("'domain' is not exist, \n %s\n", content)
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
		fmt.Println("error read config file, please create a config file.")
		return
	}

	ip := getOwnIP()
	fmt.Println("Got own IP: ", ip)
}
