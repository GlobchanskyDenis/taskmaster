package main

import (
	"github.com/GlobchanskyDenis/taskmaster.git/internal/supervisor"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/u_conf"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
)

func initializeConfigs(configFileName string) (supervisor.UnitListConfig, error) {
	/*	Read config file  */
	print("Считываю конфигурационный файл\t- ")
	if err := u_conf.SetConfigFile(configFileName); err != nil {
		println(constants.RED + "ошибка" + constants.NO_COLOR)
		return supervisor.UnitListConfig{}, err
	}
	println(constants.GREEN + "успешно" + constants.NO_COLOR)
	/*	supervisor  */
	print("настраиваю пакет supervisor\t- ")
	unitListConfig := supervisor.UnitListConfig{}
	if err := u_conf.ParsePackageConfig(&unitListConfig, "Units"); err != nil {
		println(constants.RED + "ошибка" + constants.NO_COLOR)
		return supervisor.UnitListConfig{}, err
	}
	println(constants.GREEN + "успешно" + constants.NO_COLOR)

	return unitListConfig, nil
}