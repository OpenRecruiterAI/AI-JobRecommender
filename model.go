package aijobrecommender

import (
	"os"
	"time"

	"gorm.io/gorm"
)

type TblJobs struct {
	Id                int       `gorm:"primaryKey;type:serial"`
	Uuid              string    `gorm:"type:character varying"`
	CompanyId         int       `gorm:"type:integer"`
	RecruiterId       int       `gorm:"type:integer"`
	JobTitle          string    `gorm:"type:character varying"`
	MinExperience     string    `gorm:"type:character varying"`
	MaxExperience     string    `gorm:"type:character varying"`
	MinSalary         string    `gorm:"type:character varying"`
	MaxSalary         string    `gorm:"type:character varying"`
	Industry          string    `gorm:"type:character varying"`
	WorkMode          string    `gorm:"type:character varying"`
	Location          string    `gorm:"type:character varying"`
	JobDescription    string    `gorm:"type:character varying"`
	SkillsRequired    string    `gorm:"type:character varying"`
	EducationRequired string    `gorm:"type:character varying"`
	EmploymentType    string    `gorm:"type:character varying"`
	PerksAndBenefits  string    `gorm:"type:character varying"`
	LastDateToApply   time.Time `gorm:"type:timestamp without time zone"`
	OpeningDate       time.Time `gorm:"type:timestamp without time zone"`
	JobStatus         string    `gorm:"type:character varying"`
	AdditionalNotes   string    `gorm:"type:character varying"`
	CreatedOn         time.Time `gorm:"type:timestamp without time zone"`
	CreatedBy         int       `gorm:"type:integer"`
	ModifiedOn        time.Time `gorm:"type:timestamp without time zone;DEFAULT:NULL"`
	ModifiedBy        int       `gorm:"DEFAULT:NULL;type:integer"`
	IsDeleted         int       `gorm:"DEFAULT:0;type:integer"`
	DeletedBy         int       `gorm:"DEFAULT:NULL;type:integer"`
	DeletedOn         time.Time `gorm:"type:timestamp without time zone;DEFAULT:NULL"`
	IsActive          int       `gorm:"DEFAULT:1;type:integer"`
	IsProxy           int       `gorm:"DEFAULT:NULL;type:integer"`
	CandidatesId      string    `gorm:"type:character varying"`
	CandidateCount    int       `gorm:"-"`
	InterviewLink     string    `gorm:"column:interview_link"`
	Skills            []string  `gorm:"-"`
}
type Jobs struct {
	Id int
}

type JsonData struct {
	Status  string `json:"status"`
	Data    Data   `json:"data"`
	Message string `json:"message"`
}
type Data struct {
	JobIds       []string `json:"jobIds"`
	CandidateIds []string `json:"candidateIds"`
}

type TblJobsRecommendation struct {
	Id         int
	CandidateId int
	JobsId     string
	CreatedBy  int
	CreatedOn  time.Time
	ModifiedOn time.Time
	ModifiedBy int
}

type JobRecommendationRequest struct {
    ResumeFilePath string
    ResumeData     []byte
    UserSkills     []string
    CandidateID    int
    URL            string
    JSONFile       *os.File
}

func GetJobsListBySkills(skills []string, DB *gorm.DB) (jobList []Jobs, err error) {

	for _, skill := range skills {
		var tempJobs []Jobs
		query := DB.Debug().Table("tbl_jobs").Where("is_deleted = 0 and  LOWER(TRIM(skills_required)) like LOWER(TRIM(?))", "%"+skill+"%").Find(&tempJobs)
		if query.Error != nil {
			return []Jobs{}, query.Error
		}

		for i := 0; i < len(tempJobs); i++ {
			jobList = append(jobList, tempJobs[i])
		}

	}

	return jobList, nil
}
func GetJobList(jobIds []string, DB *gorm.DB) (jobLists []Jobs, err error) {

	query := DB.Debug().Table("tbl_jobs")

	query = query.Where("id in ? and is_deleted = 0", jobIds)

	result := query.Find(&jobLists)
	if result.Error != nil {
		return []Jobs{}, err
	}

	return jobLists, nil
}
func GetJobRecommendationById(candidateid int, DB *gorm.DB) (recjob TblJobsRecommendation, err error) {

	query := DB.Debug().Table("tbl_jobs_recommendations").Where("candidate_id = ? and is_deleted = 0", candidateid).First(&recjob)
	if query.Error != nil {
		return TblJobsRecommendation{}, query.Error
	}

	return recjob, nil
}
func CreateJobRecommendation(jobData TblJobsRecommendation, DB *gorm.DB) error {

	query := DB.Debug().Table("tbl_jobs_recommendations").Create(&jobData)

	if query.Error != nil {
		return query.Error
	}

	return nil
}
func UpdateJobRecommendation(recjob *TblJobsRecommendation, id int, DB *gorm.DB) (err error) {

	query := DB.Debug().Table("tbl_jobs_recommendations").Where("id = ? and is_deleted = 0", id).Updates(map[string]interface{}{"jobs_id": recjob.JobsId, "modified_on": recjob.ModifiedOn, "modified_by": recjob.ModifiedBy})
	if query.Error != nil {
		return query.Error
	}

	return nil
}