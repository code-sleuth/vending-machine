package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/code-sleuth/vending-machine/helpers"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

var db = dbConnect(os.Getenv("ENVIRONMENT"))

// implement enum for role
type userRole string

func (r *userRole) Scan(value interface{}) error {
	*r = userRole(value.([]byte))
	return nil
}

func (r *userRole) Value() (driver.Value, error) {
	return string(*r), nil
}

// implement enum for access
type userAccess string

func (a *userAccess) Scan(value interface{}) error {
	*a = userAccess(value.([]byte))
	return nil
}

func (a userAccess) Value() (driver.Value, error) {
	return string(a), nil
}

// User defines the structure of the users of the system.
type User struct {
	gorm.Model
	Username string   `sql:"unique;unique_index;not null"`
	Password string   `sql:"type:VARCHAR(255);not null"`
	Deposit  int      `sql:"type:VARCHAR(255);not null"`
	Role     userRole `sql:"type:user_role;not null;DEFAULT:'buyer'"`
}

// CreateUser function
func (u *User) CreateUser(name, password, role string, deposit int) (*User, error) {
	pwd := getPwdBytes(password)
	user := User{
		Username: name,
		Password: hashAndSalt(pwd),
		Deposit:  deposit,
		Role:     userRole(role),
	}

	tx := db.Begin()

	// save new user to database
	if err := tx.Create(&user).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	tx.Commit()

	return &user, nil
}

// GetUsers function
func (u *User) GetUsers() (*[]User, error) {
	var userList []User

	// get all users from database
	if err := db.Find(&userList).Error; err != nil {
		return nil, errors.New("error getting users list from database: " + err.Error())
	}

	return &userList, nil
}

// GetUser function
func (u *User) GetUser(id uint) (*User, error) {
	var user User

	if err := db.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, errors.New("error getting user from database: " + err.Error())
	}

	return &user, nil
}

// UpdateUser function
func (u *User) UpdateUser(id uint, username, role string, deposit int) (*User, error) {
	user, err := u.GetUser(id)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	user.Username = username
	user.Role = userRole(role)
	user.Deposit = deposit

	tx := db.Begin()

	if err := db.Save(user).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to update user: " + err.Error())
	}
	tx.Commit()

	return user, nil
}

//ChangePassword func
func (u *User) ChangePassword(id uint, oldPassword, newPassword, confirmNewPassword string) (*User, error) {
	user, err := u.GetUser(id)
	if err != nil {
		return nil, errors.New(err.Error())
	}

	pwdMatch := comparePasswords(user.Password, getPwdBytes(oldPassword))
	if !pwdMatch {
		return nil, errors.New("old password does not match")
	}

	if !(newPassword == confirmNewPassword) {
		return nil, errors.New("new passwords and confirm password do not match")
	}

	user.Password = hashAndSalt(getPwdBytes(newPassword))

	tx := db.Begin()
	if err := db.Save(user).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to change password: " + err.Error())
	}
	tx.Commit()
	return user, nil
}

// DeleteUser function
func (u *User) DeleteUser(id uint) (string, error) {
	user, err := u.GetUser(id)
	if err != nil {
		return "", err
	}

	tx := db.Begin()
	// don't dereference user because GetUser returns a pointer
	if err := db.Delete(user).Error; err != nil {
		tx.Rollback()
		return "", errors.New("failed to delete user with id %d, err: " + err.Error())
	}
	tx.Commit()

	return fmt.Sprintf("successfully deleted user with id: %d ", id), nil

}

// Login function
func (u *User) Login(email, password string) (string, error) {
	var user User
	plainPwd := getPwdBytes(password)

	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return "", errors.New("user not in database: " + err.Error())
	}

	pwdMatch := comparePasswords(user.Password, plainPwd)
	if !pwdMatch {
		return "", errors.New("invalid username or password")
	}

	token, err := helpers.GenerateJWT(user.ID, user.Username)
	if err != nil {
		return "", errors.New("failed to create token: " + err.Error())
	}

	return token, nil
}

// GetUserByUsername function
func (u *User) GetUserByUsername(username string) (*User, error) {
	var user User

	if err := db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, errors.New("error getting user from database: " + err.Error())
	}

	return &user, nil
}

func getPwdBytes(password string) []byte {
	// Return the password as a byte slice
	return []byte(password)
}

func hashAndSalt(pwd []byte) string {
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

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
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

// UserCanCRUD function
func UserCanCRUD(username string) (uint, bool) {
	var u User

	user, err := u.GetUserByUsername(username)
	if err != nil {
		return 0, false
	}

	if user.Role == "buyer" {
		return user.ID, true
	}

	return 0, false
}

// ChangePassword struct
type ChangePassword struct {
	OldPassword     string `json:"oldpassword"`
	NewPassword     string `json:"newpassword"`
	ConfirmPassword string `json:"confirmpassword"`
}