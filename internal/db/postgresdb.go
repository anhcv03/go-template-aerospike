package db

import (
	"fmt"
	
	config "gitlab.vht.vn/tt-kttt/lae-project/utm/utm-track-manager/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"

	"github.com/rs/zerolog/log"
)

type PostgresDB struct {
	DbConfig  *config.PostgresqlConfig
}

func NewPostgresConnection(dbCfg config.PostgresqlConfig) *PostgresDB {
	return &PostgresDB {
		DbConfig: &dbCfg,
	}
}

func (pd *PostgresDB) ConnectDatabase() *gorm.DB {
	pd.ensureDatabase()
	
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		pd.DbConfig.DBHost, pd.DbConfig.DBPort, pd.DbConfig.DBUser, pd.DbConfig.DBPass, pd.DbConfig.DBName)
	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{
		// TranslateError: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "track_manager.",
			SingularTable: false,
		},
	})

	if err != nil {
		log.Fatal().Err(err).Msg("cannot connect to Postgresql")
		return nil
	}

	return db
}

func (pd *PostgresDB) ensureDatabase() {
	dnsValidate := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable" , 
	pd.DbConfig.DBHost, pd.DbConfig.DBPort, pd.DbConfig.DBUser, pd.DbConfig.DBPass)
	dbValidate, err := gorm.Open(postgres.Open(dnsValidate), &gorm.Config{})
	if err != nil {
		log.Error().Err(err).Msg("check database exists failed")
		return 
	}

	databaseName := pd.DbConfig.DBName
	var existDatabase bool
	queryCheckExist := `SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)`
	if err := dbValidate.Raw(queryCheckExist, databaseName).Scan(&existDatabase).Error; err != nil {
		log.Error().Err(err).Msg("check database exists failed")
		return 
	}

	log.Debug().Msgf("Exist Database %s: %v \n", databaseName, existDatabase)

	if !existDatabase {
		queryCreateDatabase := fmt.Sprintf(`CREATE DATABASE "%s"`, databaseName)
		if err := dbValidate.Exec(queryCreateDatabase).Error; err != nil {
			log.Error().Err(err).Msg("create database failed")
			return
		}
		sqlDBValidate, _ := dbValidate.DB()
		sqlDBValidate.Close()
		log.Debug().Msgf("Created Database Successfully And Close Connection %s \n", databaseName)
	}
}