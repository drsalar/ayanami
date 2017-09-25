package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var accessToken string
var expire int
var lock *sync.RWMutex

var appid = "wxa7a0b1617b7559e2"
var secret = ""
var port = ""

type conf struct {
	Appid  string `json:"appid"`
	Secret string `json:"secret"`
	Port   string `json:"port"`
}

type access_token_return struct {
	Access_token string `json:"access_token"`
	Expires_in   int    `json:"expires_in"`
	Errcode      int    `json:"errcode"`
	Errmsg       string `json:"errmsg"`
}

func main() {
	Init()
	go runServer()
	for {
		time.Sleep(time.Second * time.Duration(expire))
		err := getAccessToken()
		if err != nil {
			log.Println("error:", err.Error())
			time.Sleep(5 * time.Second)
		} else {
			log.Println("update token successed!")
		}
	}
}

func Init() {
	log.SetOutput(os.Stdout)

	dir, err := os.Getwd()
	if err != nil {
		log.Panicln("error:", err.Error())
	}
	bytes, err := ioutil.ReadFile(dir + "/conf.json")
	if err != nil {
		log.Fatalln(err.Error())
	}

	var c conf
	err = json.Unmarshal(bytes, &c)
	if err != nil {
		log.Fatalln(err.Error())
	}
	appid = c.Appid
	secret = c.Secret
	port = c.Port

	expire = 0
	lock = new(sync.RWMutex)
	log.Println("Init conf successed!")
}

// successful return {"access_token":"ACCESS_TOKEN","expires_in":7200}
// {"errcode":40013,"errmsg":"invalid appid"}
func getAccessToken() error {
	url := "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=" + appid + "&secret=" + secret
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var v access_token_return
	err = json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	if v.Errcode != 0 {
		err = errors.New(string(data))
		return err
	}

	lock.Lock()
	accessToken = v.Access_token
	expire = v.Expires_in
	lock.Unlock()

	return nil
}

func runServer() {
	http.HandleFunc("/token", token)
	http.HandleFunc("/refresh", refresh)
	http.ListenAndServe(port, nil)
}

func token(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(accessToken))
	return
}

func refresh(w http.ResponseWriter, r *http.Request) {
	err := getAccessToken()
	if err != nil {
		log.Println("error:", err.Error())
	}
}
