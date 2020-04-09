package dashboard

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/markbates/pkger"
)

// DashboardHandler expose dashboard routes
type DashboardHandler struct {
	Assets *pkger.Dir
}

// Append add dashboard routes on a router
func (g DashboardHandler) Append(router *mux.Router) {
	if g.Assets == nil {
		log.Printf("No assets for dashboard")
		return
	}

	// Expose dashboard
	router.Methods(http.MethodGet).
		Path("/").
		HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
			http.Redirect(response, request, request.Header.Get("X-Forwarded-Prefix")+"/dashboard/", http.StatusFound)
		})

	router.Methods(http.MethodGet).
		PathPrefix("/dashboard/").
		Handler(http.StripPrefix("/dashboard/", http.FileServer(g.Assets)))
}
