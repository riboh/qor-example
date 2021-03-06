package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/csrf"
	"github.com/qor/qor-example/config"
	"github.com/qor/qor-example/config/admin"
	"github.com/qor/qor-example/config/admin/bindatafs"
	"github.com/qor/qor-example/config/api"
	_ "github.com/qor/qor-example/config/i18n"
	"github.com/qor/qor-example/config/routes"
	_ "github.com/qor/qor-example/db/migrations"
	"github.com/qor/qor/utils"
)

func main() {
	var compileTemplate = flag.Bool("compile-templates", false, "Compile Templates")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/", routes.Router())
	admin.Admin.MountTo("/admin", mux)

	api.API.MountTo("/api", mux)
	admin.Filebox.MountTo("/downloads", mux)

	mux.Handle("/system/", utils.FileServer(http.Dir("public")))

	assetFS := bindatafs.AssetFS.FileServer(http.Dir("public"), "javascripts", "stylesheets", "images")
	for _, path := range []string{"javascripts", "stylesheets", "images"} {
		mux.Handle(fmt.Sprintf("/%s/", path), assetFS)
	}

	skipCheck := func(h http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, "/auth") {
				r = csrf.UnsafeSkipCheck(r)
			}
			h.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
	handler := csrf.Protect([]byte("3693f371bf91487c99286a777811bd4e"), csrf.Secure(false))(mux)

	if *compileTemplate {
		bindatafs.AssetFS.Compile()
	} else {
		fmt.Printf("Listening on: %v\n", config.Config.Port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", config.Config.Port), skipCheck(handler)); err != nil {
			panic(err)
		}
	}
}
