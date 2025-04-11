package aijobrecommender

import "gorm.io/gorm"

type Type string

const (
	Postgres Type = "postgres"
	Mysql    Type = "mysql"
)

type Config struct {
	DB           *gorm.DB
	DataBaseType Type
}

type Jobrecommender struct{
	DB               *gorm.DB
}