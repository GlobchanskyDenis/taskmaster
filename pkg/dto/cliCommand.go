package dto

import (
)

type CliCommand struct {
	CommandName string
	CommandType uint // Номер из констант совпадающий с именем данной команды
	Args     []CliArgument
	UnitName string
}

type CliArgument struct {
	Name   string
	Number uint  // Номер из констант совпадающий с именем данного аргумента
	Value  uint
}
