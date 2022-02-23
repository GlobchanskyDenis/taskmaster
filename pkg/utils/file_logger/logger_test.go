package file_logger

import (
	// "10.10.11.220/ursgis/cdocs_epgu_sender_receiver.git/pkg/colors"
	// "10.10.11.220/ursgis/u_conf.git"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/utils/u_conf"
	"testing"
	"time"
)

func TestInitAndTestGlobalLogger(t *testing.T) {
	if err := u_conf.SetConfigFile("../../../config/default.json"); err != nil {
		t.Errorf("%sError: %s%s", constants.RED_BG, err, constants.NO_COLOR)
		t.FailNow()
	}
	loggerConf := GetConfig()
	if err := u_conf.ParsePackageConfig(loggerConf, "Logger"); err != nil {
		t.Errorf("%sError: %s%s", constants.RED_BG, err, constants.NO_COLOR)
		t.FailNow()
	}

	if err := NewLogger(); err != nil {
		t.Errorf("%sError: %s%s", constants.RED_BG, err, constants.NO_COLOR)
		t.FailNow()
	}
	GLogger.LogInfo(map[string]interface{}{
		"worker_num": 1,
		"field_n":    "value_n",
	}, "тестовый лог")

	t.Logf("%sSuccess%s", constants.GREEN_BG, constants.NO_COLOR)
}

func TestGlobalLoggerMultithread(t *testing.T) {
	if err := u_conf.SetConfigFile("../../../config/default.json"); err != nil {
		t.Errorf("%sError: %s%s", constants.RED_BG, err, constants.NO_COLOR)
		t.FailNow()
	}
	loggerConf := GetConfig()
	if err := u_conf.ParsePackageConfig(loggerConf, "Logger"); err != nil {
		t.Errorf("%sError: %s%s", constants.RED_BG, err, constants.NO_COLOR)
		t.FailNow()
	}

	if err := NewLogger(); err != nil {
		t.Errorf("%sError: %s%s", constants.RED_BG, err, constants.NO_COLOR)
		t.FailNow()
	}

	for i := 1; i < 50; i++ {
		go t.Run("worker", func(t *testing.T) {
			time.Sleep(1000 * time.Millisecond)
			for j := 0; j < 50; j++ {
				GLogger.LogInfo(map[string]interface{}{
					"worker_num": 1,
					"field_n":    "value_n",
				}, "лог 1")
				time.Sleep(100 * time.Millisecond)
			}
		})
	}

	time.Sleep(9 * time.Second)

	if t.Failed() == false {
		t.Logf("%sSuccess%s", constants.GREEN_BG, constants.NO_COLOR)
	}
}
