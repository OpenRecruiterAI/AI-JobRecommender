package aijobrecommender

import (
	"encoding/json"
	"fmt"
	"os"
)

func CreateTempFile(dirPath, randomName, fileExtension string) (file *os.File, fileName string) {

	file, err := os.CreateTemp(dirPath+"/", fileName+"-*"+fileExtension)
	if err != nil {
		return &os.File{}, ""
	}

	return file, file.Name()
}
func GetJobListJson(userSkills []string,jobrecommender Jobrecommender) (jsonFile *os.File, job []Jobs, err error) {

	jobList, err := GetJobsListBySkills(userSkills, jobrecommender.DB)
	if err != nil {
		return &os.File{}, []Jobs{}, err
	}

	jsonData, err := json.MarshalIndent(jobList, "", " ")
	if err != nil {
		return &os.File{}, []Jobs{}, err
	}

	dirPath := "./storage/json"

	err = os.MkdirAll(dirPath, 0755)
	if err != nil {
		return &os.File{}, []Jobs{}, err
	}

	_, err = os.Stat(dirPath + "/job_list.json")
	if err != nil {
		err = os.Remove(dirPath + "/job_list.json")
		if err != nil {
			fmt.Println("erbxcjjdn", err)
		}
	}

	jsonFile, err = os.Create(dirPath + "/job_list.json")
	if err != nil {
		return &os.File{}, []Jobs{}, err
	}

	jsonFile.Write([]byte{})

	_, err = jsonFile.Write(jsonData)
	if err != nil {
		return &os.File{}, []Jobs{}, err
	}

	return jsonFile, jobList, nil

}