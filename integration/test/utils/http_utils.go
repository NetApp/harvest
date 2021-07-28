package utils

import (
	"io/ioutil"
	"log"
	"net/http"
)

func GetResponse(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}
	sb := string(body)
	return sb, err
}
