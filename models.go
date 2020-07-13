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
