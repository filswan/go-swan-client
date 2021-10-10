package client

import (
	"bytes"
	"encoding/json"
	"go-swan-client/common/utils"
	"go-swan-client/logs"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
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

func HttpPostFile(url string, tokenString string, paramTexts map[string]string, paramFilename, paramFilepath string) (string, error) {
	filename, fileContent, err := utils.ReadFile(paramFilepath)
	if err != nil {
		logs.GetLogger().Info(err)
		return "", err
	}

	bodyBuf := new(bytes.Buffer)
	bodyWriter := multipart.NewWriter(bodyBuf)

	fileWriter, err := bodyWriter.CreateFormFile(paramFilename, filename)
	if err != nil {
		logs.GetLogger().Info(err)
		return "", err
	}

	fileWriter.Write(fileContent)

	for key, val := range paramTexts {
		err = bodyWriter.WriteField(key, val)
		if err != nil {
			logs.GetLogger().Info(err)
			return "", err
		}
	}

	bodyWriter.Close()

	request, err := http.NewRequest(http.MethodPost, url, bodyBuf)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", nil
	}

	request.Header.Set("Content-Type", bodyWriter.FormDataContentType())
	if len(strings.Trim(tokenString, " ")) > 0 {
		request.Header.Set("Authorization", "Bearer "+tokenString)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", nil
	}

	defer response.Body.Close()

	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		logs.GetLogger().Error(err)
		return "", err
	}

	responseStr := string(responseBody)

	return responseStr, nil
}
