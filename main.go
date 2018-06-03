package main

import (
	"fmt"
	"go/build"
	"os"
	"os/exec"

	"github.com/CactusDev/Xerophi/redis"
	"github.com/CactusDev/Xerophi/rethink"

	log "github.com/sirupsen/logrus"
)

var config Config

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
    |   r   | Removes a user	    |
    |   g   | Generate an admin     |	
    | ----- | 	JWT auth token      |
    |   q   | Quit                  |
    +-=============================-+
    `

// λ is a dope character, why not use it?
const prompt = "λ | "

func addUser()    {}
func updateUser() {}
func removeUser() {}
func genToken()   {}

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
	rethinkConn := &rethink.Connection{
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
	redisConn := &redis.Connection{
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
		case "g":
			// Generate an admin/root JWT auth token
			genToken()
		default:
			// Some other random thing
			// Display an error
			fmt.Printf("Invalid option \"%s\".", input)
		}
		fmt.Println("Press enter to continue...")
		// Wait for acknowledgement
		fmt.Scanln(nil)
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
