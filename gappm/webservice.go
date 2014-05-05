package gappm

import (
	"fmt"
	"gabs.tv/pshub"
	"gabs.tv/pshubcl"
	"github.com/gabstv/gappm/gappm/webfiles"
	"github.com/go-martini/martini"
	"net/http"
	"time"
)

var (
	psServer *pshub.Server
)

func gappmApp(id string, name string) *pshub.ServerApp {
	app := pshub.NewApp(id, name)
	return app
}

func StartWS() error {
	psServer = pshub.NewServer(":5996")
	app0 := gappmApp("2222be5d-8491-44e1-b3f1-b1528b37fe94", "GAPPM")
	psServer.AddAdminKey("F349412")
	chan0 := app0.CreateChannel("gappm", "")
	chan0.ChanType = pshub.ChanT_Static
	psServer.AddApp(app0)
	return psServer.Run()
}

func StartHTML() {
	v := martini.Classic()
	v.Get("/index.html", func(w http.ResponseWriter) []byte {
		w.Header().Set("Content-Type", "text/html")
		return webfiles.V_index_html
	})
	v.Get("/", func(w http.ResponseWriter) []byte {
		w.Header().Set("Content-Type", "text/html")
		return webfiles.V_index_html
	})
	v.Get("/main.css", func(w http.ResponseWriter) []byte {
		w.Header().Set("Content-Type", "text/css")
		return webfiles.V_main_css
	})
	v.Get("/pshubcl.js", func(w http.ResponseWriter) []byte {
		w.Header().Set("Content-Type", "application/javascript")
		return webfiles.V_pshubcl_js
	})
	v.Get("/main.js", func(w http.ResponseWriter) []byte {
		w.Header().Set("Content-Type", "application/javascript")
		return webfiles.V_main_js
	})
	v.Options("/.*", func(r *http.Request, w http.ResponseWriter) {

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, HEAD, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Allow-Headers", "apikey, userkey, Apikey, Userkey, Content-Type")
		w.Header().Set("Access-Control-Max-Age", "1728000")
		w.WriteHeader(204)
		fmt.Println("OPTIONS!!!!!!")
	})
	fmt.Println(http.ListenAndServe(":9876", v))
}

func ClientConnect() {
	ok, errstr := pshubcl.Connect("http://localhost:5996", "2222be5d-8491-44e1-b3f1-b1528b37fe94")
	if !ok {
		fmt.Println("PSHUBCL ERROR", errstr)
		time.Sleep(time.Second * 10)
		ClientConnect()
		return
	}
	ok, errstr = pshubcl.SetAdminAuth("F349412")
	if !ok {
		fmt.Println("PSHUBCL ERROR [2]", errstr)
		time.Sleep(time.Minute)
		ClientConnect()
		return
	}
}

func Publish(msg string) (bool, string) {
	return pshubcl.Publish("gappm", msg)
}
