package routes

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"path"
	//"goji.io/pat"
)

func fetchResourceBytes(resourceDir string, proxyServer string, proxy map[string]bool, dataset string, file string) ([]byte, error) {
	if proxy[dataset] {
		return fetchRemoteResource(proxyServer, dataset, file)
	}
	return fetchLocalResource(resourceDir, file)
}

func fetchLocalResource(resourceDir string, file string) ([]byte, error) {
	// read the file locally
	filename := path.Join(resourceDir, file)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

func fetchRemoteResource(server string, dataset string, file string) ([]byte, error) {
	// create the URL based on the input
	url := fmt.Sprintf("%s/%s/%s", server, dataset, file)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// build http request
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	// check status code
	if res.StatusCode >= 400 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}
		return nil, fmt.Errorf(string(body))
	}

	// return result directly
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// // ResourceHandler provides a static file lookup route using simple directory mapping.
// func ResourceHandler(resourceDir string, proxyServer string, proxy map[string]bool) func(http.ResponseWriter, *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		// resources can either be local or remote
// 		dataset := pat.Param(r, "dataset")
// 		mediaFolder := pat.Param(r, "folder")
// 		file := pat.Param(r, "file")
//
// 		file = path.Join(mediaFolder, file)
//
// 		bytes, err := fetchResourceBytes(resourceDir, proxyServer, proxy, dataset, file)
// 		if err != nil {
// 			handleError(w, err)
// 			return
// 		}
// 		w.Write(bytes)
// 	}
// }
