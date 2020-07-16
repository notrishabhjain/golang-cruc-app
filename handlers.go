package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/jung-kurt/gofpdf"
	"golang.org/x/crypto/bcrypt"
)

func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "template/register.html")
		return
	}
	// grab user info
	username := r.FormValue("username")
	password := r.FormValue("password")
	role := r.FormValue("role")
	// Check existence of user
	var user User
	err := db.QueryRow("SELECT username, password, role FROM users WHERE username=?",
		username).Scan(&user.Username, &user.Password, &user.Role)
	switch {
	// user is available
	case err == sql.ErrNoRows:
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		checkInternalServerError(err, w)
		// insert to database
		_, err = db.Exec(`INSERT INTO users(username, password, role) VALUES(?, ?, ?)`,
			username, hashedPassword, role)
		fmt.Println("Created user: ", username)
		checkInternalServerError(err, w)
	case err != nil:
		http.Error(w, "loi: "+err.Error(), http.StatusBadRequest)
		return
	default:
		http.Redirect(w, r, "/login", http.StatusMovedPermanently)
	}
}

func forgotHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "template/forgot.html")
		return
	}
	// grab user info
	username := r.FormValue("username")
	password := r.FormValue("password")
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	checkInternalServerError(err, w)
	stmt, err := db.Prepare(`UPDATE users SET password=? WHERE username=?`)
	checkInternalServerError(err, w)
	res, err := stmt.Exec(hashedPassword, username)
	checkInternalServerError(err, w)
	_, err = res.RowsAffected()
	checkInternalServerError(err, w)
	http.Redirect(w, r, "/login", 301)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.ServeFile(w, r, "template/login.html")
		return
	}
	// grab user info from the submitted form
	username := r.FormValue("usrname")
	password := r.FormValue("psw")
	// query database to get match username
	var user User
	err := db.QueryRow("SELECT username, password FROM users WHERE username=?",
		username).Scan(&user.Username, &user.Password)
	checkInternalServerError(err, w)
	// validate password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		http.Redirect(w, r, "/login", 301)
	}
	authenticated = true
	if user.Username == "admin123" {
		http.Redirect(w, r, "/listAdmin", 301)
	} else {
		http.Redirect(w, r, "/list", 301)
	}

}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	authenticated = false
	isAuthenticated(w, r)
}

func listAdminHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
	}
	rows, err := db.Query("SELECT * FROM cost")
	checkInternalServerError(err, w)
	var funcMap = template.FuncMap{
		"multiplication": func(n float64, f float64) float64 {
			return n * f
		},
		"addOne": func(n int) int {
			return n + 1
		},
	}
	var costs []Cost
	var cost Cost
	for rows.Next() {
		err = rows.Scan(&cost.ID, &cost.Email,
			&cost.Address, &cost.City, &cost.State, &cost.ShopName, &cost.Category)
		checkInternalServerError(err, w)
		costs = append(costs, cost)
	}
	t, err := template.New("listAdmin.html").Funcs(funcMap).ParseFiles("template/listAdmin.html")
	checkInternalServerError(err, w)
	err = t.Execute(w, costs)
	checkInternalServerError(err, w)

}

func listHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusBadRequest)
	}
	rows, err := db.Query("SELECT * FROM Customer")
	checkInternalServerError(err, w)
	var funcMap = template.FuncMap{
		"addOne": func(n int) int {
			return n + 1
		},
	}
	var customers []Customer
	var customer Customer
	for rows.Next() {
		err = rows.Scan(&customer.CusID, &customer.ID, &customer.Name,
			&customer.Email, &customer.Amount, &customer.Number, &customer.CreditDate)
		checkInternalServerError(err, w)
		customers = append(customers, customer)
	}
	t, err := template.New("list.html").Funcs(funcMap).ParseFiles("template/list.html")
	checkInternalServerError(err, w)
	err = t.Execute(w, customers)
	checkInternalServerError(err, w)

}

func createRetailerHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var cost Cost
	cost.Email = r.FormValue("Email")
	cost.Address = r.FormValue("Address")
	cost.City = r.FormValue("City")
	cost.State = r.FormValue("State")
	cost.ShopName = r.FormValue("ShopName")
	cost.Category = r.FormValue("Category")
	// fmt.Println(cost)

	// Save to database
	stmt, err := db.Prepare(`
		INSERT INTO cost(email, address, city, state, shop_name, category)
		VALUES(?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		fmt.Println("Prepare query error")
		panic(err)
	}
	_, err = stmt.Exec(cost.Email, cost.Address,
		cost.City, cost.State, cost.ShopName, cost.Category)
	if err != nil {
		fmt.Println("Execute query error")
		panic(err)
	}
	http.Redirect(w, r, "/listAdmin", 301)
}

func createCustomerHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var customer Customer
	var user User
	customer.ID = user.ID
	customer.Name = r.FormValue("Name")
	customer.Email = r.FormValue("Email")
	customer.Amount = r.FormValue("Amount")
	customer.Number = r.FormValue("Number")
	customer.CreditDate = r.FormValue("CreditDate")

	// Save to database
	stmt, err := db.Prepare(`
		INSERT INTO Customer(retailerID, name, email, amount, number, creditdate)
		VALUES(?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		fmt.Println("Prepare query error")
		panic(err)
	}
	_, err = stmt.Exec(customer.ID, customer.Name, customer.Email,
		customer.Amount, customer.Number, customer.CreditDate)
	if err != nil {
		fmt.Println("Execute query error")
		panic(err)
	}
	http.Redirect(w, r, "/list", 301)
}

func updateRetailerHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var cost Cost
	cost.ID, _ = strconv.ParseInt(r.FormValue("Id"), 10, 64)
	cost.Email = r.FormValue("Email")
	cost.Address = r.FormValue("Address")
	cost.City = r.FormValue("City")
	cost.State = r.FormValue("State")
	cost.ShopName = r.FormValue("ShopName")
	cost.Category = r.FormValue("Category")
	fmt.Println(cost)
	stmt, err := db.Prepare(`
		UPDATE cost SET email=?, address=?, city=?, state=?, shop_name=?, category=?
		WHERE id=?
	`)
	checkInternalServerError(err, w)
	res, err := stmt.Exec(cost.Email, cost.Address, cost.City, cost.State, cost.ShopName, cost.Category, cost.ID)
	checkInternalServerError(err, w)
	_, err = res.RowsAffected()
	checkInternalServerError(err, w)
	http.Redirect(w, r, "/listAdmin", 301)
}

func updateCustomerHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var customer Customer
	customer.CusID, _ = strconv.ParseInt(r.FormValue("Id"), 10, 64)
	customer.Name = r.FormValue("Name")
	customer.Email = r.FormValue("Email")
	customer.Amount = r.FormValue("Amount")
	customer.Number = r.FormValue("Number")
	customer.CreditDate = r.FormValue("CreditDate")
	//fmt.Println(customer)
	stmt, err := db.Prepare(`
		UPDATE Customer SET name=?, email=?, amount=?, number=?, creditDate=?
		WHERE cusid=?
	`)
	checkInternalServerError(err, w)
	res, err := stmt.Exec(customer.Name, customer.Email, customer.Amount, customer.Number, customer.CreditDate, customer.CusID)
	checkInternalServerError(err, w)
	_, err = res.RowsAffected()
	checkInternalServerError(err, w)
	http.Redirect(w, r, "/list", 301)
}

func deleteRetailerHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var costID, _ = strconv.ParseInt(r.FormValue("Id"), 10, 64)
	stmt, err := db.Prepare("DELETE FROM cost WHERE id=?")
	checkInternalServerError(err, w)
	res, err := stmt.Exec(costID)
	checkInternalServerError(err, w)
	_, err = res.RowsAffected()
	checkInternalServerError(err, w)
	http.Redirect(w, r, "/listAdmin", 301)

}

func deleteCustomerHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var cusID, _ = strconv.ParseInt(r.FormValue("Id"), 10, 64)
	stmt, err := db.Prepare("DELETE FROM Customer WHERE cusid=?")
	checkInternalServerError(err, w)
	res, err := stmt.Exec(cusID)
	checkInternalServerError(err, w)
	_, err = res.RowsAffected()
	checkInternalServerError(err, w)
	http.Redirect(w, r, "/list", 301)

}

func generatePDFHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	if r.Method != "GET" {
		http.Redirect(w, r, "/", 301)
	}
	var customer Customer
	var cusID, _ = strconv.ParseInt(r.FormValue("Id"), 10, 64)
	err := db.QueryRow("SELECT * from Customer WHERE cusid=?", cusID).Scan(&customer.CusID,
		&customer.ID, &customer.Name, &customer.Email, &customer.Amount,
		&customer.Number, &customer.CreditDate)
	checkInternalServerError(err, w)
	fmt.Println(customer)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)
	//pdf.CellFormat(190, 7, strconv.Itoa(int(customer.CusID)), "0", 0, "CM", false, 0, "")
	pdf.CellFormat(190, 7, strconv.Itoa(int(customer.CusID))+" "+customer.Name+" "+customer.Email+
		" "+customer.Amount+" "+customer.Number+" "+customer.CreditDate, "0", 0, "CM", false, 0, "")
	pdf.OutputFileAndClose("download1.pdf")
	http.ServeFile(w, r, "download1.pdf")
	checkInternalServerError(err, w)
	if err != nil {
		panic(err)
	}
	checkInternalServerError(err, w)
	http.Redirect(w, r, "/list", 301)

}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	http.Redirect(w, r, "/", 301)
}
