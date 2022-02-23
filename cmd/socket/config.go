package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/u_conf"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/file_logger"
)

func initializeConfigs(configFileName string) (dto.UnitListConfig, error) {
	/*	Read config file  */
	print("Считываю конфигурационный файл\t- ")
	if err := u_conf.SetConfigFile(configFileName); err != nil {
		println(constants.RED + "ошибка" + constants.NO_COLOR)
		return dto.UnitListConfig{}, err
	}
	println(constants.GREEN + "успешно" + constants.NO_COLOR)
	/*	supervisor  */
	print("настраиваю пакет supervisor\t- ")
	unitListConfig := dto.UnitListConfig{}
	if err := u_conf.ParsePackageConfig(&unitListConfig, "Units"); err != nil {
		println(constants.RED + "ошибка" + constants.NO_COLOR)
		return dto.UnitListConfig{}, err
	}
	println(constants.GREEN + "успешно" + constants.NO_COLOR)

	/*	file_logger  */
	print("настраиваю пакет file_logger\t- ")
	loggerConf := file_logger.GetConfig()
	if err := u_conf.ParsePackageConfig(loggerConf, "Logger"); err != nil {
		println(constants.RED + "ошибка" + constants.NO_COLOR)
		return dto.UnitListConfig{}, err
	}
	println(constants.GREEN + "успешно" + constants.NO_COLOR)

	return unitListConfig, nil
}