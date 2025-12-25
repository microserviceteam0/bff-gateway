package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"go.uber.org/zap"

	"github.com/microserviceteam0/bff-gateway/product-service/internal/dto"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/middleware"
	"github.com/microserviceteam0/bff-gateway/product-service/internal/service"
	"github.com/microserviceteam0/bff-gateway/product-service/pkg/logger"
	"github.com/microserviceteam0/bff-gateway/product-service/pkg/validator"
)

type ProductHandler struct {
	service   service.ProductService
	validator *validator.Validator
}

func NewProductHandler(service service.ProductService) *ProductHandler {
	return &ProductHandler{
		service:   service,
		validator: validator.New(),
	}
}

func (h *ProductHandler) RegisterRoutes(router *mux.Router) {
	router.HandleFunc("/api/products", h.GetAll).Methods(http.MethodGet)
	router.HandleFunc("/api/products/{id}", h.GetByID).Methods(http.MethodGet)
	router.HandleFunc("/api/products", h.Create).Methods(http.MethodPost)
	router.HandleFunc("/api/products/{id}", h.Update).Methods(http.MethodPut)
	router.HandleFunc("/api/products/{id}", h.Delete).Methods(http.MethodDelete)
}

// GetAll получить все продукты
func (h *ProductHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	logger.Debug("fetching all products",
		zap.String("request_id", requestID),
	)

	products, err := h.service.GetAll(r.Context())
	if err != nil {
		logger.Error("failed to fetch products",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Info("products fetched successfully",
		zap.String("request_id", requestID),
		zap.Int("count", len(products)),
	)

	respondJSON(w, http.StatusOK, products)
}

// GetByID получить продукт по id
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		logger.Warn("invalid product id format",
			zap.String("request_id", requestID),
			zap.String("id", vars["id"]),
			zap.Error(err),
		)
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	logger.Debug("fetching product by id",
		zap.String("request_id", requestID),
		zap.Int64("product_id", id),
	)

	product, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		logger.Error("failed to fetch product",
			zap.String("request_id", requestID),
			zap.Int64("product_id", id),
			zap.Error(err),
		)
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	logger.Info("product fetched successfully",
		zap.String("request_id", requestID),
		zap.Int64("product_id", product.ID),
		zap.String("product_name", product.Name),
	)

	respondJSON(w, http.StatusOK, product)
}

// Create создает новый продукт
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode create request",
			zap.String("request_id", requestID),
			zap.Error(err),
		)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		logger.Warn("create product validation failed",
			zap.String("request_id", requestID),
			zap.String("product_name", req.Name),
			zap.Error(err),
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	logger.Debug("creating new product",
		zap.String("request_id", requestID),
		zap.String("product_name", req.Name),
		zap.Float64("price", req.Price),
		zap.Int("stock", req.Stock),
	)

	product, err := h.service.Create(r.Context(), &req)
	if err != nil {
		logger.Error("failed to create product",
			zap.String("request_id", requestID),
			zap.String("product_name", req.Name),
			zap.Error(err),
		)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	logger.Info("product created successfully",
		zap.String("request_id", requestID),
		zap.Int64("product_id", product.ID),
		zap.String("product_name", product.Name),
		zap.Float64("price", product.Price),
		zap.Int("stock", product.Stock),
	)

	respondJSON(w, http.StatusCreated, product)
}

// Update обновляет продукт
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		logger.Warn("invalid product id format",
			zap.String("request_id", requestID),
			zap.String("id", vars["id"]),
			zap.Error(err),
		)
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Error("failed to decode update request",
			zap.String("request_id", requestID),
			zap.Int64("product_id", id),
			zap.Error(err),
		)
		respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.validator.Validate(&req); err != nil {
		logger.Warn("update product validation failed",
			zap.String("request_id", requestID),
			zap.Int64("product_id", id),
			zap.Error(err),
		)
		respondError(w, http.StatusBadRequest, err.Error())
		return
	}

	logger.Debug("updating product",
		zap.String("request_id", requestID),
		zap.Int64("product_id", id),
		zap.String("product_name", req.Name),
	)

	product, err := h.service.Update(r.Context(), id, &req)
	if err != nil {
		logger.Error("failed to update product",
			zap.String("request_id", requestID),
			zap.Int64("product_id", id),
			zap.Error(err),
		)
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	logger.Info("product updated successfully",
		zap.String("request_id", requestID),
		zap.Int64("product_id", product.ID),
		zap.String("product_name", product.Name),
		zap.Float64("price", product.Price),
		zap.Int("stock", product.Stock),
	)

	respondJSON(w, http.StatusOK, product)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	requestID := middleware.GetRequestID(r.Context())

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		logger.Warn("invalid product id format",
			zap.String("request_id", requestID),
			zap.String("id", vars["id"]),
			zap.Error(err),
		)
		respondError(w, http.StatusBadRequest, "invalid product id")
		return
	}

	logger.Debug("deleting product",
		zap.String("request_id", requestID),
		zap.Int64("product_id", id),
	)

	if err := h.service.Delete(r.Context(), id); err != nil {
		logger.Error("failed to delete product",
			zap.String("request_id", requestID),
			zap.Int64("product_id", id),
			zap.Error(err),
		)
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	logger.Info("product deleted successfully",
		zap.String("request_id", requestID),
		zap.Int64("product_id", id),
	)

	w.WriteHeader(http.StatusNoContent)
}

func respondJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	if data != nil {
		response, err := json.Marshal(data)
		if err != nil {
			http.Error(w, "failed to encode response", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(statusCode)
		_, _ = w.Write(response)
	} else {
		w.WriteHeader(statusCode)
	}
}

func respondError(w http.ResponseWriter, statusCode int, message string) {
	respondJSON(w, statusCode, map[string]string{"error": message})
}
