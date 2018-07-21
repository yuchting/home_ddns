package main

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// CloudXNSAPI ...
type CloudXNSAPI struct {
	Config HomeDDNSConfig
}

type DomainData struct {
	Domain string
	ID     string
}

type RecordData struct {
	Host     string
	DomainID string
	RecordID string
	HostID   string
	Value    string
}

func (api CloudXNSAPI) setCloudXNSHeader(request *http.Request, paramsBody []byte) {
	dateStr := time.Now().Format(time.RFC1123Z)

	hashStr := api.Config.CloudXNS_API_Key + request.URL.String()
	if paramsBody != nil {
		hashStr += string(paramsBody)
	}
	hashStr += dateStr + api.Config.CloudXNS_API_Secret

	md5str := fmt.Sprintf("%x", md5.Sum([]byte(hashStr)))

	request.Header.Add("API-KEY", api.Config.CloudXNS_API_Key)
	request.Header.Add("API-REQUEST-DATE", dateStr)
	request.Header.Add("API-HMAC", md5str)
}

func (api CloudXNSAPI) getRequestJson(url string) (jsonContent map[string]interface{}, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	api.setCloudXNSHeader(request, nil)

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
func (api CloudXNSAPI) getCloudXNSDomainList() (domains []DomainData, err error) {

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

	jsonData, err := api.getRequestJson("https://www.cloudxns.net/api2/domain")
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
func (api CloudXNSAPI) getDomainRecords(domain DomainData) (records []RecordData, err error) {

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
	jsonData, err := api.getRequestJson(url)
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
					d.Value, _ = record["value"].(string)

					records = append(records, d)
				}
			}
		}
	}

	return records, nil
}

func (api CloudXNSAPI) getPostPutJson(postOrPut bool, url string, params map[string]interface{}) (map[string]interface{}, error) {

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

	api.setCloudXNSHeader(request, jsonByte)

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
func (api CloudXNSAPI) updateDomainAAA(record RecordData, ip string) error {
	params := map[string]interface{}{
		"domain_id": record.DomainID,
		"host":      record.Host,
		"value":     ip,
		"ttl":       "600",
		"type":      "A",
	}

	jsonData, err := api.getPostPutJson(false, "https://www.cloudxns.net/api2/record/"+record.RecordID, params)
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

func (api CloudXNSAPI) addDomainAAA(domainData DomainData, host string, ip string) error {
	// id, err := strconv.Atoi(domainData.ID)
	// if err != nil {
	// 	return fmt.Errorf("domain id is not integer %+v", err)
	// }

	params := map[string]interface{}{
		"domain_id": domainData.ID,
		"host":      host,
		"value":     ip,
		"line_id":   "1",
		"ttl":       "600",
		"type":      "A",
	}

	jsonData, err := api.getPostPutJson(true, "https://www.cloudxns.net/api2/record", params)
	if err != nil {
		return err
	}

	// {
	// 	"code":1,
	// 	"message":"success",
	// 	"record_id":[1234]
	// }

	if message, exist := jsonData["message"]; !exist || message != "success" {
		return fmt.Errorf("error response : %+v", jsonData)
	}

	return nil
}
