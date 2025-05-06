package aijobrecommender

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
)

func JobrecommenderSetup(config Config) *Jobrecommender {

	return &Jobrecommender{
		DB: config.DB,
	}

}
// function ResumeAnalyze
func (jobrecommender *Jobrecommender) ResumeAnalyze(file *multipart.FileHeader, url string) (responsedata map[string]interface{}, statuscode string, err error) {

	if file == nil {
		return nil, "", fmt.Errorf("no file provided")
	}

	fileContent, err := file.Open()
	if err != nil {
		return nil, "", fmt.Errorf("unable to open uploaded file: %w", err)
	}
	defer fileContent.Close()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", file.Filename)
	if err != nil {
		return nil, "", fmt.Errorf("unable to create file part: %w", err)
	}

	if _, err := io.Copy(part, fileContent); err != nil {
		return nil, "", fmt.Errorf("error copying file content: %w", err)
	}
	writer.Close()

	baseURL := url

	resp, err := http.Post(baseURL, writer.FormDataContentType(), body)
	if err != nil {
		return nil, "", fmt.Errorf("failed to call analyse API: %w", err)
	}
	defer resp.Body.Close()
	fmt.Println("respresp", resp.Body)
	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, resp.Status, fmt.Errorf("failed to decode response: %w", err)
	}

	return responseData, resp.Status, nil

}


func (jobrecommender *Jobrecommender) Jobrecommendation(req JobRecommendationRequest) (JsonData, error) {

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	jsonPart, err := writer.CreateFormFile("files", filepath.Base(req.JSONFile.Name()))
	if err != nil {
		return JsonData{}, fmt.Errorf("failed to create form file for JSON: %w", err)
	}
	if _, err := req.JSONFile.Seek(0, io.SeekStart); err != nil {
		return JsonData{}, fmt.Errorf("failed to seek JSON file: %w", err)
	}
	if _, err := io.Copy(jsonPart, req.JSONFile); err != nil {
		return JsonData{}, fmt.Errorf("failed to write JSON content: %w", err)
	}

	resumePart, err := writer.CreateFormFile("files", filepath.Base(req.ResumeFilePath))
	if err != nil {
		return JsonData{}, fmt.Errorf("failed to create form file for resume: %w", err)
	}
	if _, err := io.Copy(resumePart, bytes.NewReader(req.ResumeData)); err != nil {
		return JsonData{}, fmt.Errorf("failed to write resume data: %w", err)
	}

	writer.Close()

	baseURL := req.URL
	resp, err := http.Post(baseURL, writer.FormDataContentType(), body)
	if err != nil {
		return JsonData{}, fmt.Errorf("failed to call AI job API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return JsonData{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if strings.HasPrefix(string(respBytes), "<!DOCTYPE html>") {
		return JsonData{}, errors.New("HTML received instead of JSON: probably an API error")
	}

	var jsonResp JsonData
	if err := json.Unmarshal(respBytes, &jsonResp); err != nil {
		return JsonData{}, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	return jsonResp, nil
}
