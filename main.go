package main

import (
	"booking-app/helper"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"runtime"
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

//var wg = sync.WaitGroup{}

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

	ctx, finish := context.WithCancel(context.Background())
	defer finish()
	input := make(chan *UserData, 1)
	//wg.Add(1)
	for i := 0; i <= 10; i++ {
		go sendTicket(ctx, input, i)
	}

	for {
		var confirmedExit string
		fmt.Println("Do you want to exit [yes/no]?")
		fmt.Scan(&confirmedExit)

		if confirmedExit == "yes" {
			break
		}

		firstName, lastName, email, userTickets := getUserInput()
		isValidName, isValidEmail, isValidTicketNumber := helper.ValidateUserInput(firstName, lastName, email, userTickets, remainingTickets)

		if isValidName && isValidEmail && isValidTicketNumber {

			err := bookTicket(db, input, userTickets, firstName, lastName, email)

			if err != nil {
				log.Fatal(err)
			}

			firstNames := getFirstNames()
			fmt.Printf("The first names of bookings are: %v\n", firstNames)

			if remainingTickets == 0 {
				fmt.Println("Our conference iis booked out. Come back next year.")
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
	}
	//wg.Wait()
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

func bookTicket(storage *sql.DB, output chan *UserData, userTickets uint, firstName, lastName, email string) error {
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
		return err
	}

	userFromDb, err := getData(storage, *lastID)

	fmt.Printf("List of bookings is %v\n", bookings)

	fmt.Printf("Thank you %v %v for bookng %v tickets. You will receive a confirmation email at %v\n", firstName, lastName, userTickets, email)
	fmt.Printf("%v tickets remaining for %v\n", remainingTickets, conferenceName)

	fmt.Printf("UserData: %v", userFromDb)
	output <- userFromDb

	return nil
}

func sendTicket(ctx context.Context, input <-chan *UserData, workerIndex int) {

	timer := time.NewTimer(10 * time.Minute)
	defer func() {
		if !timer.Stop() {
			<-timer.C
		}
		fmt.Printf("End of worker %d\n", workerIndex)
	}()

LOOP:
	for {
		select {
		case <-ctx.Done():
			break LOOP
		case user := <-input:
			var ticket = fmt.Sprintf("%v tickets for %v %v %v", user.numberOfTickets, *user.id, user.firstName, user.lastName)
			fmt.Println("#################")
			fmt.Printf("Worker number %d\n", workerIndex)
			fmt.Printf("Sending ticket:\n %v\n to email address %v\n", ticket, user.email)
			fmt.Println("################")
		case timeOut := <-timer.C:
			fmt.Printf("Timeout happends %s", timeOut.Local())
		default:
			runtime.Gosched()
		}
	}
	//wg.Done()
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
