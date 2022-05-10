package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var tokenGlobal = ""
var statusCode = 1

// slice.. map ... channel .. reference types

type timeChecker struct {
	token  string
	expire time.Time
}

var timeNow = time.Now().UTC()             // all times utc
var timeMaster = &timeChecker{"", timeNow} // timeNow.. one time use... intialize the current time

func main() {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	getSetToken()
	fmt.Println("starting server... port 8090.... /delete func pass msgid and rcpt values in string format")

	http.HandleFunc("/delete", delete)

	//http.HandleFunc("/code1", code1)

	http.ListenAndServe(":8090", nil)
}

func getSetToken() {
	endpoint := "https://login.microsoftonline.com/2bac89e1-0311-4f18-bbb5-a229530a794a/oauth2/v2.0/token"

	data := url.Values{}
	data.Set("tenant", "2bac89e1-0311-4f18-bbb5-a229530a794a")
	data.Set("client_id", "4ce551d1-f7d1-479f-85f7-236d9c102fbb")
	data.Set("scope", "https://graph.microsoft.com/.default")
	data.Set("client_secret", "0GWkpCyOOE~gw9~0E7FptOykogX~yd3~z-")
	data.Set("grant_type", "client_credentials")

	client := &http.Client{}
	r, err := http.NewRequest("POST", endpoint, strings.NewReader(data.Encode())) // URL-encoded payload
	if err != nil {
		log.Fatal(err)
	}
	r.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	r.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))

	res, err := client.Do(r)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(res.Status)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	// log.Println(string(body))
	mappy1 := map[string]interface{}{}
	json.Unmarshal(body, &mappy1)
	fmt.Println(mappy1["access_token"], reflect.TypeOf(mappy1["access_token"]))
	fmt.Println("expires.....", mappy1["expires_in"], reflect.TypeOf(mappy1["expires_in"]))
	fmt.Println(mappy1["token_type"])

	access_token := mappy1["access_token"].(string)

	t1 := time.Now().UTC()
	t1 = t1.Add(time.Minute * 59)
	fmt.Println("time now Local", time.Now())
	fmt.Println("t1 .....UTC ...  ", time.Now().UTC())
	fmt.Println("t1 + 59 minutes..... ", t1)
	timeMaster = &timeChecker{access_token, t1}
	fmt.Println("timeMaster set..........", timeMaster, reflect.TypeOf(timeMaster))

	tokenGlobal = access_token
	fmt.Println("*********************************************")
	fmt.Println(string(body))
}

var param_rcpt = ""

func delete(w http.ResponseWriter, r *http.Request) {
	//http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	timeNowCheck := time.Now().UTC()
	if timeNowCheck.After(timeMaster.expire) {
		getSetToken()
	}
	query := r.URL.Query()
	param_msgid := query.Get("msgid")
	param_rcpt = query.Get("rcpt")

	param_msgid = strings.Replace(param_msgid, " ", "+", -1) // <--------------encoding issue

	fmt.Println(param_msgid, param_rcpt)

	msgid3 := param_msgid
	// filter=internetMessageId = '<messageid here>'&select=subject,id

	//custom_url := "https://graph.microsoft.com/v1.0/users/" + param_rcpt + "/messages?$filter=internetMessageId%20eq%20%27" + msgid3 + "%27" + "&$select=subject,id"

	custom_url := "https://graph.microsoft.com/v1.0/users/" + param_rcpt + "/messages?$filter=internetMessageId%20eq%20%27" + msgid3 + "%27"

	fmt.Println("URL:", custom_url)

	client := &http.Client{}
	req, err := http.NewRequest("GET", custom_url, nil)

	if err != nil {
		fmt.Print(err)
	}

	req.Header.Add("Authorization", `Bearer `+tokenGlobal)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}
	data_resp, _ := ioutil.ReadAll(resp.Body)
	data2 := data_resp

	fmt.Println("level 1:------------------------------", string(data2))
	fmt.Println("---------------------------------->")

	mappy1 := map[string]interface{}{}

	json.Unmarshal(data2, &mappy1)

	p100 := ""

	fmt.Println("mappy1......", mappy1)

	if len(mappy1["value"].([]interface{})) == 0 {
		fmt.Println("NO Values...Cannot locate messageID !!!!!! ")
		responsecode1 := map[string]string{"error code": "No data found", "message 1": "said again nada.."}
		fmt.Println(responsecode1)
		json.NewEncoder(w).Encode(responsecode1)

		//os.Exit(3)
	} else {

		mappysummary := map[string]interface{}{}

		p100 = mappy1["value"].([]interface{})[0].(map[string]interface{})["id"].(string) // this is the microsoft email id

		mappy2 := getSummary(p100)

		fmt.Println("mappy2:.............................................", mappy2)

		p_messageid := mappy2["internetMessageId"]
		p_isread := mappy2["isRead"]
		p_recipients := mappy2["toRecipients"]
		p_sender := mappy2["sender"]

		mappysummary["id"] = p100
		mappysummary["messageID"] = p_messageid
		mappysummary["isread"] = p_isread
		mappysummary["recipients"] = p_recipients
		mappysummary["sender"] = p_sender

		fmt.Println(p100)

		//DELETE MESSAGE Uncomment ....

		//return_data := executeDelete(p100)

		//fmt.Println("returned data..... ", return_data, reflect.TypeOf(return_data))
		//fmt.Fprintf(w, return_data)
		fmt.Println(mappysummary)

		mappy2["deleteStatusCode"] = statusCode

		json.NewEncoder(w).Encode(mappy2)
	}
}

func getSummary(id string) map[string]interface{} {
	fmt.Println("getSumarry() ........................")
	var messageid = id

	var custom_url = `https://graph.microsoft.com/v1.0/users/` + param_rcpt + `/messages/` + messageid // special microsoft email id
	client := &http.Client{}
	req, err := http.NewRequest("GET", custom_url, nil)

	if err != nil {
		fmt.Print(err)
	}

	req.Header.Add("Authorization", `Bearer `+tokenGlobal)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}
	data_resp, _ := ioutil.ReadAll(resp.Body)
	data2 := data_resp
	//
	fmt.Println("data2 output:------------------------------------------------", string(data2))
	fmt.Println("--------------------------->>")
	mappy1 := map[string]interface{}{}
	json.Unmarshal(data_resp, &mappy1)

	//mappy11:=mappy1.(map[string]interface{})
	//delete(mappy11,"body")

	return mappy1
}

// ---------------------------------------------------------------------------------------------
// ---------------------------------------------------------------------------------------------
// ------------------------NOT CALLED----------------------when executed deletes MESSAGE !!!!!!

func executeDelete(p100 string) string {
	msgid3 := p100

	client := &http.Client{}
	custom_url := "https://graph.microsoft.com/v1.0/users/" + param_rcpt + "/messages/" + msgid3

	req, err := http.NewRequest("DELETE", custom_url, nil)

	if err != nil {
		fmt.Print(err)
	}

	req.Header.Add("Authorization", `Bearer `+tokenGlobal)
	resp, err := client.Do(req)
	if err != nil {
		fmt.Print(err)
	}

	fmt.Println("status code:", resp.StatusCode)
	data_resp, _ := ioutil.ReadAll(resp.Body)
	data2 := data_resp

	fmt.Println("level 2:-------------------", string(data2))
	fmt.Println("----------------------------------------")
	statusCode = resp.StatusCode
	fmt.Println(statusCode, reflect.TypeOf(statusCode))

	mappyReturn := map[string]interface{}{}
	mappyReturn["statusCode"] = resp.StatusCode
	mappyReturn["messageID"] = msgid3

	mappyReturnString, _ := json.Marshal(mappyReturn)
	var prettyJSON bytes.Buffer
	err = json.Indent(&prettyJSON, data2, "", "\t")
	return_data := string(prettyJSON.Bytes())
	fmt.Println("first.. return data... ", return_data)

	fmt.Println("statusCode:", statusCode)
	return string(mappyReturnString)
}
