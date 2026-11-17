package snmp

import "errors"

var (
	ErrUnreachable   = errors.New("принтер не отвечает")
	ErrTonerNotFound = errors.New("строка тонера не найдена")
	ErrLevelUnknown  = errors.New("уровень тонера неизвестен")
	ErrManyToners    = errors.New("найдено несколько строк тонера")
)
