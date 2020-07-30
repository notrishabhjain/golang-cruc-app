package main

//Cost ..
type Cost struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Address  string `json:"address"`
	City     string `json:"city"`
	State    string `json:"state"`
	ShopName string `json:"shop_name"`
	Category string `json:"category"`
}

//User ..
type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role     int64  `json:"role"`
}

//Customer ..
type Customer struct {
	ID         int64  `json:"id"`
	CusID      int64  `json:"cusId"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Amount     string `json:"amount"`
	Number     string `json:"number"`
	CreditDate string `json:"credit_date"`
}

//Products ..
type Products struct {
	Brand string  `json:"brand"`
	Model string  `json:"model"`
	Price int64   `json:"price"`
	Tax   float64 `json:"tax"`
}
