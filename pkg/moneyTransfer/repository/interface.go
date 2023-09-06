package repository

import (
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/application"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain"
	"github.com/shopspring/decimal"
)

// CreateRepository abstracts the Create contract that any repository should adhere to
type CreateRepository interface {
	CreateAccount(account *domain.Account) (*application.AccountInformationOutput, error)
	CreateTransaction(
		description string,
		drEntry *domain.AccountEntry,
		crEntry *domain.AccountEntry,
	) (*domain.Transaction, error)
}

// GetRepository abstracts the Get contract that any repository should adhere to
type GetRepository interface {
	Account(accountID string) (*application.AccountInformationOutput, error)
	AccountDebitTotal(account *domain.Account) (*decimal.Decimal, error)
	AccountCreditTotal(account *domain.Account) (*decimal.Decimal, error)
	AccountBalance(account *domain.Account) (*decimal.Decimal, error)
}
