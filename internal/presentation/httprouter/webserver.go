package httprouter

import (
	"fmt"
	"net/http"

	g "github.com/maragudk/gomponents"
	. "github.com/maragudk/gomponents/html"
)

const (
	backgroundColor = "#111827"
	textColor       = "white"

	tableHeaderBackgroundColor = "#374151"
	tableBackgroundColor       = "#1F2937"
)

var (
	colorStyleAttr            = StyleAttr(fmt.Sprintf("background-color: %s; color: %s;", backgroundColor, textColor))
	tableHeaderColorStyleAttr = StyleAttr(fmt.Sprintf("background-color: %s; color: %s;",
		tableHeaderBackgroundColor, textColor))
	tableColorStyleAttr = StyleAttr(fmt.Sprintf("background-color: %s; color: %s;", tableBackgroundColor, textColor))

	buttonClassRow = Class(`flex-wrap items-center py-2.5 px-5 mr-2 mb-2 text-sm font-medium rounded-lg border 
		focus:outline-none bg-gray-800 text-gray-100 border-gray-600 hover:text-white hover:bg-gray-700`)
)

func (router *HTTPRouter) initWebRoutes() {

	router.router.HandleFunc("/", router.createWebHandler(router.webIndex)).Methods("GET")
	router.router.HandleFunc("/search", router.createWebHandler(router.webIndex)).
		Methods("GET").Queries(
		"from", "{from}",
		"to", "{to}")
	router.router.HandleFunc("/login", router.createWebHandler(router.webLogin)).Methods("GET")
	router.router.HandleFunc("/stats", router.createWebHandler(router.webStats)).Methods("GET")
	router.router.HandleFunc("/admin", router.createWebHandler(router.webAdmin)).Methods("GET")
}

type pageHandlerFunc func(http.ResponseWriter, *http.Request) g.Node

func (router *HTTPRouter) createWebHandler(pageHandler pageHandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		body := pageHandler(w, r)
		info := getNavInfoByPath(r.URL.Path)
		_ = page(info.name, r.URL.Path, body).Render(w)
	}
}
