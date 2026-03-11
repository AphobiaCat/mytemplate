package util

import (
	"encoding/json"
	"time"
)

func NowTimeS() int64 {
	now := time.Now()
	seconds := now.Unix()
	return seconds
}

func NowTimeMS() int64 {
	ms := time.Now().UnixMilli()
	return ms
}

func NowTimeUS() int64 {
	us := time.Now().UnixMicro()
	return us
}

func NowTimeNS() int64 {
	ns := time.Now().UnixNano()
	return ns
}

func Sleep(sleep_ms int) {
	time.Sleep(time.Duration(sleep_ms) * time.Millisecond)
}

func BuildJson(obj interface{}) string {

	jsonData, _ := json.Marshal(obj)

	return string(jsonData)
}

func ParserJson(message string, res interface{}) {

	/*
		type Parser_Json_Test_Data struct{
			Code	int		`json:"code"`
			Encode	uint32	`json:"encode"`
		}
	*/

	json.Unmarshal([]byte(message), res)
}
