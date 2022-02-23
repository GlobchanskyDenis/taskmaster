package file_logger

import (
	"errors"
)

type config struct {
	DaemonName  string `conf:"DaemonName"`
	LogFolder   string `conf:"LogFolder"`
	Permissions string `conf:"Permissions"`
}

/*	Глобальная структура конфига  */
var gConf *config

/*	Возвращает структуру настроек - без заполнения ее полей пакет считается неинициализированным  */
func GetConfig() *config {
	if gConf == nil {
		gConf = &config{}
	}
	return gConf
}

/*	*/
func checkConfig() error {
	if gConf == nil {
		return errors.New("модуль pkg/utils/file_logger не сконфигурирован")
	}
	if gConf.LogFolder == "" {
		return errors.New("параметр LogFolder не может быть пустым")
	}
	return nil
}
