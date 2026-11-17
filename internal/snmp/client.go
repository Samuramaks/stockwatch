package snmp

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/gosnmp/gosnmp"
)

const (
	suppliesOID = "1.3.6.1.2.1.43.11.1.1"

	colType  = "5" // prtMarkerSuppliesType
	colMax   = "8" // prtMarkerSuppliesMaxCapacity
	colLevel = "9" // prtMarkerSuppliesLevel

	supplyTypeToner          = 3
	supplyTypeTonerCartridge = 21

	snmpPort = 161
)

// Debug включает вывод сырых данных опроса в лог.
// Community в лог не попадает.
var Debug bool

// supply — одна строка таблицы расходников. Указатели нужны, чтобы
// отличить "значение не пришло" от настоящего нуля.
type supply struct {
	supplyType *int
	max        *int
	level      *int
}

type Client struct{}

func NewClient() *Client { return &Client{} }

// GetTonerLevel опрашивает один принтер и возвращает остаток тонера в процентах.
// При ошибке возвращается 0 и ошибка: ErrUnreachable, ErrTonerNotFound,
// ErrLevelUnknown, ErrManyToners либо ошибка контекста (отмена или дедлайн).
func (c *Client) GetTonerLevel(ctx context.Context, target, community string) (int, error) {
	g := &gosnmp.GoSNMP{
		Target:    target,
		Port:      snmpPort,
		Community: community,
		Version:   gosnmp.Version2c,
		Context:   ctx,
		Retries:   2,
		Timeout:   3 * time.Second,
	}

	if err := g.Connect(); err != nil {
		return 0, fmt.Errorf("%w: %w", ErrUnreachable, err)
	}
	defer g.Conn.Close()

	pdus, err := g.BulkWalkAll(suppliesOID)
	if err != nil {
		// отмена или дедлайн контекста — это не недоступность принтера
		if ctxErr := ctx.Err(); ctxErr != nil {
			return 0, fmt.Errorf("опрос %s прерван: %w", target, ctxErr)
		}
		return 0, fmt.Errorf("%w: %w", ErrUnreachable, err)
	}

	if Debug {
		logPDUs(target, pdus)
	}

	return tonerPercent(pdus)
}

// tonerPercent не работает с сетью, поэтому легко тестируется
// на заранее подготовленном слайсе PDU.
func tonerPercent(pdus []gosnmp.SnmpPDU) (int, error) {
	supplies := parseSupplies(pdus)

	var toner *supply
	for _, s := range supplies {
		if s.supplyType == nil {
			continue
		}
		if *s.supplyType != supplyTypeToner && *s.supplyType != supplyTypeTonerCartridge {
			continue
		}
		if toner != nil {
			return 0, ErrManyToners
		}
		toner = s
	}

	if toner == nil {
		return 0, ErrTonerNotFound
	}

	return calcPercent(toner)
}

// parseSupplies группирует значения таблицы по идентификатору строки.
// Некорректные и служебные записи пропускаются, а не прерывают разбор.
func parseSupplies(pdus []gosnmp.SnmpPDU) map[string]*supply {
	prefix := "." + suppliesOID + "."
	supplies := make(map[string]*supply)

	for _, pdu := range pdus {
		rest, ok := strings.CutPrefix(pdu.Name, prefix)
		if !ok {
			continue
		}

		col, idx, ok := strings.Cut(rest, ".")
		if !ok {
			continue
		}

		if col != colType && col != colMax && col != colLevel {
			continue
		}

		switch pdu.Type {
		case gosnmp.NoSuchObject, gosnmp.NoSuchInstance, gosnmp.EndOfMibView, gosnmp.Null:
			continue
		}

		v, ok := toInt(pdu.Value)
		if !ok {
			continue
		}

		s := supplies[idx]
		if s == nil {
			s = &supply{}
			supplies[idx] = s
		}

		switch col {
		case colType:
			s.supplyType = &v
		case colMax:
			s.max = &v
		case colLevel:
			s.level = &v
		}
	}

	return supplies
}

func calcPercent(s *supply) (int, error) {
	if s.level == nil || s.max == nil {
		return 0, ErrLevelUnknown
	}

	level, capacity := *s.level, *s.max

	// отрицательные значения: -1 другое, -2 неизвестно, -3 "что-то осталось"
	if level < 0 || capacity <= 0 {
		return 0, ErrLevelUnknown
	}

	pct := level * 100 / capacity
	if pct > 100 {
		pct = 100
	}

	return pct, nil
}

// toInt приводит значение SNMP к int. Разные SNMP-типы приходят
// в разных Go-типах (Integer — int, Gauge32 — uint и так далее).
func toInt(v any) (int, bool) {
	switch n := v.(type) {
	case int:
		return n, true
	case int32:
		return int(n), true
	case int64:
		return int(n), true
	case uint:
		if n > math.MaxInt32 {
			return 0, false
		}
		return int(n), true
	case uint32:
		if n > math.MaxInt32 {
			return 0, false
		}
		return int(n), true
	case uint64:
		if n > math.MaxInt32 {
			return 0, false
		}
		return int(n), true
	default:
		return 0, false
	}
}

func logPDUs(target string, pdus []gosnmp.SnmpPDU) {
	for _, pdu := range pdus {
		val := pdu.Value
		if b, ok := val.([]byte); ok {
			val = string(b)
		}
		log.Printf("snmp %s: %s type=%v value=%v", target, pdu.Name, pdu.Type, val)
	}
}
