package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	wr "github.com/stuwilli/go-web-response"
)

func getFileList(loc string) []string {

	list := []string{}

	err := filepath.Walk(loc, func(path string, f os.FileInfo,
		err error) error {

		if !f.IsDir() && filepath.Ext(path) == ".json" {

			list = append(list, path)
		}

		return nil
	})

	if err != nil {
		log.Fatal(err.Error())
	}

	return list
}

//ReadServiceConfigs ...
func ReadServiceConfigs(loc string) []ServiceConfig {

	configs := []ServiceConfig{}

	list := getFileList(loc)

	for _, f := range list {

		content, err := ioutil.ReadFile(f)

		if err != nil {
			log.Fatal(err)
		}
		sc := ServiceConfig{}
		err = json.Unmarshal(content, &sc)

		if err != nil {
			fmt.Println("Dropping", f, " it is invalid", err)
			continue
		}

		configs = append(configs, sc)

	}

	return configs
}

func main() {

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.StripSlashes)

	configLoc := "./test_configs"

	if cf, ok := os.LookupEnv("CONFIG_LOCATION"); ok {
		configLoc = cf
	}

	configs := ReadServiceConfigs(configLoc)

	for _, c := range configs {
		c.CreateProxy(r)
	}

	// r.HandleFunc("/backend*", func(w http.ResponseWriter, r *http.Request) {
	//
	// 	str, _ := httputil.DumpRequest(r, false)
	// 	fmt.Printf("%s", str)
	// 	w.WriteHeader(200)
	// 	w.Write(([]byte("CHEESE")))
	// })

	r.NotFound(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		resp := wr.NewBuilder().Status(http.StatusNotFound).Build()
		resp.WriteJSON(w)
	}))

	r.MethodNotAllowed(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		resp := wr.NewBuilder().Status(http.StatusMethodNotAllowed).Build()
		resp.WriteJSON(w)
	}))

	http.ListenAndServe(":8080", r)
}
