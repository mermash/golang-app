package main

import (
	"booking-app/helper"
	"database/sql"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	conferenceTickets int = 50
)

var conferenceName = os.Getenv("CONFNAME")
var remainingTickets uint = 50
var bookings = make([]UserData, 0)

var dbUser = os.Getenv("DB_USER")
var dbPassword = os.Getenv("DB_PASSWORD")
var dbHost = os.Getenv("DB_HOST")
var dbPort = os.Getenv("DB_PORT")
var dbDB = os.Getenv("DB_DB")

type UserData struct {
	id                     *uint
	firstName              string
	lastName               string
	email                  string
	numberOfTickets        uint
	isOptedInForNewsLetter uint
}

var wg = sync.WaitGroup{}

func main() {

	greetUsers()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPassword, dbHost, dbPort, dbDB)
	fmt.Println(dsn)
	db, err := sql.Open("mysql", dsn)
	defer db.Close()

	if err != nil {
		log.Fatal(err)
	}

	db.SetMaxOpenConns(10)

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	firstName, lastName, email, userTickets := getUserInput()
	isValidName, isValidEmail, isValidTicketNumber := helper.ValidateUserInput(firstName, lastName, email, userTickets, remainingTickets)

	if isValidName && isValidEmail && isValidTicketNumber {

		lastID, err := bookTicket(db, userTickets, firstName, lastName, email)

		if err != nil {
			log.Fatal(err)
		}

		userFromDb, err := getData(db, *lastID)

		if err != nil {
			log.Fatal(err)
		}

		wg.Add(1)
		go sendTicket(userFromDb)

		firstNames := getFirstNames()
		fmt.Printf("The first names of bookings are: %v\n", firstNames)

		if remainingTickets == 0 {
			// end program
			fmt.Println("Our conference iis booked out. Come back next year.")
			//break
		}

	} else {
		if !isValidName {
			fmt.Println("first name or last name you entered is too short")
		}
		if !isValidEmail {
			fmt.Println("email address you entered doesn't contain @ sign")
		}
		if !isValidTicketNumber {
			fmt.Println("number of tickets you entered is invalid")
		}
	}
	wg.Wait()
}

func greetUsers() {
	fmt.Printf("Welcome to %v booking application\n", conferenceName)
	fmt.Printf("We have total of %v tickets and %v are still available\n", conferenceTickets, remainingTickets)
	fmt.Println("Get your tickets here to attend")
}

func getFirstNames() []string {
	firstNames := []string{}
	for _, booking := range bookings {
		firstNames = append(firstNames, booking.firstName)
	}

	return firstNames
}

func getUserInput() (string, string, string, uint) {
	var firstName string
	var lastName string
	var email string
	var userTickets uint
	// ask user for their name
	fmt.Println("Enter your first name: ")
	fmt.Scan(&firstName)

	fmt.Println("Enter your last name: ")
	fmt.Scan(&lastName)

	fmt.Println("Enter your email address: ")
	fmt.Scan(&email)

	fmt.Println("Enter number of tickets: ")
	fmt.Scan(&userTickets)

	return firstName, lastName, email, userTickets
}

func bookTicket(storage *sql.DB, userTickets uint, firstName string, lastName string, email string) (*int64, error) {
	remainingTickets = remainingTickets - userTickets

	var userData = UserData{
		firstName:       firstName,
		lastName:        lastName,
		email:           email,
		numberOfTickets: userTickets,
	}

	bookings = append(bookings, userData)
	lastID, err := insertData(storage, &userData)

	if err != nil {
		return nil, err
	}

	fmt.Printf("List of bookings is %v\n", bookings)

	fmt.Printf("Thank you %v %v for bookng %v tickets. You will receive a confirmation email at %v\n", firstName, lastName, userTickets, email)
	fmt.Printf("%v tickets remaining for %v\n", remainingTickets, conferenceName)

	return lastID, nil
}

func sendTicket(user *UserData) {
	time.Sleep(2 * time.Second)
	var ticket = fmt.Sprintf("%v tickets for %v %v %v", user.numberOfTickets, *user.id, user.firstName, user.lastName)
	fmt.Println("#################")
	fmt.Printf("Sending ticket:\n %v\n to email address %v\n", ticket, user.email)
	fmt.Println("################")
	wg.Done()
}

func insertData(db *sql.DB, user *UserData) (*int64, error) {
	query := "INSERT INTO `users` (first_name, last_name, email, number_of_tickets) VALUES(?, ?, ?, ?)"

	insert, err := db.Prepare(query)
	defer insert.Close()

	if err != nil {
		return nil, err
	}

	result, err := insert.Exec(user.firstName, user.lastName, user.email, user.numberOfTickets)

	if err != nil {
		return nil, err
	}

	lastID, err := result.LastInsertId()

	if err != nil {
		return nil, err
	}

	return &lastID, nil
}

func getData(db *sql.DB, id int64) (*UserData, error) {
	user := &UserData{}

	row := db.QueryRow("SELECT * FROM users WHERE id = ?", id)

	err := row.Scan(&user.id, &user.firstName, &user.lastName, &user.email, &user.numberOfTickets)

	if err != nil {
		return nil, err
	}

	return user, nil
}
