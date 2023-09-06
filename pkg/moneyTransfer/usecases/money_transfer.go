package usecases

import (
	"fmt"
	"log"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/application"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain/data"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/repository"
)

// MoneyTransferUsecases defines a contract the money transfer usecase adheres to
type MoneyTransferUsecases interface {
	CreateCustomerAccount(accountInput application.AccountCreationInput) (*application.AccountInformationOutput, error)
	Account(accountID string) (*application.AccountInformationOutput, error)
	Transfer(transferInput application.TransferInput) (*domain.Transaction, error)
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
func (mt MoneyTransfer) CreateCustomerAccount(accountInput application.AccountCreationInput) (*application.AccountInformationOutput, error) {
	depositAmount := accountInput.Amount
	if depositAmount == nil {
		return nil, fmt.Errorf("a deposit amount should be provided for a new account")
	}

	accountInfo := domain.Account{
		Name:        fmt.Sprintf("%s %s account", accountInput.CustomerName, accountInput.Header),
		Description: fmt.Sprintf("%s %s account", accountInput.CustomerName, accountInput.Header),
		Header:      accountInput.Header,
		Currency:    *accountInput.Currency,
	}

	switch accountInput.Header {
	case domain.Deposit:
		accountInfo.BalanceType = domain.Credit

	case domain.Loan:
		accountInfo.BalanceType = domain.Debit
	}

	account, err := mt.Create.CreateAccount(&accountInfo)
	if err != nil {
		return nil, err
	}

	systemAccount, err := mt.Account(data.SYSTEM_CASH_ACCOUNT)
	if err != nil {
		return nil, err
	}

	transferInput := application.TransferInput{
		SourceAccount:      systemAccount,
		DestinationAccount: account,
		Amount:             accountInput.Amount,
	}

	if _, err = mt.Transfer(transferInput); err != nil {
		return nil, err
	}

	return mt.Account(account.UUID)
}

// Account retrieves an account given it identifier
func (mt MoneyTransfer) Account(accountID string) (*application.AccountInformationOutput, error) {
	return mt.Get.Account(accountID)
}

// Transfer handles the movement of money from a source to a destination account
func (mt MoneyTransfer) Transfer(transferInput application.TransferInput) (*domain.Transaction, error) {
	sourceAccount := transferInput.SourceAccount
	destinationAccount := transferInput.DestinationAccount

	if sourceAccount == nil {
		return nil, fmt.Errorf("source account is required")
	}

	if destinationAccount == nil {
		return nil, fmt.Errorf("destination account is required")
	}

	amount := transferInput.Amount
	sourceAccountBalance := sourceAccount.Balance

	if !sourceAccount.IsSystemAccount && amount.GreaterThan(*sourceAccountBalance) {
		return nil, fmt.Errorf("%v is more than %s current account's balance of %v",
			amount,
			sourceAccount.Name,
			sourceAccountBalance,
		)
	}

	var description string
	var crEntry domain.AccountEntry
	var drEntry domain.AccountEntry
	switch sourceAccount.Header {
	case domain.Deposit, domain.Cash:
		description = fmt.Sprintf("Deposit of %v from account %s to account %s",
			amount,
			sourceAccount.Number,
			destinationAccount.Number,
		)
		crEntry = domain.AccountEntry{
			CreditAmount: *amount,
			AccountID:    sourceAccount.UUID,
		}

		drEntry = domain.AccountEntry{
			DebitAmount: *amount,
			AccountID:   destinationAccount.UUID,
		}
	}

	return mt.Create.CreateTransaction(description, &drEntry, &crEntry)
}
