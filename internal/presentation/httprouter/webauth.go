package httprouter

import (
	"fmt"
	"net/http"

	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
	"github.com/n-r-w/log-server/internal/domain/model"
)

const (
	// JS для логина
	loginJS = `
function loginREST()
{ 
	var login = document.getElementById("login").value
	var password = document.getElementById("password").value

	var json = JSON.stringify({
	  login: login,
	  password: password
	});

	var xhr = new XMLHttpRequest()	
	// xhr.withCredentials = true	
	xhr.onreadystatechange = function () {
	    if (xhr.readyState === 4) {			
			if (xhr.status == 200) {
	            console.log("authenticated OK")				
			} else {
				console.log("authentication ERROR:", xhr.responseText)
			}
			location.reload();
	    }		
	}
	
	xhr.open("post", "/api/auth/login", true);
	xhr.setRequestHeader("Content-Type", 'application/json');
	xhr.send(json);
}`
	// JS для логаута
	logoutJS = `
function logoutREST()
{ 	
	var xhr = new XMLHttpRequest();	
	xhr.onreadystatechange = function () {
	    if (xhr.readyState === 4) {			
			if (xhr.status == 200) {
	            console.log("session closed OK")				
			} else {
				console.log("session close ERROR:", xhr.responseText)
			}
			location.reload();
	    }		
	}
	
	xhr.open("delete", "/api/auth/close", true);	
	xhr.send();
}`
)

var buttonClassRowNewLineSmall = Class(`flex-wrap space-x-44 items-center py-2.5 px-5 mr-2 mb-3 text-sm font-medium 
	rounded-lg border focus:outline-none bg-gray-800 text-gray-100 border-gray-600 hover:text-white hover:bg-gray-700`)
var buttonClassRowNewLineBig = Class(`flex-wrap space-x-44 items-center py-2.5 px-5 mr-2 mb-5 text-sm font-medium 
	rounded-lg border focus:outline-none bg-gray-800 text-gray-100 border-gray-600 hover:text-white hover:bg-gray-700`)
var buttonClassRowSameLine = Class(`calender-black flex-wrap items-center py-2.5 px-5 ml-2 mr-2 mb-2 text-sm font-medium 
	rounded-lg border focus:outline-none bg-gray-800 text-gray-100 border-gray-600 hover:text-white hover:bg-gray-700`)
var textClassRowSameLine = Class(`flex-wrap items-center py-2.5 px-0 ml-0 mr-0 mb-2 text-sm font-medium 
	focus:outline-none text-gray-100`)

func (router *HTTPRouter) webLogin(w http.ResponseWriter, r *http.Request) g.Node {
	user, httpCode, err := router.isAuthenticated(r)
	if err != nil {
		return router.renderLoginNo(w, r, httpCode, err)
	}

	return router.renderLoginOK(w, r, user)
}

func (router *HTTPRouter) renderLoginNo(w http.ResponseWriter, r *http.Request, httpCode int, err error) g.Node {
	var message string
	switch httpCode {
	case http.StatusUnauthorized:
		message = "Вход не выполнен"
	case http.StatusInternalServerError:
		message = fmt.Sprintf("Ошибка сервера: %v", err)
	case http.StatusNotFound:
		message = "Пользователь не существует"
	default:
		message = err.Error()
	}

	body := FormEl(
		Div(g.Text("Логин")),
		Div(Input(buttonClassRowNewLineSmall, ID("login"), Type("text"))),
		Div(g.Text("Пароль")),
		Div(Input(buttonClassRowNewLineBig, ID("password"), Type("password"))),
		Input(buttonClassRowNewLineSmall,
			ID("loginButton"), Type("button"), Value("Войти"), g.Attr("onclick", "loginREST()")),
		Div(g.Text(message)),
		Script(g.Raw(loginJS)))
	return body
}

func (router *HTTPRouter) renderLoginOK(w http.ResponseWriter, r *http.Request, user *model.User) g.Node {
	body := FormEl(
		Div(Class("py-2.5 px-0 mr-2 mb-2 text-sm"), g.Text("Привет "), B(g.Text(user.Name)), g.Text("!")),
		Input(buttonClassRowNewLineSmall,
			ID("logoutButton"), Type("button"), Value("Выйти"), g.Attr("onclick", "logoutREST()")),
		Script(g.Raw(logoutJS)))
	return body
}
