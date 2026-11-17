package https

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	"stockwatch/internal/printer"

	"github.com/gorilla/mux"
)

type PrinterHandlers struct {
	repo printer.Repository
}

func NewPrinterHandlers(repo printer.Repository) *PrinterHandlers {
	return &PrinterHandlers{repo: repo}
}

// AddPrinterHandlers создаёт принтер.
//
// pattern: /printers
// method:  POST
//
// succeed:
//   - status code: 201 Created
//   - response body: -
//
// failed:
//   - status code: 400, 409, 500
//   - response body: JSON with error + time
func (h *PrinterHandlers) AddPrinterHandlers(w http.ResponseWriter, r *http.Request) {
	var req AddPrinterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := req.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	item := printer.Printers{
		Name:          req.Name,
		IP:            req.IP,
		Model:         req.Model,
		Location:      req.Location,
		SNMPCommunity: req.SNMPCommunity,
		CartridgeSKU:  req.CartridgeSKU,
	}

	if err := h.repo.AddPrinter(r.Context(), item); err != nil {
		if errors.Is(err, printer.ErrPrinterAlreadyExists) {
			writeError(w, http.StatusConflict, err.Error())
			return
		}
		log.Printf("add printer: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to add printer")
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// GetPrinterByIdHandlers возвращает принтер по id.
//
// pattern: /printers/{id}
// method:  GET
//
// succeed:
//   - status code: 200 OK
//   - response body: JSON with printer
//
// failed:
//   - status code: 400, 404, 500
//   - response body: JSON with error + time
func (h *PrinterHandlers) GetPrinterByIdHandlers(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	p, err := h.repo.GetPrinter(r.Context(), id)
	if err != nil {
		if errors.Is(err, printer.ErrPrinterNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("get printer: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to get printer")
		return
	}

	writeJSON(w, http.StatusOK, toPrinterResponse(p))
}

// GetListPrintersHandlers возвращает все принтеры со статусом.
//
// pattern: /printers
// method:  GET
//
// succeed:
//   - status code: 200 OK
//   - response body: JSON array of printers
//
// failed:
//   - status code: 500
//   - response body: JSON with error + time
func (h *PrinterHandlers) GetListPrintersHandlers(w http.ResponseWriter, r *http.Request) {
	printers, err := h.repo.ListPrintersWithStatus(r.Context())
	if err != nil {
		log.Printf("list printers: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to list printers")
		return
	}

	result := make([]PrinterResponse, 0, len(printers)) // пустой слайс, чтобы в JSON был [], а не null
	for _, p := range printers {
		result = append(result, toPrinterResponse(p))
	}

	writeJSON(w, http.StatusOK, result)
}

// DeletePrinterHandlers удаляет принтер по id.
//
// pattern: /printers/{id}
// method:  DELETE
//
// succeed:
//   - status code: 204 No Content
//   - response body: -
//
// failed:
//   - status code: 400, 404, 500
//   - response body: JSON with error + time
func (h *PrinterHandlers) DeletePrinterHandlers(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(mux.Vars(r)["id"], 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid id")
		return
	}

	if err := h.repo.DeletePrinter(r.Context(), id); err != nil {
		if errors.Is(err, printer.ErrPrinterNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		log.Printf("delete printer: %v", err)
		writeError(w, http.StatusInternalServerError, "failed to delete printer")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ---------- Helpers ----------

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err := w.Write([]byte(NewErrorDTO(msg, time.Now()).toJson())); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
