package cli_parser

import (
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/constants"
	"github.com/GlobchanskyDenis/taskmaster.git/pkg/dto"
	"strconv"
	"strings"
	"errors"
)

/*	Методы этой структуры парсят изначальную строку в подготовленную dto  */
type parser struct {
	rawCliCommand      string
	rawCliCommandParts []string
	cliCommand         dto.CliCommand
}

func ParseCliCommand(rawCliCommand string) (dto.CliCommand, error) {
	/*	Создаем сущность парсера  */
	entity := newParser(rawCliCommand)

	/*	Избавляемся от пробелов и табуляций  */
	entity.split()

	if err := entity.parseCommandName(); err != nil {
		return entity.cliCommand, err
	}

	if err := entity.parseUnitName(); err != nil {
		return entity.cliCommand, err
	}

	if err := entity.parseArguments(); err != nil {
		return entity.cliCommand, err
	}

	/*	Не предусмотрено более одного аргумента к команде  */
	if len(entity.rawCliCommandParts) != 0 {
		return  entity.cliCommand, errors.New("Избыточный аргумент " + entity.rawCliCommandParts[0])
	}

	return entity.cliCommand, nil
}

func newParser(rawCliCommand string) *parser {
	return &parser{
		rawCliCommand: rawCliCommand,
	}
}

/*	Избавляемся от пробелов и табуляций  */
func (entity *parser) split() {
	splittedTabParts := strings.Split(entity.rawCliCommand, "	")
	for _, part := range splittedTabParts {
		if part != "" {
			splittedSpaceParts := strings.Split(part, " ")
			for _, part := range splittedSpaceParts {
				trimmedPart := strings.Trim(part, "	 ")
				if trimmedPart != "" {
					entity.rawCliCommandParts = append(entity.rawCliCommandParts, trimmedPart)
				}
			}
		}
	}
}

func (entity *parser) parseCommandName() error {
	/*	Это не является причиной для ошибки. Может пользователь вообще пустую строку ввел. Нет команды - нет ошибки  */
	if len(entity.rawCliCommandParts) == 0 {
		return nil
	}

	commandName := entity.rawCliCommandParts[0]

	switch commandName {
	case "status", "Status", "STATUS":
		entity.cliCommand.CommandName = "status"
		entity.cliCommand.CommandType = 100
	case "stop", "Stop", "STOP":
		entity.cliCommand.CommandName = "stop"
		entity.cliCommand.CommandType = 101
	case "start", "Start", "START":
		entity.cliCommand.CommandName = "start"
		entity.cliCommand.CommandType = 102
	case "restart", "Restart", "RESTART":
		entity.cliCommand.CommandName = "restart"
		entity.cliCommand.CommandType = 103
	case "kill", "Kill", "KILL":
		entity.cliCommand.CommandName = "kill"
		entity.cliCommand.CommandType = 104
	default:
		return errors.New("Неизвестная команда " + commandName)
	}

	entity.rawCliCommandParts = entity.rawCliCommandParts[1:]

	return nil
}

func (entity *parser) parseUnitName() error {
	/*	Если команда уже найдена, а вот имени процесса не указано */
	if len(entity.rawCliCommandParts) == 0 && entity.cliCommand.CommandType != 0 {
		return errors.New("Не указано имя процесса")
	}

	/*	Если команда не найдена и имени процесса не указано это валидный случай. Может нам вообще пустую строку прислали   */
	if len(entity.rawCliCommandParts) == 0 && entity.cliCommand.CommandType == 0 {
		return nil
	}

	/*	Если предполагаемое имя юнита начинается с знака '-' - значит это не имя юнита а флаг
	**	Есть вариант команд без имени юнита например STATUS -ALL
	**	В этом случае проверка ошибок идет на этапе парсинга аргументов  */
	if isArgumentLooksLikeUnitName(entity.rawCliCommandParts[len(entity.rawCliCommandParts) - 1]) == false {
		return nil
	}

	entity.cliCommand.UnitName = entity.rawCliCommandParts[len(entity.rawCliCommandParts) - 1]
	entity.rawCliCommandParts = entity.rawCliCommandParts[:len(entity.rawCliCommandParts) - 1]
	
	return nil
}

func isArgumentLooksLikeUnitName(arg string) bool {
	if arg[0] == '-' {
		return false
	}

	// if _, err := strconv.ParseUint(arg, 10, 64); err == nil {
	// 	return false
	// }

	return true
}

func (entity *parser) parseArguments() error {
	/*	Если аргументов нет  */
	if len(entity.rawCliCommandParts) == 0 {
		return nil
	}

	/*	У каждой комманды аргументы могут быть разными. У некоторых они вообще не предусмотрены  */
	switch entity.cliCommand.CommandName {
	case "status":
		return entity.parseStatusCommandArguments()
	case "stop":
		return entity.parseStopCommandArguments()
	case "start":
		return entity.parseStartCommandArguments()
	case "restart":
		return entity.parseRestartCommandArguments()
	case "kill":
		return entity.parseKillCommandArguments()
	}

	return nil
}

func (entity *parser) parseStatusCommandArguments() error {
	switch entity.rawCliCommandParts[0] {
	case "-all", "-All", "-ALL":
		return entity.parseStatusAllCommandArgument()
	case "-n":
		return entity.parseStatusNCommandArgument()
	}
	return errors.New("У команды status не предусмотрено аргумента " + entity.rawCliCommandParts[0])
}

func (entity *parser) parseStatusAllCommandArgument() error {
	if entity.cliCommand.UnitName != "" {
		return errors.New("status -all невозможно сочетать с указанным именем юнита " + entity.cliCommand.UnitName)
	}

	/*	Удаляю аргумент из нераспарсеных  */
	entity.rawCliCommandParts = entity.rawCliCommandParts[1:]

	/*	Добавляю аргумент к распарсеным  */
	entity.cliCommand.Args = append(entity.cliCommand.Args, dto.CliArgument{
		Name: "-all",
		Number: constants.ARGUMENT_STATUS_ALL,
		Value: 0,
	}) 
	return nil
}

func (entity *parser) parseStatusNCommandArgument() error {
	/*	Удаляю аргумент из нераспарсеных  */
	entity.rawCliCommandParts = entity.rawCliCommandParts[1:]

	if len(entity.rawCliCommandParts) == 0 {
		return errors.New("У аргумента команды status -n не обнаружено численного значения")
	}
	value, err := strconv.ParseUint(entity.rawCliCommandParts[0], 10, 64)
	if err != nil {
		return errors.New("У аргумента команды status -n неправильное числовое значение (" + err.Error() + ")")
	}

	/*	Удаляю аргумент из нераспарсеных  */
	entity.rawCliCommandParts = entity.rawCliCommandParts[1:]

	/*	Добавляю аргумент к распарсеным  */
	entity.cliCommand.Args = append(entity.cliCommand.Args, dto.CliArgument{
		Name: "-n",
		Number: constants.ARGUMENT_STATUS_NUMBER,
		Value: uint(value),
	}) 
	return nil
}

func (entity *parser) parseStopCommandArguments() error {
	return errors.New("У команды stop не предусмотрено аргумента " + entity.rawCliCommandParts[0])
}

func (entity *parser) parseStartCommandArguments() error {
	return errors.New("У команды start не предусмотрено аргумента " + entity.rawCliCommandParts[0])
}

func (entity *parser) parseRestartCommandArguments() error {
	return errors.New("У команды restart не предусмотрено аргумента " + entity.rawCliCommandParts[0])
}

func (entity *parser) parseKillCommandArguments() error {
	return errors.New("У команды kill не предусмотрено аргумента " + entity.rawCliCommandParts[0])
}
