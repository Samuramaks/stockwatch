package printer

import (
	"errors"
	"time"
)

var (
	ErrPrinterNotFound      = errors.New("принтер не найден")
	ErrPrinterAlreadyExists = errors.New("принтер уже существует")
)

type PrinterStatus struct {
	LevelPct      int
	LastPolledAt  time.Time
	LastAlertedAt *time.Time // может не быть, если алерта ещё не было
}

type Printers struct {
	ID            int64
	Name          string
	IP            string
	Model         string
	Location      string
	CartridgeSKU  string
	SNMPCommunity string
	Status        *PrinterStatus // nil, если принтер ещё ни разу не опрашивался
}
