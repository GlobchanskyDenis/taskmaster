package file_logger

import (
	// "10.10.11.220/ursgis/cdocs_epgu_sender_receiver.git/pkg/utils"
	"fmt"
	"os"
	"sync"
	"time"
)

var GLogger Logger

/*	мьютекс необходим так как под капотом логгер может менять файл и необходимо обеспечить потокобезопасность  */
type logger struct {
	filePath        string
	fileName        string
	permissions     string
	currentFileDate time.Time
	fileOS          *os.File
	mu              *sync.Mutex
}

type Logger interface {
	LogPanic(Fields map[string]interface{}, err error, msg string)
	LogFatal(Fields map[string]interface{}, err error, msg string)
	LogError(Fields map[string]interface{}, err error, msg string)
	LogWarning(Fields map[string]interface{}, err error, msg string)
	LogInfo(Fields map[string]interface{}, msg string)
	LogDebug(Fields map[string]interface{}, msg string)
}

func NewLogger() error {
	if err := checkConfig(); err != nil {
		return err
	}
	conf := GetConfig()

	newLogger := &logger{
		filePath:    conf.LogFolder,
		permissions: conf.Permissions,
		mu:          &sync.Mutex{},
	}

	if err := newLogger.setNewLogFile(); err != nil {
		return err
	}

	GLogger = newLogger

	return nil
}

func (l *logger) LogPanic(fields map[string]interface{}, err error, msg string) {
	if newErr := l.changeLogFileIfItNeeded(); newErr != nil {
		fmt.Printf("При смене лог файла %s %s %#v %s\n", newErr, err, fields, msg)
	} else {
		message := prepareToLogThis("panic", fields, err, msg)
		/*	Повторная попытка нужна для случая когда дата сменилась и какой-то воркер
		**	в другом потоке начал менять дескриптор файла для логгирования. Считаю секунды
		**	достаточно. Не блокировать же мьютексами запись в файл :)  */
		if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
			time.Sleep(1 * time.Second)
			if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
				print(err.Error() + " => " + message)
			}
		}
	}
}

func (l *logger) LogFatal(fields map[string]interface{}, err error, msg string) {
	if newErr := l.changeLogFileIfItNeeded(); newErr != nil {
		fmt.Printf("При смене лог файла %s %s %#v %s\n", newErr, err, fields, msg)
	} else {
		message := prepareToLogThis("fatal", fields, err, msg)
		/*	Повторная попытка нужна для случая когда дата сменилась и какой-то воркер
		**	в другом потоке начал менять дескриптор файла для логгирования. Считаю секунды
		**	достаточно. Не блокировать же мьютексами запись в файл :)  */
		if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
			time.Sleep(1 * time.Second)
			if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
				print(err.Error() + " => " + message)
			}
		}
	}
}

func (l *logger) LogError(fields map[string]interface{}, err error, msg string) {
	if newErr := l.changeLogFileIfItNeeded(); newErr != nil {
		fmt.Printf("При смене лог файла %s %s %#v %s\n", newErr, err, fields, msg)
	} else {
		message := prepareToLogThis("error", fields, err, msg)
		/*	Повторная попытка нужна для случая когда дата сменилась и какой-то воркер
		**	в другом потоке начал менять дескриптор файла для логгирования. Считаю секунды
		**	достаточно. Не блокировать же мьютексами запись в файл :)  */
		if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
			time.Sleep(1 * time.Second)
			if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
				print(err.Error() + " => " + message)
			}
		}
	}
}

func (l *logger) LogWarning(fields map[string]interface{}, err error, msg string) {
	if newErr := l.changeLogFileIfItNeeded(); newErr != nil {
		fmt.Printf("При смене лог файла %s %s %#v %s\n", newErr, err, fields, msg)
	} else {
		message := prepareToLogThis("warning", fields, err, msg)
		/*	Повторная попытка нужна для случая когда дата сменилась и какой-то воркер
		**	в другом потоке начал менять дескриптор файла для логгирования. Считаю секунды
		**	достаточно. Не блокировать же мьютексами запись в файл :)  */
		if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
			time.Sleep(1 * time.Second)
			if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
				print(err.Error() + " => " + message)
			}
		}
	}
}

func (l *logger) LogInfo(fields map[string]interface{}, msg string) {
	if newErr := l.changeLogFileIfItNeeded(); newErr != nil {
		fmt.Printf("При смене лог файла %s %#v %s\n", newErr, fields, msg)
	} else {
		message := prepareToLogThis("info", fields, nil, msg)
		/*	Повторная попытка нужна для случая когда дата сменилась и какой-то воркер
		**	в другом потоке начал менять дескриптор файла для логгирования. Считаю секунды
		**	достаточно. Не блокировать же мьютексами запись в файл :)  */
		if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
			time.Sleep(1 * time.Second)
			if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
				print(err.Error() + " => " + message)
			}
		}
	}
}

func (l *logger) LogDebug(fields map[string]interface{}, msg string) {
	if newErr := l.changeLogFileIfItNeeded(); newErr != nil {
		fmt.Printf("При смене лог файла %s %#v %s\n", newErr, fields, msg)
	} else {
		message := prepareToLogThis("debug", fields, nil, msg)
		/*	Повторная попытка нужна для случая когда дата сменилась и какой-то воркер
		**	в другом потоке начал менять дескриптор файла для логгирования. Считаю секунды
		**	достаточно. Не блокировать же мьютексами запись в файл :)  */
		if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
			time.Sleep(1 * time.Second)
			if _, err := fmt.Fprintf(l.fileOS, "%s", message); err != nil {
				print(err.Error() + " => " + message)
			}
		}
	}
}

func prepareToLogThis(level string, fields map[string]interface{}, err error, msg string) string {
	var dst string

	dst += `{"level":"` + level + `",`
	for key, value := range fields {
		switch typed := value.(type) {
		case uint, int, uint64, uint32, int32, int64:
			dst += fmt.Sprintf("\"%s\":%d,", key, typed)
		case string:
			dst += fmt.Sprintf("\"%s\":\"%s\",", key, typed)
		default:
			dst += fmt.Sprintf("\"%s\":%#v,", key, value)
		}
	}
	if err != nil {
		dst += `"error":"` + err.Error() + `",`
	}
	dst += `"message":"` + msg + `",`
	now := time.Now()
	dst += fmt.Sprintf("\"time\":\"%d-%d-%d %d:%02d:%02d\"}\n", now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), now.Second())
	return dst
}

/*	Меняет файл в который записывается логгирование в случае если уже сменилась дата
**	Использует мьютекс, поэтому выполняется горутинобезопасно  */
func (l *logger) changeLogFileIfItNeeded() error {
	if isSameDate(l.currentFileDate, time.Now()) == false {
		l.mu.Lock()
		defer l.mu.Unlock()
		if err := l.setNewLogFile(); err != nil {
			return err
		}
	}
	return nil
}

func isSameDate(oldDate, now time.Time) bool {
	if oldDate.Year() == now.Year() && oldDate.Month() == now.Month() && oldDate.Day() == now.Day() {
		return true
	}
	return false
}

/*	Данную функцию в многопоточном режиме нужно запускать в   */
func (l *logger) setNewLogFile() error {
	/*	Если ранее уже был открыт файл - закрываю его  */
	if l.fileOS != nil {
		if err := l.fileOS.Close(); err != nil {
			return err
		}
	}

	if err := checkConfig(); err != nil {
		return err
	}
	conf := GetConfig()

	/*	Открываю новый файл  */
	l.currentFileDate = time.Now()
	l.fileName = fmt.Sprintf("%s_%d-%d-%d.log", conf.DaemonName, l.currentFileDate.Year(), l.currentFileDate.Month(), l.currentFileDate.Day())
	osFile, err := openOrCreateNewFile(l.filePath, l.fileName, l.permissions)
	if err != nil {
		return err
	}
	l.fileOS = osFile
	return nil
}
