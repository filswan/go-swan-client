package utils

import (
	"bytes"
	"encoding/json"
	"go-swan-client/logs"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
)

const HTTP_CONTENT_TYPE_FORM = "application/x-www-form-urlencoded"
const HTTP_CONTENT_TYPE_JSON = "application/json; charset=utf-8"

func HttpPostNoToken(uri string, params interface{}) string {
	response := httpRequest(http.MethodPost, uri, "", params)
	return response
}

func HttpPost(uri, tokenString string, params interface{}) string {
	response := httpRequest(http.MethodPost, uri, tokenString, params)
	return response
}

func HttpGetNoToken(uri string, params interface{}) string {
	response := httpRequest(http.MethodGet, uri, "", params)
	return response
}

func HttpGet(uri, tokenString string, params interface{}) string {
	response := httpRequest(http.MethodGet, uri, tokenString, params)
	return response
}

func HttpPut(uri, tokenString string, params interface{}) string {
	response := httpRequest(http.MethodPut, uri, tokenString, params)
	return response
}

func HttpDelete(uri, tokenString string, params interface{}) string {
	response := httpRequest(http.MethodDelete, uri, tokenString, params)
	return response
}

func httpRequest(httpMethod, uri, tokenString string, params interface{}) string {
	var request *http.Request
	var err error

	switch params := params.(type) {
	case io.Reader:
		request, err = http.NewRequest(httpMethod, uri, params)
		if err != nil {
			logs.GetLogger().Error(err)
			return ""
		}
		request.Header.Set("Content-Type", HTTP_CONTENT_TYPE_FORM)
	default:
		jsonReq, errJson := json.Marshal(params)
		if errJson != nil {
			logs.GetLogger().Error(errJson)
			return ""
		}

		request, err = http.NewRequest(httpMethod, uri, bytes.NewBuffer(jsonReq))
		if err != nil {
			logs.GetLogger().Error(err)
			return ""
		}
		request.Header.Set("Content-Type", HTTP_CONTENT_TYPE_JSON)
	}

	if len(strings.Trim(tokenString, " ")) > 0 {
		request.Header.Set("Authorization", "Bearer "+tokenString)
	}

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.GetLogger().Error(err)
		return ""
	}

	return string(responseBody)
}

func HttpPostFile(url string, paramTexts map[string]interface{}, paramFiles ...string) (*string, error) {
	bodyBuf := &bytes.Buffer{}
	bodyWriter := multipart.NewWriter(bodyBuf)
	defer bodyWriter.Close()

	for k, v := range paramTexts {
		bodyWriter.WriteField(k, v.(string))
	}

	for _, paramFile := range paramFiles {
		paramFileStat, err := os.Stat(paramFile)
		if err != nil {
			logs.GetLogger().Error(err)
			return nil, err
		}

		fileWriter, err := bodyWriter.CreateFormFile(paramFileStat.Mode().Type().String(), paramFileStat.Name())
		if err != nil {
			logs.GetLogger().Info(err)
			return nil, err
		}

		data, err := ReadFile(paramFile)
		if err != nil {
			logs.GetLogger().Info(err)
			return nil, err
		}

		fileWriter.Write(data)
	}

	contentType := bodyWriter.FormDataContentType()

	response, err := http.Post(url, contentType, bodyBuf)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.GetLogger().Error(err)
		return nil, err
	}

	responseStr := string(responseBody)

	return &responseStr, nil
}
