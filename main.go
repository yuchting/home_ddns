package main

import (
	"crypto/md5"
	"fmt"
	"home_ddns/config"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"
)

func getOwnIP() (ip string) {
	request := &http.Client{
		Timeout: 5 * time.Second,
	}

	const GetOwnIPWebsite string = "http://ip.cn"
	response, err := request.Get(GetOwnIPWebsite)
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

	reg := regexp.MustCompile(`(\d{1-3}\.\d{1-3}\.\d{1-3}\.\d{1-3})`)

	content := (string)(body)
	ownIP := reg.FindStringSubmatch(content)
	if ownIP == nil || len(ownIP) != 1 {
		fmt.Printf("cannot parse the ip address from the website body:\n" + content)
		return
	}

	ip = ownIP[0]
	return
}

func setCloudXNSHeader(request *http.Request, config config.HomeDDNSConfig) {
	dataStr := time.Now().Format(time.RFC1123Z)
	md5str := fmt.Sprintf("%x", md5.Sum(config.CloudXNS_API_Key+request.URL.String()+dataStr+config.CloudXNS_API_Secret))

	request.Header.Add("API-KEY", config.CloudXNS_API_Key)
	request.Header.Add("API-REQUEST-DATE", dataStr)
	request.Header.Add("API-HMAC", md5str)
}

func getCloudXNSDomainList(config config.HomeDDNSConfig) (domains string, err error) {
	const CloudXNSDomainListURL string = "https://www.cloudxns.net/api2/domain"
	request, err := http.NewRequest("GET", CloudXNSDomainListURL, nil)
	if err != nil {
		return "", err
	}

	if err := setCloudXNSHeader(request, config); err != nil {
		return "", err
	}

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Do(reqest)
	if err != nil {
		return "", err
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
	domains = string(ioutil.ReadAll(response.Body))
	return
}

func main() {
	config := config.HomeDDNSConfig{}
	if err := config.Read("config.json"); err != nil {
		fmt.Println("error read config file, please create a config file.")
		return
	}

	ip := getOwnIP()

}
