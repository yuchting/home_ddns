package main

import (
	"bytes"
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

type DomainData struct {
	Domain string
	ID     string
}

type RecordData struct {
	Host     string
	DomainID string
	RecordID string
	HostID   string
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

func setCloudXNSHeader(request *http.Request, config config.HomeDDNSConfig, paramsBody []byte) {
	dateStr := time.Now().Format(time.RFC1123Z)

	hashStr := config.CloudXNS_API_Key + request.URL.String()
	if paramsBody != nil {
		hashStr += string(paramsBody)
	}
	hashStr += dateStr + config.CloudXNS_API_Secret

	md5str := fmt.Sprintf("%x", md5.Sum([]byte(hashStr)))

	request.Header.Add("API-KEY", config.CloudXNS_API_Key)
	request.Header.Add("API-REQUEST-DATE", dateStr)
	request.Header.Add("API-HMAC", md5str)
}

func getRequestJson(url string, cfg config.HomeDDNSConfig) (jsonContent map[string]interface{}, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	setCloudXNSHeader(request, cfg, nil)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	jsonData := make(map[string]interface{})
	if err = json.Unmarshal([]byte(content), &jsonData); err != nil {
		return nil, err
	}

	return jsonData, nil
}

// https://www.cloudxns.net/Support/detail/id/1361.html
func getCloudXNSDomainList(cfg config.HomeDDNSConfig) (domains []DomainData, err error) {

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

	jsonData, err := getRequestJson("https://www.cloudxns.net/api2/domain", cfg)
	if err != nil {
		return nil, err
	}

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
		}
	}

	return domains, nil
}

// https://www.cloudxns.net/Support/detail/id/1361.html
func getDomainRecords(cfg config.HomeDDNSConfig, domain DomainData) (records []RecordData, err error) {

	// {
	// 	"code": 1,
	// 	"message": "success",
	// 	"total": "2",
	// 	"offset": "0",
	// 	"row_num": "10",
	// 	"data": [{
	// 			"record_id": "31295",
	// 			"host_id": "12618",
	// 			"host": "1",
	// 			"line_zh": "\u5168\u7f51\u9ed8\u8ba4",
	// 			"line_en": "DEFAULT",
	// 			"line_id”:1,
	// 			"mx": null,
	// 			"value": "2.2.2.3",
	// 			"type": ”A”,
	// 			"status": "userstop",
	// 			"create_time": "2015-01-01 08:00:00",
	// 			"update_time": "2015-03-01 08:00:00"
	// 		},
	// 		{
	// 			"record_id": "31355",
	// 			"host_id": "12618",
	// 			"host": "1",
	// 			"line_zh": "\u5168\u7f51\u9ed8\u8ba4",
	// 			"line_en": "DEFAULT",
	// 			"line_id”:1,
	// 			"mx": null,
	// 			"value": "2.2.2.2",
	// 			"type": ”A”,
	// 			"status": "ok",
	// 			"create_time": "2015-01-01 08:00:00",
	// 			"update_time": "2015-03-01 08:00:00"
	// 		}
	// 	]
	// }

	url := fmt.Sprintf("https://www.cloudxns.net/api2/record/%s?host_id=0&offset=0&row_num=2000", domain.ID)
	jsonData, err := getRequestJson(url, cfg)
	if err != nil {
		return nil, err
	}

	if data, exist := jsonData["data"]; exist {
		if dataVal, ok := data.([]interface{}); ok {
			records = make([]RecordData, 0, 5)
			for _, v := range dataVal {
				if record, ok := v.(map[string]interface{}); ok {
					d := RecordData{}
					d.Host, _ = record["host"].(string)
					d.RecordID, _ = record["record_id"].(string)
					d.HostID, _ = record["host_id"].(string)
					d.DomainID = domain.ID

					records = append(records, d)
				}
			}
		}
	}

	return records, nil
}

func getPostPutJson(postOrPut bool, url string, params map[string]string, cfg config.HomeDDNSConfig) (map[string]interface{}, error) {

	jsonByte, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	method := "POST"
	if !postOrPut {
		method = "PUT"
	}

	request, err := http.NewRequest(method, url, bytes.NewBuffer(jsonByte))
	if err != nil {
		return nil, err
	}

	setCloudXNSHeader(request, cfg, jsonByte)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	jsonData := make(map[string]interface{})
	if err = json.Unmarshal([]byte(content), &jsonData); err != nil {
		return nil, err
	}

	return jsonData, nil
}

// https://www.cloudxns.net/Support/detail/id/1361.html
func updateDomainAAA(cfg config.HomeDDNSConfig, record RecordData, ip string) error {
	params := map[string]string{
		"domain_id": record.DomainID,
		"host":      record.Host,
		"value":     ip,
		"ttl":       "600",
		"type":      "A",
	}

	jsonData, err := getPostPutJson(false, "https://www.cloudxns.net/api2/record/"+record.RecordID, params, cfg)
	if err != nil {
		return err
	}

	// {
	// 	"code":1,
	// 	"message":" success",
	// 	"data":{
	// 		"id":63389,
	// 		"domain_name":"x.1s45test.com.",
	// 		"value":"9.2.4.3"
	// 	}
	// }

	if message, exist := jsonData["message"]; !exist || message != "success" {
		return fmt.Errorf("error response : %+v", jsonData)
	}

	return nil
}

func addDomainAAA(cfg config.HomeDDNSConfig) {

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

	records, err := getDomainRecords(config, *domainData)
	if err != nil {
		fmt.Println("error to get records ", err)
		os.Exit(-1)
	}

	hostName := config.DDNS_Domain[0:strings.Index(config.DDNS_Domain, ".")]
	foundHost := false
	for _, v := range records {
		if hostName == v.Host {
			foundHost = true
			fmt.Printf("update exist host '%s.%s' record as '%s'...\n", v.Host, domainData.Domain, ip)

			if err = updateDomainAAA(config, v, ip); err != nil {
				fmt.Println("error : ", err)
			}

			break
		}
	}

	if !foundHost {
		fmt.Printf("add host record '%s'...\n", config.DDNS_Domain)
	}
}
