package controller

import (
	"net/http"
	"github.com/pangolin-lab/atom/linuxAP/golib"
	"log"
	"encoding/json"
)

type AjaxController struct {

}

func (ac *AjaxController)BlankDo(w http.ResponseWriter, r *http.Request)   {
	message := "Url: "

	message += r.URL.Path
	message += " not correct"

	w.Write([]byte(message))
}



func (ac *AjaxController)AccountDo(w http.ResponseWriter, r *http.Request)   {
	message := "Url: "

	message += r.URL.Path
	message += " not correct"

	w.Write([]byte(message))
}

type protonaccount struct {
	Address string
	CipherTxt string
}

func (ac *AjaxController)AccountCreateDo(w http.ResponseWriter, r *http.Request)   {
	pwds,ok:=r.URL.Query()["password"]
	if !ok || len(pwds)==0{
		w.WriteHeader(500)
		w.Write([]byte("{}"))
		return
	}

	pwd := pwds[0]

	pa:=&protonaccount{}
	pa.Address,pa.CipherTxt = golib.LibCreateAccount(pwd)

	bpa,err:=json.Marshal(*pa)
	if err!=nil{
		w.WriteHeader(500)
		w.Write([]byte("{}"))
		return
	}
	log.Println("create account:",string(bpa))

	w.WriteHeader(200)
	w.Write(bpa)
}
