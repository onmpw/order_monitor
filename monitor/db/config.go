package db

import "monitor/monitor/config"

type dbConfig map[string]string

var connections map[string]dbConfig

func loadConfig() error {
	connections = map[string]dbConfig{
		"production": {
			"driver": "mysql",
			"database": config.Conf.C("database"),
			"host":config.Conf.C("host"),
			"username": config.Conf.C("username"),
			"password": config.Conf.C("password"),
		},
		"jd_production": {
			"driver": "mysql",
			"database": config.Conf.C("database"),
			"host":config.Conf.C("jd_host"),
			"username": config.Conf.C("jd_username"),
			"password": config.Conf.C("jd_password"),
			"port": config.Conf.C("jd_port"),
		},
		"uc_production": {
			"driver": "mysql",
			"database": config.Conf.C("uc_database"),
			"host":config.Conf.C("host"),
			"username": config.Conf.C("username"),
			"password": config.Conf.C("password"),
		},
		"development":{
			"driver": "mysql",
			"database": config.Conf.C("local_database"),
			"host":config.Conf.C("local_host"),
			"username": config.Conf.C("local_username"),
			"password": config.Conf.C("local_password"),
		},
	}

	return nil
}