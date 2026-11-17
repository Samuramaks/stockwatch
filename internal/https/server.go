package https

import (
	"net/http"

	"github.com/gorilla/mux"
)

type HTTPServer struct {
	stockHandlers   *HTTPHandlers
	printerHandlers *PrinterHandlers
}

func NewHTTPServer(handlers *HTTPHandlers, printerHandlers *PrinterHandlers) *HTTPServer {
	return &HTTPServer{
		stockHandlers:   handlers,
		printerHandlers: printerHandlers,
	}
}

func (s *HTTPServer) StartServer() error {
	router := mux.NewRouter()
	//-----------------StockItem------------------
	router.Path("/item").Methods("GET").HandlerFunc(s.stockHandlers.GetListStockItemHandlers)
	router.Path("/item").Methods("POST").HandlerFunc(s.stockHandlers.AddStockItemHandlers)
	router.Path("/item/{id:[0-9]+}").Methods("GET").HandlerFunc(s.stockHandlers.GetStockItemByIdHandlers)
	router.Path("/item/{sku}").Methods("GET").HandlerFunc(s.stockHandlers.GetStockItemBySKUHandlers)
	router.Path("/item/{sku}").Methods("PATCH").HandlerFunc(s.stockHandlers.SetQuantityHandlers)
	router.Path("/item/{sku}").Methods("DELETE").HandlerFunc(s.stockHandlers.DeleteStockItemBySKUHandlers)

	//-----------------Printer------------------
	router.Path("/printers").Methods("POST").HandlerFunc(s.printerHandlers.AddPrinterHandlers)
	router.Path("/printers").Methods("GET").HandlerFunc(s.printerHandlers.GetListPrintersHandlers)
	router.Path("/printers/{id}").Methods("GET").HandlerFunc(s.printerHandlers.GetPrinterByIdHandlers)
	router.Path("/printers/{id}").Methods("DELETE").HandlerFunc(s.printerHandlers.DeletePrinterHandlers)

	//-----------------Static / Frontend------------------
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))

	return http.ListenAndServe(":9091", router)
}
