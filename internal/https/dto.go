package https

import (
	"encoding/json"
	"errors"
	"net"
	"stockwatch/internal/printer"
	"strings"
	"time"
)

type StockItemDTO struct {
	Name      string `json:"name"`
	SKU       string `json:"sku"`
	Quantity  int    `json:"quantity"`
	Threshold int    `json:"threshold"`
	Unit      string `json:"unit"`
}

func (st StockItemDTO) ValidateForCreate() error {
	if st.Name == "" {
		return errors.New("Имя не должно быть пустым!")
	}

	if st.SKU == "" {
		return errors.New("SKU не должно быть пустым!")
	}

	return nil
}

type ErrDTO struct {
	Message string
	Time    time.Time
}

func NewErrorDTO(message string, time time.Time) ErrDTO {
	return ErrDTO{
		Message: message,
		Time:    time,
	}
}

func (e ErrDTO) toJson() string {
	b, err := json.MarshalIndent(e, "", "    ")

	if err != nil {
		panic(err)
	}

	return string(b)
}

type SetQuantityDTO struct {
	Quantity int `json:"quantity"`
}

// AddPrinterRequest содержит только те поля, которые задаёт клиент.
type AddPrinterRequest struct {
	Name          string `json:"name"`
	IP            string `json:"ip"`
	Model         string `json:"model"`
	Location      string `json:"location"`
	SNMPCommunity string `json:"snmp_community"`
	CartridgeSKU  string `json:"cartridge_sku"`
}

func (r AddPrinterRequest) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("name is required")
	}
	if net.ParseIP(r.IP) == nil {
		return errors.New("ip is invalid")
	}
	if strings.TrimSpace(r.SNMPCommunity) == "" {
		return errors.New("snmp_community is required")
	}
	if strings.TrimSpace(r.CartridgeSKU) == "" {
		return errors.New("cartridge_sku is required")
	}
	return nil
}

// PrinterStatusResponse описывает статус в ответе API.
type PrinterStatusResponse struct {
	LevelPct      int        `json:"level_pct"`
	LastPolledAt  time.Time  `json:"last_polled_at"`
	LastAlertedAt *time.Time `json:"last_alerted_at"`
}

// PrinterResponse намеренно не содержит snmp_community,
// чтобы не отдавать его через API.
type PrinterResponse struct {
	ID           int64                  `json:"id"`
	Name         string                 `json:"name"`
	IP           string                 `json:"ip"`
	Model        string                 `json:"model"`
	Location     string                 `json:"location"`
	CartridgeSKU string                 `json:"cartridge_sku"`
	Status       *PrinterStatusResponse `json:"status"` // null, если принтер ещё не опрашивался
}

func toPrinterResponse(p printer.Printers) PrinterResponse {
	resp := PrinterResponse{
		ID:           p.ID,
		Name:         p.Name,
		IP:           p.IP,
		Model:        p.Model,
		Location:     p.Location,
		CartridgeSKU: p.CartridgeSKU,
	}

	if p.Status != nil {
		resp.Status = &PrinterStatusResponse{
			LevelPct:      p.Status.LevelPct,
			LastPolledAt:  p.Status.LastPolledAt,
			LastAlertedAt: p.Status.LastAlertedAt,
		}
	}

	return resp
}
