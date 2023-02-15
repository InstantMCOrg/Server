package db

import (
	"github.com/instantminecraft/server/pkg/models"
	"github.com/instantminecraft/server/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

const sessionTokenLength = 32

func Init() {
	dbConnection, err := gorm.Open(sqlite.Open("data.db"), &gorm.Config{})
	db = dbConnection
	if err != nil {
		panic(err)
	}

	// Migrate schemas
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Session{})

	//db.Create(&models.User{Username: "admin", Password: utils.SHA256([]byte("admin"))})
}

// Login searches and returns a ´models.User´ struct if username and the sha256 password matches a record otherwise returns an error
func Login(username string, password string) (models.User, error) {
	// check if user exists
	var user models.User
	err := db.First(&user, "username = ? AND password = ?", username, password).Error
	return user, err
}

// CreateSession Creates a session and returns the token
// If not successful an error is returned
func CreateSession(userModel *models.User) (string, error) {
	sessionToken := utils.RandomString(sessionTokenLength)
	err := db.Create(&models.Session{
		Token:  utils.SHA256([]byte(sessionToken)),
		UserID: int(userModel.ID),
		User:   *userModel,
	}).Error
	return sessionToken, err
}

func GetSession(token string) (models.Session, error) {
	var session models.Session
	err := db.First(&session, "token = ?", token).Error
	return session, err
}
