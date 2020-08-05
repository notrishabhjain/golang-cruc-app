package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
	var cost, oldCost Cost
	cost.ID, _ = strconv.ParseInt(r.FormValue("Id"), 10, 64)
	oldCost.ID, _ = strconv.ParseInt(r.FormValue("Id"), 10, 64)
	err := db.QueryRow("SELECT email, address, city, state, shop_name, category from cost WHERE id=?", oldCost.ID).Scan(&oldCost.Email,
		&oldCost.Address, &oldCost.City, &oldCost.State, &oldCost.ShopName, &oldCost.Category)
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
	if r.Method != "POST" {
		http.Redirect(w, r, "/", 301)
	}
	var customer Customer
	var p1 Products
	var p2, p3, p4, p5 Products
	var tA1, tA2, tA3, tA4, tA5 int64
	var cusID, _ = strconv.ParseInt(r.FormValue("Id"), 10, 64)
	//var cat = r.FormValue("category1")
	var subcat1 = r.FormValue("subcategory1")
	var subcat2 = r.FormValue("subcategory2")
	var subcat3 = r.FormValue("subcategory3")
	var subcat4 = r.FormValue("subcategory4")
	var subcat5 = r.FormValue("subcategory5")
	var q1, _ = strconv.ParseInt(r.FormValue("quantity1"), 10, 64)
	var q2, _ = strconv.ParseInt(r.FormValue("quantity2"), 10, 64)
	var q3, _ = strconv.ParseInt(r.FormValue("quantity3"), 10, 64)
	var q4, _ = strconv.ParseInt(r.FormValue("quantity4"), 10, 64)
	var q5, _ = strconv.ParseInt(r.FormValue("quantity5"), 10, 64)

	err := db.QueryRow("SELECT * from Customer WHERE cusid=?", cusID).Scan(&customer.CusID,
		&customer.ID, &customer.Name, &customer.Email, &customer.Amount,
		&customer.Number, &customer.CreditDate)
	checkInternalServerError(err, w)
	fmt.Println(customer)

	err = db.QueryRow("Select * from products where model=?", subcat1).Scan(&p1.Brand,
		&p1.Model, &p1.Price, &p1.Tax)
	if err == sql.ErrNoRows {
		http.Redirect(w, r, "/list", http.StatusMovedPermanently)
	}
	a1 := q1 * p1.Price

	tA1 = a1 + int64(float64(a1)*p1.Tax)
	fmt.Println(tA1)

	if subcat2 != "" && q2 > 0 {
		err = db.QueryRow("Select * from products where model=?", subcat2).Scan(&p2.Brand,
			&p2.Model, &p2.Price, &p2.Tax)
		if err == sql.ErrNoRows {
			goto pdfgen
		}
		a2 := q2 * p2.Price

		tA2 = a2 + int64(float64(a2)*p2.Tax)
		fmt.Println(tA2)
	}

	if subcat3 != "" && q3 > 0 {
		err = db.QueryRow("Select * from products where model=?", subcat3).Scan(&p3.Brand,
			&p3.Model, &p3.Price, &p3.Tax)
		if err == sql.ErrNoRows {
			goto pdfgen
		}
		a3 := q3 * p3.Price

		tA3 = a3 + int64(float64(a3)*p3.Tax)
		fmt.Println(tA3)
	}
	if subcat4 != "" && q4 > 0 {
		err = db.QueryRow("Select * from products where model=?", subcat4).Scan(&p4.Brand,
			&p4.Model, &p4.Price, &p4.Tax)
		if err == sql.ErrNoRows {
			goto pdfgen
		}
		a4 := q4 * p4.Price

		tA4 = a4 + int64(float64(a4)*p4.Tax)
		fmt.Println(tA4)
	}

	if subcat5 != "" && q5 > 0 {
		err = db.QueryRow("Select * from products where model=?", subcat5).Scan(&p5.Brand,
			&p5.Model, &p5.Price, &p5.Tax)
		if err == sql.ErrNoRows {
			goto pdfgen
		}
		a5 := q5 * p5.Price

		tA5 = a5 + int64(float64(a5)*p5.Tax)
		fmt.Println(tA5)
	}
pdfgen:
	totalAmount := tA1 + tA2 + tA3 + tA4 + tA5
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(100, 7, "Customer Details", "0", 0, "LT", false, 0, "")
	pdf.CellFormat(100, 7, "Retailer Details    ", "0", 0, "RT", false, 0, "")
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(100, 7, strings.Title(customer.Name), "0", 0, "LT", false, 0, "")
	pdf.CellFormat(100, 7, "Mobile Solutions    ", "0", 0, "RT", false, 0, "")
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(100, 7, customer.Email, "0", 0, "LT", false, 0, "")
	pdf.CellFormat(100, 7, "admin@mobile.in    ", "0", 0, "RT", false, 0, "")
	pdf.Ln(-1)
	pdf.SetFont("Arial", "", 12)
	pdf.CellFormat(100, 7, customer.Number, "0", 0, "LT", false, 0, "")
	pdf.CellFormat(100, 7, "+919999888877    ", "0", 0, "RT", false, 0, "")
	pdf.Ln(-1)
	pdf.CellFormat(100, 7, time.Now().Format("02-01-2006"), "0", 0, "LT", false, 0, "")
	pdf.Ln(25)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 7, "PARTICULARS", "0", 0, "CM", false, 0, "")
	pdf.Ln(10)
	pdf.CellFormat(100, 7, "S.No.     Items                                                     Quantity    Price    Tax    Amount", "0", 0, "LT", false, 0, "")
	pdf.Ln(10)
	pdf.SetFont("Arial", "", 14)
	pdf.CellFormat(100, 7, "1."+"            "+strings.Title(p1.Brand)+" "+strings.Title(p1.Model), "0", 0, "LT", false, 0, "")
	pdf.CellFormat(11, 7, "          "+strconv.Itoa(int(q1)), "0", 0, "LT", false, 0, "")
	pdf.CellFormat(20, 7, "              "+strconv.Itoa(int(p1.Price)), "0", 0, "LT", false, 0, "")
	pdf.CellFormat(16, 7, "            "+strconv.FormatFloat(p1.Tax, 'f', 2, 64), "0", 0, "LT", false, 0, "")
	pdf.CellFormat(34, 7, "             "+strconv.Itoa(int(tA1)), "0", 0, "LT", false, 0, "")
	if subcat2 != "" {
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 14)
		pdf.CellFormat(100, 7, "2."+"            "+strings.Title(p2.Brand)+" "+strings.Title(p2.Model), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(11, 7, "          "+strconv.Itoa(int(q2)), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(20, 7, "              "+strconv.Itoa(int(p2.Price)), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(16, 7, "            "+strconv.FormatFloat(p2.Tax, 'f', 2, 64), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(34, 7, "             "+strconv.Itoa(int(tA2)), "0", 0, "LT", false, 0, "")
	}
	if subcat3 != "" {
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 14)
		pdf.CellFormat(100, 7, "3."+"            "+strings.Title(p3.Brand)+" "+strings.Title(p3.Model), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(11, 7, "          "+strconv.Itoa(int(q3)), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(20, 7, "              "+strconv.Itoa(int(p3.Price)), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(16, 7, "            "+strconv.FormatFloat(p3.Tax, 'f', 2, 64), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(34, 7, "             "+strconv.Itoa(int(tA3)), "0", 0, "LT", false, 0, "")
	}
	if subcat4 != "" {
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 14)
		pdf.CellFormat(100, 7, "4."+"            "+strings.Title(p4.Brand)+" "+strings.Title(p4.Model), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(11, 7, "          "+strconv.Itoa(int(q4)), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(20, 7, "              "+strconv.Itoa(int(p4.Price)), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(16, 7, "            "+strconv.FormatFloat(p4.Tax, 'f', 2, 64), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(34, 7, "             "+strconv.Itoa(int(tA4)), "0", 0, "LT", false, 0, "")
	}
	if subcat5 != "" {
		pdf.Ln(10)
		pdf.SetFont("Arial", "", 14)
		pdf.CellFormat(100, 7, "5."+"            "+strings.Title(p5.Brand)+" "+strings.Title(p5.Model), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(11, 7, "          "+strconv.Itoa(int(q5)), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(20, 7, "              "+strconv.Itoa(int(p5.Price)), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(16, 7, "            "+strconv.FormatFloat(p5.Tax, 'f', 2, 64), "0", 0, "LT", false, 0, "")
		pdf.CellFormat(34, 7, "             "+strconv.Itoa(int(tA5)), "0", 0, "LT", false, 0, "")
	}
	pdf.Ln(25)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 7, "                                                                         	  Total: "+"   "+strconv.Itoa(int(totalAmount)), "0", 0, "CM", false, 0, "")
	pdf.Ln(70)
	pdf.SetFont("Arial", "", 14)
	pdf.CellFormat(100, 7, "Thank you for trading with us. Errors and Omissions expected.", "0", 0, "LT", false, 0, "")
	// pdf.SetFont("Arial", "", 10)
	// pdf.CellFormat(190, 7, strconv.Itoa(int(customer.CusID))+" "+customer.Name+" "+customer.Email+
	// 	" "+customer.Amount+" "+customer.Number+" "+customer.CreditDate+subcat1+strconv.Itoa(int(totalAmount)), "0", 0, "CM", false, 0, "")
	pdf.OutputFileAndClose("./download.pdf")
	uploadFile(w, r)
	showFile(w, r)
	// http.ServeFile(w, r, "download.pdf")
	checkInternalServerError(err, w)
	if err != nil {
		panic(err)
	}
	checkInternalServerError(err, w)
	http.Redirect(w, r, "/list", 301)

}

func uploadFile(w http.ResponseWriter, r *http.Request) {
	// Maximum upload of 10 MB files
	// r.ParseMultipartForm(10 << 20)

	// // Get handler for filename, size and headers
	// file, handler, err := r.FormFile("download1.pdf")
	// if err != nil {
	// 	fmt.Println("Error Retrieving the File")
	// 	fmt.Println(err)
	// 	return
	// }

	r.ParseMultipartForm(32 << 20)
	m := r.MultipartForm
	for _, v := range m.File {
		for _, f := range v {
			file, err := f.Open()
			if err != nil {
				fmt.Println(err)
				return
			}

			defer file.Close()
			// fmt.Printf("Uploaded File: %+v\n", handler.Filename)
			// fmt.Printf("File Size: %+v\n", handler.Size)
			// fmt.Printf("MIME Header: %+v\n", handler.Header)

			// Create file
			dst, err := os.Create("download.pdf")
			defer dst.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Copy the uploaded file to the created file on the filesystem
			if _, err := io.Copy(dst, file); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}
	}
	fmt.Fprintf(w, "Successfully Uploaded File\n")
}

func showFile(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "download.pdf")
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	isAuthenticated(w, r)
	http.Redirect(w, r, "/", 301)
}
