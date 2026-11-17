package https

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"stockwatch/internal/warehouse"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

type HTTPHandlers struct {
	stockItem warehouse.Repository
}

func NewHTTPHandler(stockItem warehouse.Repository) *HTTPHandlers {
	return &HTTPHandlers{
		stockItem: stockItem,
	}
}

func (h *HTTPHandlers) AddStockItemHandlers(w http.ResponseWriter, r *http.Request) {
	var stockItemDTO StockItemDTO

	ctx := r.Context()

	if err := json.NewDecoder(r.Body).Decode(&stockItemDTO); err != nil {
		errDto := NewErrorDTO(err.Error(), time.Now())

		http.Error(w, errDto.toJson(), http.StatusBadRequest)
		return
	}

	if err := stockItemDTO.ValidateForCreate(); err != nil {
		errDto := NewErrorDTO(err.Error(), time.Now())

		http.Error(w, errDto.toJson(), http.StatusBadRequest)
		return
	}

	item := warehouse.StockItem{
		Name:      stockItemDTO.Name,
		SKU:       stockItemDTO.SKU,
		Quantity:  stockItemDTO.Quantity,
		Threshold: stockItemDTO.Threshold,
		Unit:      stockItemDTO.Unit,
	}

	if err := h.stockItem.AddStockItem(ctx, item); err != nil {
		if errors.Is(err, warehouse.ErrStockItemAlreadyExists) {
			errDto := NewErrorDTO(err.Error(), time.Now())

			http.Error(w, errDto.toJson(), http.StatusConflict)
			return
		}

		errDto := NewErrorDTO("failed to add stock item", time.Now())

		http.Error(w, errDto.toJson(), http.StatusInternalServerError)
		return

	}

	w.WriteHeader(http.StatusCreated)

}

func (h *HTTPHandlers) DeleteStockItemBySKUHandlers(w http.ResponseWriter, r *http.Request) {
	sku := mux.Vars(r)["sku"]

	ctx := r.Context()

	if err := h.stockItem.DeleteStockItem(ctx, sku); err != nil {
		errDTO := NewErrorDTO(err.Error(), time.Now())
		if errors.Is(err, warehouse.ErrStockItemNotFound) {
			http.Error(w, errDTO.toJson(), http.StatusNotFound)
		} else {
			http.Error(w, errDTO.toJson(), http.StatusInternalServerError)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)

}

func (h *HTTPHandlers) GetStockItemByIdHandlers(w http.ResponseWriter, r *http.Request) {
	idStr := mux.Vars(r)["id"]
	ctx := r.Context()
	id, err := strconv.ParseInt(idStr, 10, 64)

	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	stockItem, err := h.stockItem.GetStockItem(ctx, id)

	if err != nil {
		errDTO := NewErrorDTO(err.Error(), time.Now())
		if errors.Is(err, warehouse.ErrStockItemNotFound) {
			http.Error(w, errDTO.toJson(), http.StatusNotFound)
		} else {
			http.Error(w, errDTO.toJson(), http.StatusInternalServerError)
		}

		return
	}

	b, err := json.MarshalIndent(stockItem, "", " ")

	if err != nil {
		errDto := NewErrorDTO(err.Error(), time.Now())

		http.Error(w, errDto.toJson(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Failed to write response:", err)
		return
	}

}

func (h *HTTPHandlers) GetStockItemBySKUHandlers(w http.ResponseWriter, r *http.Request) {
	sku := mux.Vars(r)["sku"]
	ctx := r.Context()
	stockItem, err := h.stockItem.GetStockItemBySKU(ctx, sku)

	if err != nil {
		errDTO := NewErrorDTO(err.Error(), time.Now())
		if errors.Is(err, warehouse.ErrStockItemNotFound) {
			http.Error(w, errDTO.toJson(), http.StatusNotFound)
		} else {
			http.Error(w, errDTO.toJson(), http.StatusInternalServerError)
		}

		return
	}

	b, err := json.MarshalIndent(stockItem, "", " ")

	if err != nil {
		errDto := NewErrorDTO(err.Error(), time.Now())

		http.Error(w, errDto.toJson(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Failed to write response:", err)
		return
	}

}

func (h *HTTPHandlers) GetListStockItemHandlers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	result, err := h.stockItem.ListStockItems(ctx)

	if err != nil {
		errDTO := NewErrorDTO(err.Error(), time.Now())
		http.Error(w, errDTO.toJson(), http.StatusInternalServerError)
		return
	}

	b, err := json.MarshalIndent(result, "", " ")

	if err != nil {
		errDto := NewErrorDTO(err.Error(), time.Now())
		http.Error(w, errDto.toJson(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Failed to write response:", err)
		return
	}

}

func (h *HTTPHandlers) SetQuantityHandlers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	sku := mux.Vars(r)["sku"]
	var quantity SetQuantityDTO

	if err := json.NewDecoder(r.Body).Decode(&quantity); err != nil {
		errDTO := NewErrorDTO(err.Error(), time.Now())
		http.Error(w, errDTO.toJson(), http.StatusBadRequest)
		return
	}

	stockItem, err := h.stockItem.SetQuantity(ctx, sku, quantity.Quantity)

	if err != nil {
		errDTO := NewErrorDTO(err.Error(), time.Now())
		if errors.Is(err, warehouse.ErrStockItemNotFound) {
			http.Error(w, errDTO.toJson(), http.StatusNotFound)
		} else {
			http.Error(w, errDTO.toJson(), http.StatusInternalServerError)
		}
		return
	}

	b, err := json.MarshalIndent(stockItem, "", " ")
	if err != nil {
		errDTO := NewErrorDTO(err.Error(), time.Now())
		http.Error(w, errDTO.toJson(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("Failed to write response:", err)
		return
	}

}
