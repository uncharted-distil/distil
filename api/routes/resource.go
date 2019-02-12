package routes

import (
	"io/ioutil"
	"path"
	//"goji.io/pat"
)

func fetchResourceBytes(resourceDir string, dataset string, file string) ([]byte, error) {
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
