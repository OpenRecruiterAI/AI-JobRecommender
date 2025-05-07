package aijobrecommender

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Db connection
func DBSetup() (*gorm.DB, error) {

	dbConfig := map[string]string{
		"username": "postgres",
		"password": "picco123@",
		"host":     "localhost",
		"port":     "5432",
		"dbname":   "Anyopeningsai",
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN: "user=" + dbConfig["username"] + " password=" + dbConfig["password"] +
			" dbname=" + dbConfig["dbname"] + " host=" + dbConfig["host"] +
			" port=" + dbConfig["port"] + " sslmode=disable TimeZone=Asia/Kolkata",
	}), &gorm.Config{})

	if err != nil {

		log.Fatal("Failed to connect to database:", err)

	}
	if err != nil {

		return nil, err

	}

	return db, nil
}

func TestResumeAnalyze(t *testing.T) {


	db, _ := DBSetup()
	fmt.Println("connection", db)
	filePath := "sample_resume.pdf" // your test file
	file, err := os.Open(filePath)
	assert.NoError(t, err)
	defer file.Close()

	
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	

	formFile, err := writer.CreateFormFile("file", filepath.Base(filePath))
	assert.NoError(t, err)

	
	_, err = io.Copy(formFile, file)
	assert.NoError(t, err)
	writer.Close()

	
	req := httptest.NewRequest(http.MethodPost, "/upload", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	
	err = req.ParseMultipartForm(10 << 20) // 10 MB
	assert.NoError(t, err)

	
	fileHeader := req.MultipartForm.File["file"][0]

	
	jobrecommender := &Jobrecommender{
		DB: db,
	}
	respData, statusCode, err := jobrecommender.ResumeAnalyze(fileHeader, "")

	fmt.Println("respData err", err)
	fmt.Println("respData", respData, statusCode)
}
func TestJobrecommendation(t *testing.T) {
	

	resumeFilePath := "sample_resume.pdf"
	resumeFile, err := os.Open(resumeFilePath)
	assert.NoError(t, err)
	defer resumeFile.Close()


	resumeData, err := io.ReadAll(resumeFile)
	assert.NoError(t, err)

	
	jsonFilePath := "sample_jobs.json"
	jsonFile, err := os.Open(jsonFilePath)
	assert.NoError(t, err)
	defer jsonFile.Close()

	
	req := JobRecommendationRequest{
		ResumeFilePath: resumeFilePath,
		ResumeData:     resumeData,
		UserSkills:     []string{"Go", "API", "PostgreSQL"},
		CandidateID:    1,
		URL:            "", // Replace with mock if possible
		JSONFile:       jsonFile,
	}

	jobrecommender := &Jobrecommender{}
	respData, err := jobrecommender.Jobrecommendation(req)

	// Assertions
	fmt.Println("respData err", err)
	fmt.Println("respData", respData, respData.Status)
}

