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
	case "2":
		atms, err := core.AtmsList(db)
		printBadOrGood("", "Список банкоматов\n", err)
		if err == nil {
			for _, atm := range atms {
				fmt.Fprintln(outstream, atm)
			}
		}
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
		err = handleBankAccountsList(db)
		printBadOrGood("", "", err)
	case "2":
		err := handleTransferToClient(db)
		printBadOrGood("", "Перевод успешно совершён!", err)
	case "3":
		err := handleTransferToClientByPhone(db)
		printBadOrGood("", "Перевод успешно совершён!", err)
	case "4":
		err := handlePayForService(db)
		printBadOrGood("", "Оплата успешно совершена!", err)
	case "5":
		atms, err := core.AtmsList(db)
		printBadOrGood("", "Список банкоматов\n", err)
		if err == nil {
			for _, atm := range atms {
				fmt.Fprintln(outstream, atm)
			}
		}
	case "q":
		return true
	default:
		fmt.Fprintf(outstream, "Вы выбрали неверную команду: %s\n", cmd)
	}

	return false
}

func handleTransferToClientByPhone(db *sql.DB) error {
	clientId, err := core.GetClientIdByLogin(login, db)
	if err != nil {
		return nil
	}
	var phone string
	fmt.Fprintln(outstream, "Введите номер телефона клиента")
	fmt.Fscan(instream, &phone)

	receiverId, err := core.GetClientIdByPhoneNumber(phone, db)
	if err != nil {
		return err
	}

	var receiverAccountNumber int64
	fmt.Fprintln(outstream, "Выберите номер счёта получателя")
	bankAccounts, err := core.GetAllAccountNumbersByClientId(receiverId, db)
	for _, account := range bankAccounts {
		fmt.Fprintln(outstream, account)
	}
	fmt.Fscan(instream, &receiverAccountNumber)
	if err != nil {
		return err
	}
	transfer := core.MoneyTransfer{}
	transfer.SenderId = clientId

	fmt.Fprintln(outstream, "Выберите номер счёта для перевода (последние четыре цифры)")
	err = handleBankAccountsList(db)
	if err != nil {
		return err
	}
	fmt.Fscan(instream, &transfer.SenderAccountNumber)
	transfer.ReceiverId = receiverId
	transfer.ReceiverAccountNumber = receiverAccountNumber
	fmt.Fprintln(outstream, "Введите сумму перевода")
	fmt.Fscan(instream, &transfer.Amount)
	err = core.TransferToClient(transfer, db)
	return err
}

func handlePayForService(db *sql.DB) error {
	clientId, err := core.GetClientIdByLogin(login, db)
	if err != nil {
		return nil
	}
	var serviceNumber string
	fmt.Fprintln(outstream, "Введите номер услуги")
	fmt.Fscan(instream, &serviceNumber)

	var accountNumber int64
	fmt.Fprintln(outstream, "Выберите номер счёта для перевода (последние четыре цифры)")
	err = handleBankAccountsList(db)
	if err != nil {
		return err
	}
	fmt.Fscan(instream, &accountNumber)

	var amount int64
	fmt.Fprintln(outstream, "Введите сумму перевода")
	fmt.Fscan(instream, &amount)

	err = core.PayForService(serviceNumber, amount, clientId, accountNumber, db)
	return err
}

func handleTransferToClient(db *sql.DB) error {
	clientId, err := core.GetClientIdByLogin(login, db)
	if err != nil {
		return nil
	}
	var clientNumber string
	fmt.Fprintln(outstream, "Введите номер счета клиента")
	fmt.Fscan(instream, &clientNumber)

	receiverId, receiverAccountNumber, err := core.ServiceNumberToIdAndAccountNumber(clientNumber)
	if err != nil {
		return err
	}
	transfer := core.MoneyTransfer{}
	transfer.SenderId = clientId

	fmt.Fprintln(outstream, "Выберите номер счёта для перевода (последние четыре цифры)")
	err = handleBankAccountsList(db)
	if err != nil {
		return err
	}
	fmt.Fscan(instream, &transfer.SenderAccountNumber)
	transfer.ReceiverId = receiverId
	transfer.ReceiverAccountNumber = receiverAccountNumber
	fmt.Fprintln(outstream, "Введите сумму перевода")
	fmt.Fscan(instream, &transfer.Amount)
	err = core.TransferToClient(transfer, db)
	return err
}

func handleBankAccountsList(db *sql.DB) error {
	clientId, err := core.GetClientIdByLogin(login, db)
	if err != nil {
		return err
	}
	bankAccounts, err := core.BankAccountsList(clientId, db)
	if err != nil {
		return err
	}
	for _, account := range bankAccounts {
		fmt.Fprintf(outstream, "account: %09d%04d balance: %v\n", account.UserId, account.AccountId, account.Balance)
	}
	return nil
}

var login string
func handleLogin(db *sql.DB) (bool, error) {
	var password string
	fmt.Println("Введите логин")
	fmt.Scan(&login)
	fmt.Println("Введите пароль")
	fmt.Scan(&password)

	ok, err := core.LoginForClient(login, password, db)
	if err != nil {
		return false, err
	}
	return ok, nil
}

func printBadOrGood(bad, good string, err error) {
	if err != nil {
		log.Println(err)
		fmt.Fprintln(outstream, "Упс... что-то пошло не так")
	} else {
		fmt.Fprintln(outstream, good)
	}
}