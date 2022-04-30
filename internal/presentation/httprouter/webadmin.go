package httprouter

import (
	"net/http"

	g "github.com/maragudk/gomponents"
)

func (router *HTTPRouter) webAdmin(w http.ResponseWriter, r *http.Request) g.Node {
	if u, _, _ := router.isAuthenticated(r); u == nil {
		return router.renderNotLoginGeneral()
	}

	return router.renderNotImplemeted()
}
