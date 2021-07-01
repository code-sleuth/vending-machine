package models

import (
	"errors"
	"fmt"
	"github.com/jinzhu/gorm"
)

// Product table
type Product struct {
	gorm.Model
	AmountAvailable int    `sql:"type:INTEGER;not null"`
	Cost            int    `sql:"type:INTEGER;not null"`
	ProductName     string `sql:"type:VARCHAR(255);not null"`
	SellerID        uint   `sql:"type:INTEGER;not null"`
}

// CreateProduct function
func (p *Product) CreateProduct(cost, amountAvailable int, productName string, sellerId uint) (*Product, error) {
	product := Product{
		AmountAvailable: amountAvailable,
		Cost:            cost,
		ProductName:     productName,
		SellerID:        sellerId,
	}

	tx := db.Begin()

	// save new user to database
	if err := tx.Create(&product).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &product, nil
}

// GetProducts function
func (p *Product) GetProducts() (*[]Product, error) {
	var productList []Product

	// get all users from database
	if err := db.Find(&productList).Error; err != nil {
		return nil, errors.New("error getting products list from database: " + err.Error())
	}

	return &productList, nil
}

// GetProduct function
func (p *Product) GetProduct(id uint) (*Product, error) {
	var product Product

	if err := db.Where("id = ?", id).First(&product).Error; err != nil {
		return nil, errors.New("error getting product from database: " + err.Error())
	}

	return &product, nil
}

// UpdateProduct function
func (p *Product) UpdateProduct(id, sellerID uint, cost, amountAvailable int, productName string) (*Product, error) {
	product, err := p.GetProduct(id)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	product.Cost = cost
	product.AmountAvailable = amountAvailable
	product.ProductName = productName
	product.SellerID = sellerID

	tx := db.Begin()

	if err := db.Save(product).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to update product: " + err.Error())
	}
	tx.Commit()

	return product, nil
}

// DeleteProduct function
func (p *Product) DeleteProduct(id uint) (string, error) {
	product, err := p.GetProduct(id)
	if err != nil {
		return "", err
	}

	tx := db.Begin()
	if err := db.Delete(product).Error; err != nil {
		tx.Rollback()
		return "", errors.New("failed to delete product with id %d, err: " + err.Error())
	}
	tx.Commit()

	return fmt.Sprintf("successfully deleted product with id: %d ", id), nil
}
