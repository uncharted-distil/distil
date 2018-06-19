package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"goji.io/pat"
)

// ResourceHandler provides a static file lookup route using simple directory mapping.
func ResourceHandler(resourceDir string, proxy bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// resources can either be local or remote
		mediaFolder := pat.Param(r, "folder")
		file := pat.Param(r, "file")
		if proxy {
			proxyResourceHandler(resourceDir, mediaFolder, file).ServeHTTP(w, r)
		} else {
			http.FileServer(http.Dir(resourceDir)).ServeHTTP(w, r)
		}
	}
}

func proxyResourceHandler(server string, folder string, file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// create the URL based on the input
		url := fmt.Sprintf("%s/%s/%s", server, folder, file)
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			handleError(w, err)
		}

		// build http request
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			handleError(w, err)
		}
		defer res.Body.Close()

		// check status code
		if res.StatusCode >= 400 {
			body, err := ioutil.ReadAll(res.Body)
			if err != nil {
				handleError(w, err)
				return
			}
			handleError(w, fmt.Errorf(string(body)))
		}

		// return result directly
		bytes, err := ioutil.ReadAll(res.Body)
		if err != nil {
			handleError(w, err)
		}
		w.Write(bytes)
	}
}
