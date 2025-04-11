package aijobrecommender

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func JobrecommenderSetup(config Config) *Jobrecommender {

	return &Jobrecommender{
		DB: config.DB,
	}

}
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

	var responseData map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&responseData); err != nil {
		return nil, resp.Status, fmt.Errorf("failed to decode response: %w", err)
	}

	return responseData, resp.Status, nil

}

// Function signature is cleaner now
func (jobrecommender *Jobrecommender) Jobrecommendation(resumeFilePath string, resumeData []byte, userSkills []string, candidateID int, url string) ([]Jobs, error) {

	jsonFile, Jobslist, err := GetJobListJson(userSkills, *jobrecommender)
	if err != nil {
		return nil, fmt.Errorf("failed to get job list JSON: %w", err)
	}
	defer jsonFile.Close()

	_, err = jsonFile.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat JSON file: %w", err)
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	jsonPart, err := writer.CreateFormFile("files", filepath.Base(jsonFile.Name()))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file for JSON: %w", err)
	}
	if _, err := jsonFile.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("failed to seek JSON file: %w", err)
	}
	if _, err := io.Copy(jsonPart, jsonFile); err != nil {
		return nil, fmt.Errorf("failed to write JSON content: %w", err)
	}

	resumePart, err := writer.CreateFormFile("files", filepath.Base(resumeFilePath))
	if err != nil {
		return nil, fmt.Errorf("failed to create form file for resume: %w", err)
	}
	if _, err := io.Copy(resumePart, bytes.NewReader(resumeData)); err != nil {
		return nil, fmt.Errorf("failed to write resume data: %w", err)
	}

	writer.Close()

	baseURL := os.Getenv("CHAT_API_URL") + "api/analysis/ai-job-list"
	resp, err := http.Post(baseURL, writer.FormDataContentType(), body)
	if err != nil {
		return nil, fmt.Errorf("failed to call AI job API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if strings.HasPrefix(string(respBytes), "<!DOCTYPE html>") {
		return nil, errors.New("HTML received instead of JSON: probably an API error")
	}

	var jsonResp JsonData
	if err := json.Unmarshal(respBytes, &jsonResp); err != nil {
		return nil, fmt.Errorf("failed to parse response JSON: %w", err)
	}

	joblist, err := GetJobList(jsonResp.Data.JobIds, jobrecommender.DB)
	if err != nil {
		return nil, fmt.Errorf("failed to get job list from DB: %w", err)
	}

	matchedIDs := map[int]bool{}
	var recommendedJobs []Jobs
	var recommendedJobIDs []string

	for _, job := range joblist {
		for _, j := range Jobslist {
			if job.Id == j.Id && !matchedIDs[job.Id] {
				recommendedJobs = append(recommendedJobs, job)
				recommendedJobIDs = append(recommendedJobIDs, strconv.Itoa(job.Id))
				matchedIDs[job.Id] = true
				break
			}
		}
	}

	idStr := strings.Join(recommendedJobIDs, ",")

	recommendation, err := GetJobRecommendationById(candidateID, jobrecommender.DB)
	now := time.Now().UTC()

	if err != nil {
		rejob := TblJobsRecommendation{
			CandidateId: candidateID,
			JobsId:      idStr,
			CreatedBy:   candidateID,
			CreatedOn:   now,
			ModifiedOn:  now,
		}
		if err := CreateJobRecommendation(rejob, jobrecommender.DB); err != nil {
			return nil, fmt.Errorf("failed to create recommendation: %w", err)
		}
	} else {
		recommendation.JobsId = idStr
		recommendation.ModifiedBy = candidateID
		recommendation.ModifiedOn = now

		if err := UpdateJobRecommendation(&recommendation, recommendation.Id, jobrecommender.DB); err != nil {
			return nil, fmt.Errorf("failed to update recommendation: %w", err)
		}
	}

	return recommendedJobs, nil
}
