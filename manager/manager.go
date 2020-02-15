package main

import (
	"database/sql"
	"fmt"
	"github.com/AlisherFozilov/ib-core/pkg/core"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"os"
	"strings"
)

var instream *os.File = os.Stdin
var outstream *os.File = os.Stdout

func startLogging() {
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(file)
	log.Print("start application")
	log.Print("open db")
}
func openAndInitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "db.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	err = core.Init(db)
	if err != nil {
		log.Fatal(err)
	}
	return db
}
func main() {
	startLogging()
	db := openAndInitDB()

	_, err := fmt.Fprintln(outstream, "Добро пожаловать в наше приложение")
	if err != nil {
		log.Println(err)
	}
	log.Print("start operations loop")
	operationsLoop(db, unauthorizedOperations, unauthorizedOperationsLoop)
	log.Print("finish operations loop")
	log.Print("finish application")
}

func operationsLoop(db *sql.DB, commands string, loop func(db *sql.DB, cmd string) bool) {
	for {
		fmt.Fprintln(outstream, commands)
		var cmd string
		_, err := fmt.Fscan(instream, &cmd)
		if err != nil {
			log.Fatalf("Can't read input: %v", err)
		}
		if exit := loop(db, strings.TrimSpace(cmd)); exit {
			return
		}
	}
}

func unauthorizedOperationsLoop(db *sql.DB, cmd string) (exit bool) {
	switch cmd {
	case "1":
		ok, err := handleLogin(db)
		if err != nil {
			log.Printf("can't handle login: %v", err)
			return false
		}
		if !ok {
			fmt.Fprintln(outstream, "Неправильно введён логин или пароль. Попробуйте ещё раз.")
			return false
		}
		operationsLoop(db, authorizedOperations, authorizedOperationsLoop)
	case "q":
		return true
	default:
		fmt.Fprintf(outstream, "Вы выбрали неверную команду: %s\n", cmd)
	}

	return false
}
func authorizedOperationsLoop(db *sql.DB, cmd string) (exit bool) {
	var err error
	switch cmd {
	case "1":
		err = handleAddClient(db)
		printBadOrGood("", "Клиент успешно добавлен!", err)
	case "2":
		err = handleAddBankAccountToClient(db)
		printBadOrGood("", "Счёт успешно добавлен!", err)
	case "3":
		serviceNumber, err := handleAddService(db)
		printBadOrGood("", "Сервис успешно добавлен!", err)
		fmt.Fprintf(outstream, "Номер вашего сервиса: %v\n", serviceNumber)
	case "4":
		err = handleAddATM(db)
		printBadOrGood("", "Банкомат успешно добавлен!", err)
	case "5":
		operationsLoop(db, formats, exportLoop)
	case "6":
		operationsLoop(db, formats, importLoop)
	case "7":
		err := handleAddManager(db)
		printBadOrGood("", "Менеджер успешно добавлен!", err)
	case "8":
		err := handleReplenishBankAccount(db)
		printBadOrGood("", "Баланс успешно пополнен!", err)
	case "q":
		return true
	default:
		fmt.Fprintf(outstream, "Вы выбрали неверную команду: %s\n", cmd)
	}

	return false
}

func handleReplenishBankAccount(db *sql.DB) error {
	var clientId, accountNumber, amount int64
	fmt.Fprintln(outstream, "Введите id клиента")
	fmt.Fscan(instream, &clientId)
	fmt.Fprintln(outstream, "Введите номер счёта клиента")
	fmt.Fscan(instream, &accountNumber)
	fmt.Fprintln(outstream, "Введите сумму пополнения")
	fmt.Fscan(instream, &amount)
	err := core.ReplenishBankAccount(clientId, accountNumber, amount, db)
	return err
}

func importLoop(db *sql.DB, cmd string) bool {
	switch cmd {
	case "1":
		operationsLoop(db, exportImportData, importJsonLoop)
	case "2":
		operationsLoop(db, exportImportData, importXmlLoop)
	case "q":
		return true
	}
	return false
}
func importXmlLoop(db *sql.DB, cmd string) bool {
	var err error
	switch cmd {
	case "1":
		err = core.ImportClientsFromXML(db)
		printBadOrGood("", "Успешный импорт!", err)
	case "2":
		err = core.ImportBankAccountsFromXML(db)
		printBadOrGood("", "Успешный импорт!", err)
	case "3":
		err = core.ImportAtmsFromXML(db)
		printBadOrGood("", "Успешный импорт!", err)
	case "q":
		return true
	}
	return false
}
func importJsonLoop(db *sql.DB, cmd string) bool {
	var err error
	switch cmd {
	case "1":
		err = core.ImportClientsFromJSON(db)
		printBadOrGood("", "Успешный импорт!", err)
	case "2":
		err = core.ImportBankAccountsFromJSON(db)
		printBadOrGood("", "Успешный импорт!", err)
	case "3":
		err = core.ImportAtmsFromJSON(db)
		printBadOrGood("", "Успешный импорт!", err)
	case "q":
		return true
	}
	return false
}

func exportLoop(db *sql.DB, cmd string) (exit bool) {
	switch cmd {
	case "1":
		operationsLoop(db, exportImportData, exportJsonLoop)
	case "2":
		operationsLoop(db, exportImportData, exportXmlLoop)
	case "q":
		return true
	}
	return false
}
func exportJsonLoop(db *sql.DB, cmd string) (exit bool) {
	var err error
	switch cmd {
	case "1":
		err = core.ExportClientsToJSON(db)
		printBadOrGood("", "Успешный экспорт!", err)
	case "2":
		err = core.ExportBankAccountsToJSON(db)
		printBadOrGood("", "Успешный экспорт!", err)
	case "3":
		err = core.ExportAtmsToJSON(db)
		printBadOrGood("", "Успешный экспорт!", err)
	case "q":
		return true
	}
	return false
}
func exportXmlLoop(db *sql.DB, cmd string) (exit bool) {
	var err error
	switch cmd {
	case "1":
		err = core.ExportClientsToXML(db)
		printBadOrGood("", "Успешный экспорт!", err)
	case "2":
		err = core.ExportBankAccountsToXML(db)
		printBadOrGood("", "Успешный экспорт!", err)
	case "3":
		err = core.ExportAtmsToXML(db)
		printBadOrGood("", "Успешный экспорт!", err)
	case "q":
		return true
	}
	return false
}

func handleAddATM(db *sql.DB) error {
	var atmAddress string
	fmt.Fprintln(outstream, "Введите адрес банкомата")
	fmt.Fscan(instream, &atmAddress)
	err := core.AddATM(atmAddress, db)
	return err
}
func handleAddService(db *sql.DB) (string, error) {
	service := core.Service{}
	fmt.Fprintln(outstream, "Введите название сервиса")
	fmt.Fscan(instream, &service.Name)
	serviceNumber, err := core.AddService(service, db)
	return serviceNumber, err
}
func handleAddBankAccountToClient(db *sql.DB) error {
	var phone string
	fmt.Fprintln(outstream, "Введите номер телефона")
	fmt.Fscan(instream, &phone)
	clientID, err := core.GetClientIdByPhoneNumber(phone, db)
	if err != nil {
		return err
	}
	err = core.AddBankAccountToClient(clientID, db)
	return err
}
func handleAddClient(db *sql.DB) (err error) {
	client := core.Client{}
	fmt.Fprintln(outstream, "Введите имя")
	fmt.Fscan(instream, &client.Name)
	fmt.Fprintln(outstream, "Придумайте и введите логин")
	fmt.Fscan(instream, &client.Login)
	fmt.Fprintln(outstream, "Придумайте и введите пароль")
	fmt.Fscan(instream, &client.Password)
	fmt.Fprintln(outstream, "Введите номер телефона")
	fmt.Fscan(instream, &client.Phone)
	err = core.AddClient(client, db)
	return err
}
func handleLogin(db *sql.DB) (bool, error) {
	var login string
	var password string
	fmt.Println("Введите логин")
	fmt.Scan(&login)
	fmt.Println("Введите пароль")
	fmt.Scan(&password)

	ok, err := core.LoginForManager(login, password, db)
	if err != nil {
		return false, err
	}
	return ok, nil
}
func handleAddManager(db *sql.DB) error {
	manager := core.Manager{}
	fmt.Fprintln(outstream, "Придумайте и введите логин")
	fmt.Fscan(instream, &manager.Login)
	fmt.Fprintln(outstream, "Придумайте и введите пароль")
	fmt.Fscan(instream, &manager.Password)
	err := core.AddManager(manager, db)
	return err
}

func printBadOrGood(bad, good string, err error) {
	if err != nil {
		log.Println(err)
		fmt.Fprintln(outstream, "Упс... что-то пошло не так")
	} else {
		fmt.Fprintln(outstream, good)
	}
}