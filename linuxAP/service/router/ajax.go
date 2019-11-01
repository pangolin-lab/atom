package router

import (
	"net/http"
	"strings"
	"github.com/pangolin-lab/atom/linuxAP/service/controller"
	"reflect"
)

type AjaxRouter struct {
}

func (ar *AjaxRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	pathInfo := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(pathInfo, "/")

	var action = ""
	if len(parts) > 1 {
		for _, part := range parts[1:] {
			action += strings.Title(part)
		}
		action += "Do"
	}else{
		action += strings.Title("account") + "Do"
	}

	login := &controller.AjaxController{}
	cls := reflect.ValueOf(login)
	method := cls.MethodByName(action)
	if !method.IsValid() {
		method = cls.MethodByName(strings.Title("blank") + "Do")
	}
	requestValue := reflect.ValueOf(r)
	responseValue := reflect.ValueOf(w)
	method.Call([]reflect.Value{responseValue, requestValue})

}
