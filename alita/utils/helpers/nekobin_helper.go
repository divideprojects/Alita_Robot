package helpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type mapType map[string]interface{}

// PasteToNekoBin CreateTelegraphPost function used to create a Telegraph Page/Post with provide text
// We can use '<br>' inline text to split the messages into different paragraphs
func PasteToNekoBin(text string) (pasted bool, key string) {
	var body mapType

	if len(text) > 65000 {
		text = text[:65000]
	}
	postBody, err := json.Marshal(
		map[string]string{
			"content": text,
		},
	)
	if err != nil {
		log.Error(err)
	}

	responseBody := bytes.NewBuffer(postBody)
	resp, err := http.Post("https://nekobin.com/api/documents", "application/json", responseBody)
	if err != nil {
		log.Error(err)
		return
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Error(err)
		}
	}(resp.Body)

	err = json.NewDecoder(resp.Body).Decode(&body)
	if err != nil {
		log.Error(err)
		return
	}

	key = body["result"].(map[string]interface{})["key"].(string)
	if key != "" {
		return true, key
	}
	return
}
