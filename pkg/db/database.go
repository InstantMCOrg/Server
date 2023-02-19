package db

import (
	"github.com/instantminecraft/server/pkg/config"
	"github.com/instantminecraft/server/pkg/models"
	"github.com/instantminecraft/server/pkg/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"path/filepath"
)

var db *gorm.DB

const sessionTokenLength = 32

func Init() {
	dbPath := filepath.Join(config.DataDir, "data.db")
	dbConnection, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		//Logger: logger.Default.LogMode(logger.Silent), // silent logger
	})
	db = dbConnection
	if err != nil {
		panic(err)
	}

	// Migrate schemas
	db.AutoMigrate(&models.User{})
	db.AutoMigrate(&models.Session{})

	if err := createDefaultAdminUserIfNeeded(); err != nil {
		log.Fatal().Err(err).Msg("Couldn't create default admin user")
	}
}

func createDefaultAdminUserIfNeeded() error {
	var users []models.User
	err := db.Find(&users).Error
	if len(users) == 0 {
		err = db.Create(&models.User{Username: "admin", Password: utils.SHA256([]byte("admin"))}).Error
	}

	return err
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
	token = utils.SHA256([]byte(token))
	err := db.First(&session, "token = ?", token).Error
	return session, err
}

func GetUserFromToken(token string) (models.User, error) {
	var session models.Session
	token = utils.SHA256([]byte(token))
	err := db.Preload("User").First(&session, "token = ?", token).Error
	return session.User, err
}

// UpdatePassword updates the password of the target user and deletes all sessions from this user
func UpdatePassword(user *models.User, newPassword string) error {
	hashedNewPassword := utils.SHA256([]byte(newPassword))
	user.Password = hashedNewPassword
	err := db.Save(&user).Error
	if err != nil {
		return err
	}
	return db.Delete(&models.Session{}, "user_id = ?", user.ID).Error
}
