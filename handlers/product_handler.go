package handlers

import (
	"encoding/json"
	"github.com/code-sleuth/vending-machine/helpers"
	"github.com/code-sleuth/vending-machine/models"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// CreateProduct function
func CreateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}

	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	if r.ContentLength == 0 {
		helpers.ErrorResponse(w, http.StatusBadRequest, "empty json body")
		return
	}

	if len(product.ProductName) == 0 {
		helpers.ErrorResponse(w, http.StatusBadRequest, "product name should not be empty")
		return
	}

	u, err := product.CreateProduct(product.Cost, product.ProductName, product.SellerID)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "unable to create product "+err.Error())
		return
	}
	helpers.JSONResponse(w, http.StatusCreated, u)
}

// GetProducts function
func GetProducts(w http.ResponseWriter, r *http.Request) {
	var p models.Product

	productsList, err := p.GetProducts()
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNoContent, "could not get products from database | "+err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusOK, productsList)
}

// GetProduct by id
func GetProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product

	params := mux.Vars(r)

	uid, err := helpers.ConvertStringToUint(params["id"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, "invalid character in route:"+err.Error())
		return
	}

	product, err := p.GetProduct(uid)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusOK, product)
}

// UpdateProduct function
func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	var product models.Product
	params := mux.Vars(r)

	username, ok := CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, "bad request: "+err.Error())
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Println(err)
		}
	}()

	uid, err := helpers.ConvertStringToUint(params["id"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, "invalid character in route:"+err.Error())
		return
	}

	id, ok := models.UserCanCRUDSeller(username)
	if !ok && uid != id {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to update product")
		return
	}

	u, err := product.UpdateProduct(uid, product.SellerID, product.Cost, product.ProductName)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusAccepted, u)
}

// DeleteProduct function
func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	username, ok := CheckIfUserSessionIsActive(w, r)
	if !ok {
		return
	}

	if _, ok := models.UserCanCRUDSeller(username); !ok {
		helpers.ErrorResponse(w, http.StatusForbidden, "insufficient rights to delete product")
		return
	}
	var product models.Product

	params := mux.Vars(r)

	uid, err := helpers.ConvertStringToUint(params["id"])
	if err != nil {
		helpers.ErrorResponse(w, http.StatusNotFound, "invalid character in route:"+err.Error())
		return
	}

	d, err := product.DeleteProduct(uid)
	if err != nil {
		helpers.ErrorResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	helpers.JSONResponse(w, http.StatusAccepted, map[string]string{"success": d})
}
