package main

import (
	"encoding/json"
	"log"

	"github.com/birdkiller/apiroper"
)

func main() {
	testtmp := `{
	"getUserInfo":{
		"Output":{"staffinfo":"<<staffInfo.data.staffinfo>>",
				  "iminfo": "<<imUserInfo.data.userinfo>>"},
		"Resources":{
			"imUserInfo": {
				"Url":"http://192.168.8.111:8080/im/getUserInfo?mid=1&uid=<<login.data.User.Uid>>&token=<<login.data.SessionToken>>",
				"Header": {"Content-type": "application/json", 
						   "token":"<<login.data.SessionToken>>"},
				"Method": "POST",
				"Input": {"uid":"<<login.data.User.Uid>>"},
				"Timeout": 10
			},
			"staffInfo": {
				"Url": "http://192.168.8.111:8080/organization/GetCurPosAndPoslistByUid?mid=1&uid=<<login.data.User.Uid>>&token=<<login.data.SessionToken>>",
				"Header": {"Content-type": "application/json", 
						   "token":"<<login.data.SessionToken>>"},
				"Method": "POST",
				"Input": {"uid":"<<login.data.User.Uid>>"}
			},
			"login": {
				"Url": "http://192.168.8.111:8080/account/Login",
				"Header": {"Content-type": "application/json"},
				"Method": "POST",
				"Input": {"Id":"<<mobile>>",
							"IdType": 2,
							"UsePassword": true,
							"LoginToken":"123123",
							"AuthBotToken":""}
			}
		}
	}	
}`

	tmp := map[string]apiroper.Template{}
	json.Unmarshal([]byte(testtmp), &tmp)
	log.Println(tmp)

	apiroper.Load(tmp)
	args := map[string]interface{}{
		"mobile": "13655566676",
	}
	log.Println(apiroper.Call("getUserInfo", args))

}
