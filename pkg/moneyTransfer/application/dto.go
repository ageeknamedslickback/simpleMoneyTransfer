package application

import (
	"time"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain"
	"github.com/shopspring/decimal"
)

// AccountCreationInput represents input object for account creation
type AccountCreationInput struct {
	CustomerName string
	Amount       *decimal.Decimal
	Currency     *domain.CurrencyType
	Header       domain.HeaderType
}

// TransferPayload defines the presentation layer transfer payload
type TransferPayload struct {
	SourceAccountID      string
	DestinationAccountID string
	Amount               *decimal.Decimal
}

// TransferInput represents input object for a transfer transaction
type TransferInput struct {
	SourceAccount      *AccountInformationOutput
	DestinationAccount *AccountInformationOutput
	Amount             *decimal.Decimal
}

// AccountInformationOutput represents a robust output object for accounts
type AccountInformationOutput struct {
	UUID            string
	Active          bool
	CreatedAt       *time.Time
	UpdatedAt       *time.Time
	Name            string
	Description     string
	Number          string
	Currency        domain.CurrencyType
	BalanceType     domain.BalanceType
	Header          domain.HeaderType
	IsSystemAccount bool
	Balance         *decimal.Decimal
	BalanceAsOf     *time.Time
}

// AccessToken represents Auth0 oauth2 access token
type AccessToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}
