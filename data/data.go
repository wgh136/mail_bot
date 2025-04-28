package data

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"runtime"
)

var (
	db *sql.DB
)

func init() {
	baseDir := ""
	var err error
	if runtime.GOOS == "linux" {
		baseDir = "/var/lib"
	} else {
		baseDir, err = os.UserHomeDir()
		if err != nil {
			panic(err)
		}
	}
	dataDir := baseDir + "/.mail_bot"
	if _, err := os.Stat(dataDir); os.IsNotExist(err) {
		if err := os.Mkdir(dataDir, 0755); err != nil {
			panic(err)
		}
	}
	db, err = sql.Open("sqlite3", dataDir+"/data.db")
	if err != nil {
		panic(err)
	}
	createTable()
}

func createTable() {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS mail (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		userId INTEGER NOT NULL,
		email TEXT NOT NULL,
		type TEXT NOT NULL,
		config TEXT NOT NULL
	);
	CREATE UNIQUE INDEX IF NOT EXISTS idx_mail_userId_email ON mail (userId, email);
	`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS user (
    	userId INTEGER PRIMARY KEY,
    	chatId INTEGER NOT NULL
	)`)
	if err != nil {
		panic(err)
	}
}

func AddMail(userId int64, email string, config MailConfig) {
	_, err := db.Exec("INSERT INTO mail (userId, email, type, config) VALUES (?, ?, ?, ?)", userId, email, config.MailType(), config.ToString())
	if err != nil {
		panic(err)
	}
}

func GetMail(userId int64) ([]UserEmailWithConfig, error) {
	rows, err := db.Query("SELECT email, type, config FROM mail WHERE userId = ?", userId)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	var mails []UserEmailWithConfig
	for rows.Next() {
		var email string
		var configType string
		var configData string
		if err := rows.Scan(&email, &configType, &configData); err != nil {
			return nil, err
		}

		var config MailConfig
		switch configType {
		case "imap":
			config = &ImapConfig{}
		default:
			continue
		}
		if err := config.FromString(configData); err != nil {
			return nil, err
		}
		mails = append(mails, UserEmailWithConfig{
			UserId: userId,
			Email:  email,
			Config: config,
		})
	}
	return mails, nil
}

func DeleteMail(userId int64, email string) {
	_, err := db.Exec("DELETE FROM mail WHERE userId = ? AND email = ?", userId, email)
	if err != nil {
		panic(err)
	}
}

func AddUser(userId int64, chatId int64) {
	_, err := db.Exec("INSERT OR REPLACE INTO user (userId, chatId) VALUES (?, ?)", userId, chatId)
	if err != nil {
		panic(err)
	}
}

func GetUserChatId(userId int64) (int64, error) {
	var chatId int64
	err := db.QueryRow("SELECT chatId FROM user WHERE userId = ?", userId).Scan(&chatId)
	if err != nil {
		return 0, err
	}
	return chatId, nil
}

func GetAllConfigs() []UserEmailWithConfig {
	rows, err := db.Query("SELECT userId, email, type, config FROM mail")
	if err != nil {
		panic(err)
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	var configs []UserEmailWithConfig
	for rows.Next() {
		var userId int64
		var email string
		var configType string
		var configData string
		if err := rows.Scan(&userId, &email, &configType, &configData); err != nil {
			panic(err)
		}

		var config MailConfig
		switch configType {
		case "imap":
			config = &ImapConfig{}
		default:
			continue
		}
		if err := config.FromString(configData); err != nil {
			panic(err)
		}

		configs = append(configs, UserEmailWithConfig{
			UserId: userId,
			Email:  email,
			Config: config,
		})
	}
	return configs
}

func ExistsMail(userId int64, email string) bool {
	rows, err := db.Query("SELECT COUNT(*) FROM mail WHERE userId = ? AND email = ?", userId, email)
	if err != nil {
		return false
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			log.Println(err)
		}
	}(rows)

	var count int
	if rows.Next() {
		if err := rows.Scan(&count); err != nil {
			return false
		}
	}
	return count > 0
}

func CountTotalMails() (int, error) {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM mail").Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
