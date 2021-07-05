package db

import (
	"crypto/sha1"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

type Service interface {
	RunQuery(db *sqlx.DB, tr *sql.Tx, query string, args ...interface{}) (sql.Result, error)
	ExecuteQuery(db *sqlx.DB, query string, args ...interface{}) (res sql.Result, err error)
	ExecuteTransaction(tr *sql.Tx, query string, args ...interface{}) (res sql.Result, err error)
	PrintQuery(query string, args ...interface{})
	Query(db *sqlx.DB, tr *sql.Tx, query string, args ...interface{}) (*sql.Rows, error)
	QueryNoTr(db *sqlx.DB, query string, args ...interface{}) (rows *sql.Rows, err error)
	QueryWithTr(tr *sql.Tx, query string, args ...interface{}) (rows *sql.Rows, err error)

	CreateUser(userInput *User) (user *User, err error)
	GetUser(uuid string) (user *User, err error)
	UpdateUser(userInput *User) (user *User, err error)
	DeleteUser(uuid string) (err error)
	GetUserPasswordByUsername(username string) (user *User, err error)
	Login(username, password string) (user *User, err error)

	GetProduct(uuid string) (product *Product, err error)
	CreateProduct(pInput *Product) (product *Product, err error)
	UpdateProduct(pInput *Product) (product *Product, err error)
	DeleteProduct(uuid string) (err error)

	Deposit(userUUID string, amount int) (*User, error)
	Buy(userUUID, productUUID string, numberOfProducts int) (buyRes *BuyResponse, err error)
	Reset(userUUID string) (user *User, err error)
}

type service struct {
	db *sqlx.DB
}

// New creates new instance of the database
func New(db *sqlx.DB) Service {
	return &service{
		db: db,
	}
}

// Query - query DB using transaction if provided
func (s *service) Query(db *sqlx.DB, tr *sql.Tx, query string, args ...interface{}) (*sql.Rows, error) {
	if tr == nil {
		return s.QueryNoTr(db, query, args...)
	}
	return s.QueryWithTr(tr, query, args...)
}

// QueryNoTr query db without transaction
func (s *service) QueryNoTr(db *sqlx.DB, query string, args ...interface{}) (rows *sql.Rows, err error) {
	rows, err = db.Query(query, args...)
	if err != nil {
		s.PrintQuery(query, args...)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "QueryNoTr"))
		log.Println("QueryNoTr failed")
	}
	return
}

// QueryWithTr query db with transaction
func (s *service) QueryWithTr(tr *sql.Tx, query string, args ...interface{}) (rows *sql.Rows, err error) {
	rows, err = tr.Query(query, args...)
	if err != nil {
		s.PrintQuery(query, args...)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "QueryWithTr"))
		log.Println("QueryWithTr failed")
	}
	return
}

// RunQuery executes db queries
func (s *service) RunQuery(db *sqlx.DB, tr *sql.Tx, query string, args ...interface{}) (sql.Result, error) {
	if tr == nil {
		return s.ExecuteQuery(db, query, args...)
	}
	return s.ExecuteTransaction(tr, query, args...)
}

// ExecuteQuery runs query without transaction
func (s *service) ExecuteQuery(db *sqlx.DB, query string, args ...interface{}) (res sql.Result, err error) {
	res, err = db.Exec(query, args...)
	if err != nil {
		s.PrintQuery(query, args...)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "ExecuteQuery"))
		log.Println("ExecuteQuery failed")
	}
	return
}

// ExecuteTransaction runs db transaction
func (s *service) ExecuteTransaction(tr *sql.Tx, query string, args ...interface{}) (res sql.Result, err error) {
	res, err = tr.Exec(query, args...)
	if err != nil {
		s.PrintQuery(query, args...)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "ExecuteTransaction"))
		log.Println("ExecuteTransaction failed")
	}
	return
}

// PrintQuery print query that has been executed
func (s *service) PrintQuery(query string, args ...interface{}) {
	str := ""
	if len(args) > 0 {
		for k, v := range args {
			str += fmt.Sprintf("%d:%+v ", k+1, v)
		}
	}
	fmt.Printf("%s\n", query)
	if str != "" {
		fmt.Printf("[%s]\n", str)
	}
}

func (s *service) getPwdBytes(password string) []byte {
	// Return the password as a byte slice
	return []byte(password)
}

func (s *service) hashAndSalt(pwd []byte) string {
	// Use GenerateFromPassword to hash & salt pwd.
	// MinCost is just an integer constant provided by the bcrypt
	// package along with DefaultCost & MaxCost.
	// The cost can be any value you want provided it isn't lower
	// than the MinCost (4)
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Fatal(err)
	}
	// GenerateFromPassword returns a byte slice so we need to
	// convert the bytes to a string and return it
	return string(hash)
}

func (s *service) comparePasswords(hashedPwd string, plainPwd []byte) bool {
	// Since we'll be getting the hashed password from the DB it
	// will be a string so we'll need to convert it to a byte slice
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd)
	if err != nil {
		log.Println(err)
		return false
	}

	return true
}

// GenerateUUID from unique username/product and role of a given user
func (s *service) generateUUID(val1, val2 string) (string, error) {
	data := strings.Join([]string{val1, val2}, ":")
	hash := sha1.New()
	_, err := hash.Write([]byte(data))
	if err != nil {
		return "", err
	}
	hashed := fmt.Sprintf("%x", hash.Sum(nil))
	return hashed, nil
}

// CreateUser creates a new user
func (s *service) CreateUser(userInput *User) (user *User, err error) {
	defer func() {
		log.Println(fmt.Sprintf("CreateUser(exit): username:%+v err:%v", userInput.Username, err))
	}()
	insert := "insert into users(uuid, username, password, deposit, role) select $1, $2, $3, $4, $5"

	uid, err := s.generateUUID(userInput.Username, userInput.Role)
	if err != nil {
		return
	}

	var res sql.Result
	pwd := s.getPwdBytes(userInput.Password)
	res, err = s.RunQuery(s.db, nil, insert, uid, userInput.Username, s.hashAndSalt(pwd), userInput.Deposit, userInput.Role)
	if err != nil {
		return
	}
	rowsAffected := int64(0)
	rowsAffected, err = res.RowsAffected()
	if err != nil {
		return
	}
	if rowsAffected > 1 {
		err = fmt.Errorf("user '%+v' insert affected %d rows", &userInput, rowsAffected)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "CreateUser"))
		return
	} else if rowsAffected == 0 {
		err = fmt.Errorf("create user '%+v' did not affect any rows", userInput)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "CreateUser"))
		return
	}
	user, err = s.GetUser(uid)
	if err != nil {
		return
	}
	return
}

// GetUser get user from db
func (s *service) GetUser(uuid string) (user *User, err error) {
	defer func() {
		log.Println(fmt.Sprintf("GetUser(exit): uuid:%+v err:%v", uuid, err))
	}()
	rows, err := s.Query(
		s.db,
		nil,
		"select uuid, username, deposit, role from users where uuid = $1 limit 1",
		uuid,
	)
	if err != nil {
		return
	}
	user = new(User)
	fetched := false
	for rows.Next() {
		err = rows.Scan(
			&user.UUID,
			&user.Username,
			&user.Deposit,
			&user.Role,
		)
		if err != nil {
			return
		}
		fetched = true
	}

	err = rows.Err()
	if err != nil {
		return
	}

	err = rows.Close()
	if err != nil {
		return
	}
	if !fetched {
		err = fmt.Errorf("cannot find user with uuid '%s'", uuid)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "GetUser"))
		return
	}
	return
}

// GetUserPasswordByUsername get user credentials from db
func (s *service) GetUserPasswordByUsername(username string) (user *User, err error) {
	defer func() {
		log.Println(fmt.Sprintf("GetUserPasswordByUsername(exit): uuid:%+v err:%v", username, err))
	}()
	rows, err := s.Query(
		s.db,
		nil,
		"select uuid, username, password from users where username = $1 limit 1",
		username,
	)
	if err != nil {
		return
	}
	user = new(User)
	fetched := false
	for rows.Next() {
		err = rows.Scan(
			&user.UUID,
			&user.Username,
			&user.Password,
		)
		if err != nil {
			return
		}
		fetched = true
	}

	err = rows.Err()
	if err != nil {
		return
	}

	err = rows.Close()
	if err != nil {
		return
	}
	if !fetched {
		err = fmt.Errorf("cannot find user with username '%s'", username)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "GetUserPasswordByUsername"))
		return
	}
	return
}

// UpdateUser update user details
func (s *service) UpdateUser(userInput *User) (user *User, err error) {
	defer func() {
		log.Println(fmt.Sprintf("UpdateUser(exit): uuid:%+v err:%v", userInput.UUID, err))
	}()
	res, err := s.RunQuery(s.db, nil, "update users set deposit = $1 where uuid = $2", userInput.Deposit, userInput.UUID)
	if err != nil {
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return
	}
	if rowsAffected > 1 {
		err = fmt.Errorf("user '%+v' insert affected %d rows", &user, rowsAffected)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "UpdateUser"))
		return
	} else if rowsAffected == 0 {
		err = fmt.Errorf("update user '%+v' did not affect any rows", user)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "UpdateUser"))
		return
	}
	user, err = s.GetUser(userInput.UUID)
	if err != nil {
		return
	}
	return
}

// Login get user from db
func (s *service) Login(username, password string) (user *User, err error) {
	defer func() {
		log.Println(fmt.Sprintf("Login(exit): username:%+v  err:%v", username, err))
	}()
	user, err = s.GetUserPasswordByUsername(username)
	if err != nil {
		return
	}
	plainPwd := s.getPwdBytes(password)
	pwdMatch := s.comparePasswords(user.Password, plainPwd)
	if !pwdMatch {
		return nil, errors.New("invalid username or password")
	}

	user, err = s.GetUser(user.UUID)
	if err != nil {
		return
	}

	return
}

// DeleteUser delete user details
func (s *service) DeleteUser(uuid string) (err error) {
	defer func() {
		log.Println(fmt.Sprintf("DeleteUser(exit): uuid:%+v err:%v", uuid, err))
	}()
	res, err := s.RunQuery(s.db, nil, "delete from users where uuid = $1", uuid)
	if err != nil {
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return
	}
	if rowsAffected == 0 {
		err = fmt.Errorf("delete user '%+v' did not affect any rows", uuid)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "DeleteUser"))
		return
	}
	return
}

// CreateProduct creates a new product
func (s *service) CreateProduct(pInput *Product) (product *Product, err error) {
	defer func() {
		log.Println(fmt.Sprintf("CreateProduct(exit): productData:%+v err:%v", pInput, err))
	}()
	insert := "insert into products(uuid, amount_available, cost, product_name, seller_id) select $1, $2, $3, $4, $5"

	uid, err := s.generateUUID(pInput.ProductName, pInput.SellerID)
	if err != nil {
		return
	}

	var res sql.Result
	res, err = s.RunQuery(s.db, nil, insert, uid, pInput.AmountAvailable, pInput.Cost, pInput.ProductName, pInput.SellerID)
	if err != nil {
		return
	}
	rowsAffected := int64(0)
	rowsAffected, err = res.RowsAffected()
	if err != nil {
		return
	}
	if rowsAffected > 1 {
		err = fmt.Errorf("product '%+v' insert affected %d rows", &pInput, rowsAffected)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "CreateProduct"))
		return
	} else if rowsAffected == 0 {
		err = fmt.Errorf("create product '%+v' did not affect any rows", pInput)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "CreateProduct"))
		return
	}
	product, err = s.GetProduct(uid)
	if err != nil {
		return
	}
	return
}

// GetProduct get product from db
func (s *service) GetProduct(uuid string) (product *Product, err error) {
	defer func() {
		log.Println(fmt.Sprintf("GetProduct(exit): uuid:%+v  err:%v", uuid, err))
	}()
	rows, err := s.Query(
		s.db,
		nil,
		"select uuid, amount_available, cost, product_name, seller_id from products where uuid = $1 limit 1",
		uuid,
	)
	if err != nil {
		return
	}
	product = new(Product)
	fetched := false
	for rows.Next() {
		err = rows.Scan(
			&product.UUID,
			&product.AmountAvailable,
			&product.Cost,
			&product.ProductName,
			&product.SellerID,
		)
		if err != nil {
			return
		}
		fetched = true
	}

	err = rows.Err()
	if err != nil {
		return
	}

	err = rows.Close()
	if err != nil {
		return
	}
	if !fetched {
		err = fmt.Errorf("cannot find product with uuid '%s'", uuid)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "GetProduct"))
		return
	}
	return
}

// UpdateProduct update product details
func (s *service) UpdateProduct(pInput *Product) (product *Product, err error) {
	defer func() {
		log.Println(fmt.Sprintf("UpdateProduct(exit): uuid:%+v err:%v", pInput.UUID, err))
	}()
	res, err := s.RunQuery(s.db, nil, "update products set amount_available = $1, cost = $2, product_name = $3 where uuid = $4",
		pInput.AmountAvailable, pInput.Cost, pInput.ProductName, pInput.UUID)
	if err != nil {
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return
	}
	if rowsAffected > 1 {
		err = fmt.Errorf("product '%+v' insert affected %d rows", &product, rowsAffected)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "UpdateProduct"))
		return
	} else if rowsAffected == 0 {
		err = fmt.Errorf("update product '%+v' did not affect any rows", pInput)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "UpdateProduct"))
		return
	}
	product, err = s.GetProduct(pInput.UUID)
	if err != nil {
		return
	}
	return
}

// DeleteProduct delete product details
func (s *service) DeleteProduct(uuid string) (err error) {
	defer func() {
		log.Println(fmt.Sprintf("DeleteProductHandler(exit): uuid:%+v err:%v", uuid, err))
	}()
	res, err := s.RunQuery(s.db, nil, "delete from products where uuid = $1", uuid)
	if err != nil {
		return
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return
	}
	if rowsAffected == 0 {
		err = fmt.Errorf("delete product '%+v' did not affect any rows", uuid)
		err = errors.New(fmt.Sprintf("%+v: %+v", err.Error(), "DeleteProductHandler"))
		return
	}
	return
}

// Deposit amount of coins on users account
func (s *service) Deposit(userUUID string, amount int) (user *User, err error) {
	defer func() {
		log.Println(fmt.Sprintf("Deposit(exit): userUUID:%+v amount:%+v err:%v", userUUID, amount, err))
	}()
	acceptedDenominations := []int{5, 10, 20, 50, 100}
	if ok := s.Find(acceptedDenominations, amount); !ok {
		errString := fmt.Sprintf("[%+v] is not in the acceptable denominations: use one of the following %+v", amount, acceptedDenominations)
		return nil, errors.New(errString)
	}
	user, err = s.GetUser(userUUID)
	if err != nil {
		return nil, err
	}

	user.Deposit = user.Deposit + amount

	user, err = s.UpdateUser(user)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Find returns an true if an int value is available in a slices of integers
func (s *service) Find(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

// Buy buy products
func (s *service) Buy(userUUID, productUUID string, numberOfProducts int) (buyRes *BuyResponse, err error) {
	defer func() {
		log.Println(fmt.Sprintf("Buy(exit): userUUID:%+v productUUID:%+v numberOfProducts:%+v err:%v", userUUID, productUUID, numberOfProducts, err))
	}()
	product, err := s.GetProduct(productUUID)
	if err != nil {
		return
	}

	if numberOfProducts > product.AmountAvailable {
		errString := fmt.Sprintf("requested amount %+v is greater than available amout %+v", numberOfProducts, product.AmountAvailable)
		return nil, errors.New(errString)
	}
	user, err := s.GetUser(userUUID)
	if err != nil {
		return nil, err
	}
	amountToSpend := numberOfProducts * product.Cost
	transaction := user.Deposit - amountToSpend
	if transaction < 0 {
		errString := fmt.Sprintf("insufficient funds to spend [%+v], available balance is [%+v]", amountToSpend, user.Deposit)
		return nil, errors.New(errString)
	}

	user.Deposit = user.Deposit - amountToSpend
	change := user.Deposit
	remainingProducts := product.AmountAvailable - numberOfProducts
	if remainingProducts < 0 {
		remainingProducts = 0
	}
	product.AmountAvailable = remainingProducts
	product, err = s.UpdateProduct(product)
	if err != nil {
		return
	}

	// set deposit to 0 since change is going to be returned to the user
	user.Deposit = 0
	_, err = s.UpdateUser(user)
	if err != nil {
		return
	}

	// TODO: handle change
	changeSlice := make(map[string]string, 0)
	acceptedDenominations := []int{5, 10, 20, 50, 100}
	sort.Slice(acceptedDenominations, func(i, j int) bool {
		return acceptedDenominations[i] > acceptedDenominations[j]
	})
	for _, denomination := range acceptedDenominations {
		quotient, remainder := s.divisionAndModulus(int64(change), int64(denomination))
		if quotient > 0 {
			str := fmt.Sprintf("denomination: %+v", denomination)
			changeSlice[str] = fmt.Sprintf("number of coins: %+v", quotient)
		}
		change = int(remainder)
	}
	if change != 0 {
		changeSlice["no supported denomination for change: "] = fmt.Sprintf(" %+v", change)
	}

	return &BuyResponse{
		AmountSpent:       amountToSpend,
		ProductName:       product.ProductName,
		ProductsPurchased: numberOfProducts,
		Change:            changeSlice,
	}, nil
}

func (s *service) divisionAndModulus(numerator, denominator int64) (quotient, remainder int64) {
	quotient = numerator / denominator // integer division, decimals are truncated
	remainder = numerator % denominator
	return
}

// Reset resets users deposit
func (s *service) Reset(userUUID string) (user *User, err error) {
	defer func() {
		log.Println(fmt.Sprintf("Reset(exit): userUUID:%+v  err:%v", userUUID, err))
	}()
	user, err = s.GetUser(userUUID)
	if err != nil {
		return
	}

	user.Deposit = 0

	user, err = s.UpdateUser(user)
	if err != nil {
		return
	}
	return
}
