package common

import "gorm.io/gorm"

func ProbeReadiness(dbClient *gorm.DB) error {
	db, err := dbClient.DB()
	if err != nil {
		return err
	}

	err = db.Ping()

	if err != nil {
		return err
	}
	return nil
}

