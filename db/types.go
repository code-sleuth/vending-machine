package db

// User struct
type User struct {
	UUID     string `json:"uuid"`
	Username string `json:"username"`
	Password string `json:"password,omitempty"`
	Deposit  int    `json:"deposit"`
	Role     string `json:"role"`
}

// Product struct
type Product struct {
	UUID            string `json:"uuid"`
	AmountAvailable int    `json:"amount_available"`
	Cost            int    `json:"cost"`
	ProductName     string `json:"product_name"`
	SellerID        string `json:"seller_id"`
}

// ChangePassword struct
type ChangePassword struct {
	OldPassword     string `json:"oldpassword"`
	NewPassword     string `json:"newpassword"`
	ConfirmPassword string `json:"confirmpassword"`
}

// BuyResponse response to when a user makes a purchase
type BuyResponse struct {
	AmountSpent       int               `json:"amount_spent"`
	ProductName       string            `json:"product_name"`
	ProductsPurchased int               `json:"products_purchased"`
	Change            map[string]string `json:"change"`
}
