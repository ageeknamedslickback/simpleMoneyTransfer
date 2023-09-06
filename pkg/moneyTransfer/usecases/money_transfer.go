package usecases

import (
	"fmt"
	"log"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/repository"
	"github.com/shopspring/decimal"
)

// MoneyTransferUsecases defines a contract the money transfer usecase adheres to
type MoneyTransferUsecases interface {
	CreateCustomerAccount(
		customerName string,
		startBalance *decimal.Decimal,
		currency *domain.CurrencyType,
		header domain.HeaderType,
	) (*domain.Account, error)
	Account(accountID string) (*domain.Account, error)
	Transfer(
		sourceAccountID string,
		destinationAccountID string,
		amount decimal.Decimal,
	) (*domain.Transaction, error)
}

// MoneyTransfer set up the money transfer business logic and its dependencies
type MoneyTransfer struct {
	Create repository.CreateRepository
	Get    repository.GetRepository
}

// CheckPreconditions ensures all dependencies are injected
func (mt MoneyTransfer) CheckPreconditions() {
	if mt.Create == nil {
		log.Panic("money transfer usecase did not initialize the create repository")
	}

	if mt.Get == nil {
		log.Panic("money transfer usecase did not initialize the get repository")
	}
}

// NewMoneyTransferUsecases initializes a new money transfer business usecase
func NewMoneyTransferUsecases(
	createRepo repository.CreateRepository,
	getRepo repository.GetRepository,
) *MoneyTransfer {
	mt := &MoneyTransfer{
		Create: createRepo,
		Get:    getRepo,
	}
	mt.CheckPreconditions()
	return mt
}

// CreateCustomerAccount creates a new customer's account
func (mt MoneyTransfer) CreateCustomerAccount(
	customerName string,
	startBalance *decimal.Decimal,
	currency *domain.CurrencyType,
	header domain.HeaderType,
) (*domain.Account, error) {
	accountInfo := domain.Account{
		Name:        fmt.Sprintf("%s %s account", customerName, header),
		Description: fmt.Sprintf("%s %s account", customerName, header),
		Header:      header,
	}

	switch header {
	case domain.Deposit:
		accountInfo.BalanceType = domain.Credit

	case domain.Loan:
		accountInfo.BalanceType = domain.Debit
	}

	return mt.Create.CreateAccount(&accountInfo, startBalance)
}

// Account retrieves an account given it identifier
func (mt MoneyTransfer) Account(accountID string) (*domain.Account, error) {
	return mt.Get.Account(accountID)
}

// Transfer handles the movement of money from a source to a destination account
func (mt MoneyTransfer) Transfer(
	sourceAccountID string,
	destinationAccountID string,
	amount decimal.Decimal,
) (*domain.Transaction, error) {
	sourceAccount, err := mt.Get.Account(sourceAccountID)
	if err != nil {
		return nil, err
	}

	destinationAccount, err := mt.Get.Account(destinationAccountID)
	if err != nil {
		return nil, err
	}

	sourceAccountBalance, err := mt.Get.AccountBalance(sourceAccount)
	if err != nil {
		return nil, err
	}

	if amount.GreaterThan(*sourceAccountBalance) {
		return nil, fmt.Errorf("%v is more than your current account's balance of %v", amount, sourceAccountBalance)
	}

	var description string
	var crEntry domain.AccountEntry
	var drEntry domain.AccountEntry
	switch sourceAccount.Header {
	case domain.Deposit:
		description = fmt.Sprintf("Deposit of %v from account %s to account %s", amount, sourceAccount.Number, destinationAccount.Number)
		crEntry = domain.AccountEntry{
			CreditAmount: amount,
			AccountID:    sourceAccountID,
		}

		drEntry = domain.AccountEntry{
			DebitAmount: amount,
			AccountID:   destinationAccountID,
		}
	}

	return mt.Create.CreateTransaction(description, &drEntry, &crEntry)
}
