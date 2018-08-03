package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	"goji.io/pat"
)

// ResourceHandler provides a static file lookup route using simple directory mapping.
func ResourceHandler(resourceDir string, proxyServer string, proxy map[string]bool) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// resources can either be local or remote
		dataset := pat.Param(r, "dataset")
		mediaFolder := pat.Param(r, "folder")
		file := pat.Param(r, "file")

		file = path.Join(mediaFolder, file)
		if proxy[dataset] {
			proxyResourceHandler(proxyServer, dataset, file).ServeHTTP(w, r)
		} else {
			// resource directory should be the input directory
			localFileHandler(resourceDir, file).ServeHTTP(w, r)
		}
	}
}

func localFileHandler(resourceDir string, file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// read the file locally
		filename := path.Join(resourceDir, file)
		bytes, err := ioutil.ReadFile(filename)
		if err != nil {
			handleError(w, err)
		}
		w.Write(bytes)
	}
}

func proxyResourceHandler(server string, dataset string, file string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// create the URL based on the input
		url := fmt.Sprintf("%s/%s/%s", server, dataset, file)
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
