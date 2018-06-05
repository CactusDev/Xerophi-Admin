package main

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/CactusDev/Xerophi/redis"
	"github.com/CactusDev/Xerophi/rethink"
	"github.com/CactusDev/Xerophi/secure"
	"github.com/CactusDev/Xerophi/user"

	mapstruct "github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
)

var config Config
var redisConn *redis.Connection
var rethinkConn *rethink.Connection
var empty string

const menuString = `
    +-=============================-+
    |                               |
    |         Xerophi Admin         |
    |                               |
    +-------------------------------+
    | Option |      Information     |
    +-------------------------------+
    |   a   | Adds a new user       |
    |   u   | Updates user's info   |				
    |   r   | Removes a user        |
    |   q   | Quit                  |
    +-=============================-+
    `

// λ is a dope character, why not use it?
const prompt = "λ | "

const displayUserDetails = `
  Record retrieved:
    ID: %s
    Token: %s
    Username: %s
    User ID: %d
    Service: %s
    
  `

type localUser struct {
	Hash      string  `json:"hash"`
	Token     string  `json:"token"`
	UserID    int     `json:"userId"`
	UserName  string  `json:"userName"`
	Service   string  `json:"service"`
	CreatedAt string  `json:"createdAt"`
	DeletedAt float64 `json:"deletedAt"`
}

func pauseForInput() {
	fmt.Scanln(&empty)
}

func getData(msg string) localUser {
	var userName, password, token, service string

	fmt.Println(msg)

	// Get username
	fmt.Print("Username: ")
	fmt.Scanln(&userName)

	// Get the user token
	fmt.Print("Token: ")
	fmt.Scanln(&token)

	// Get password
	fmt.Print("Password: ")
	fmt.Scanln(&password)

	// Get service
	fmt.Print("Service: ")
	fmt.Scanln(&service)

	return localUser{
		UserName: userName,
		Token:    token,
		Hash:     string(secure.HashArgon(password)),
		Service:  service,
	}
}

func addUser() bool {
	clearScreen()

	newUser := getData("New user info")

	// Get User ID
	id, err := rethinkConn.GetTotalRecords("users", nil)
	if err != nil {
		log.Error(err)
		return false
	}

	// Update user object
	newUser.CreatedAt = time.Now().Format(time.RFC3339)
	newUser.DeletedAt = 0.0
	newUser.UserID = id

	// Add object to users table in DB
	rethinkConn.Create("users")

	log.Infof("Succesfully added user %+v", newUser)

	// Done
	return true
}

func updateUser() bool {
	clearScreen()

	// Get the new info
	updateVals := getData(
		"Update info. Leave empty to leave the same, token is required")

	if updateVals.Token == "" {
		log.Error("Token required!")
		return false
	}

	// Retrieve the user by token
	res, err := rethinkConn.GetSingle(
		"users", map[string]interface{}{"token": updateVals.Token})

	// Put it into a user oject
	var dbVals user.Database
	mapstruct.Decode(res, &dbVals)

	// Update any non-empty fields
	dbVals.Token = updateVals.Token // Always set, required
	if updateVals.Hash != "" {
		dbVals.Hash = updateVals.Hash
	}
	if updateVals.UserName != "" {
		dbVals.UserName = updateVals.UserName
	}
	if updateVals.Service != "" {
		dbVals.Service = updateVals.Service
	}

	var toUpdateMap map[string]interface{}
	mapstruct.Decode(dbVals, &toUpdateMap)

	// Update the record
	_, err = rethinkConn.Update("users", dbVals.ID, toUpdateMap)
	if err != nil {
		log.Error(err)
		return false
	}

	log.Info("Succesfully updated user ID: %d, token: %s",
		dbVals.UserID, dbVals.Token)
	return true
}

func removeUser() bool {
	var token string
	clearScreen()

	// Get token to delete
	fmt.Print("Token: ")
	fmt.Scanln(&token)

	// Retrieve record from token
	res, err := rethinkConn.GetSingle(
		"users", map[string]interface{}{"token": token})
	if err != nil {
		log.Error(err)
		return false
	}

	// Decode into struct
	var dbVals user.Database
	mapstruct.Decode(res, &dbVals)

	// Print record to confirm
	fmt.Printf(displayUserDetails,
		dbVals.ID, dbVals.Token, dbVals.UserName, dbVals.UserID, dbVals.Service)

	// Confirm delete
	var confirmDel string
	fmt.Print("Confirm deletion [y/n]: ")
	fmt.Scanln(&confirmDel)

	if strings.ToLower(confirmDel) != "y" {
		// Any other character than pressing an
		log.Warn("Didn't receive confirmation, not removing")
		return true
	} else {
		// Hard delete
		fmt.Print("Hard deletion [y/n]: ")
		fmt.Scanln(&confirmDel)
		if strings.ToLower(confirmDel) != "y" {
			// Soft deletion
			_, err = rethinkConn.Disable("users", dbVals.ID)
		} else {
			// Hard deletion
			_, err = rethinkConn.Delete("users", dbVals.ID)
		}
		// Failed
		if err != nil {
			log.Error(err)
			return false
		}
	}

	log.Infof("Succesfully removed %s, UUID: %s", dbVals.Token, dbVals.ID)
	return true
}

func clearScreen() {
	var cmd = exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func init() {
	log.SetLevel(log.DebugLevel)
	// TODO: Change this actually find Xerophi and read instead of hard-coded
	config = LoadConfigFromPath(
		build.Default.GOPATH + "/src/github.com/CactusDev/Xerophi/config.json")
}

func main() {
	// TODO: Make this have a CLI UI
	log.Info("Starting up...")

	// Initialize connection to RethinkDB
	log.Info("Connecting to RethinkDB...")
	rethinkConn = &rethink.Connection{
		DB:   config.Rethink.DB,
		Opts: config.Rethink.Connection,
	}
	// Validate connection
	if err := rethinkConn.Connect(); err != nil {
		log.Fatal("RethinkDB Connection Failed! - ", err)
	}
	log.Info("Success!")

	// Initialize connection Redis
	log.Info("Connecting to Redis...")
	redisConn = &redis.Connection{
		DB:   config.Redis.DB,
		Opts: config.Redis.Connection,
	}
	// Validate connection
	if err := redisConn.Connect(); err != nil {
		log.Fatal("Redis Connection Failed! - ", err)
	}
	log.Info("Success!")

	// Display the menu
	for {
		// Clear the screen
		clearScreen()

		// Display the menu and prompt
		fmt.Println(menuString)
		fmt.Print(prompt)

		// Get the input
		var input string
		fmt.Scanln(&input)

		// Move down a line
		fmt.Println()

		// Check if we're quitting now
		if input == "q" {
			break
		}

		// If it's valid, call the appropriate function
		switch input {
		case "a":
			// Add a new user
			addUser()
		case "u":
			// Update a user
			updateUser()
		case "r":
			// Remove a user
			removeUser()
		default:
			// Some other random thing
			// Display an error
			fmt.Printf("Invalid option \"%s\".", input)
		}

		// Wait for acknowledgement
		fmt.Println("Press enter to continue...")
		pauseForInput()
	}

	// Cleanup code here
	log.Info("Closing RethinkDB connection...")
	if err := rethinkConn.Close(); err != nil {
		log.Error("Failed to close connection!")
		log.Error(err)
	} else {
		log.Info("Success!")
	}

	log.Info("Closing Redis connection...")
	if err := redisConn.Session.Close(); err != nil {
		log.Error("Failed to close connection!")
		log.Error(err)
	} else {
		log.Info("Success!")
	}
}
