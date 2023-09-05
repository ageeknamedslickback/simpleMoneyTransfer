package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

// AbstractBase is an abstract struct that can be embedded in other structs
type AbstractBase struct {
	UUID      string `gorm:"primaryKey"`
	Active    bool   `gorm:"default:true"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// BeforeCreate ensures a UUID and createdAt data is inserted
func (ab *AbstractBase) BeforeCreate(tx *gorm.DB) (err error) {
	ab.UUID = uuid.New().String()
	return
}

// CurrencyType represents the different currencies used in account transactions
type CurrencyType string

const (
	Kenyan  CurrencyType = "KSH"
	Ugandan CurrencyType = "UGX"
)

// BalanceType defines how an account's end balance is computed
type BalanceType string

const (
	Debit  BalanceType = "DR"
	Credit BalanceType = "CR"
)

// Account denotes a virtual storage and tracker for value (money/loyalty points)
type Account struct {
	AbstractBase    `gorm:"embedded"`
	Name            string
	Description     string
	Number          string
	Currency        CurrencyType `gorm:"default: KSH"`
	BalanceType     BalanceType
	IsSystemAccount bool `gorm:"default: false"`
}

// BeforeCreate ensures an account number is generated
func (acc *Account) BeforeCreate(tx *gorm.DB) (err error) {
	acc.UUID = uuid.New().String()
	acc.Number = fmt.Sprintf("AC-%v", time.Now().Unix())
	return
}

// Transaction maintains the movement/transfer of money from one account to another
type Transaction struct {
	AbstractBase `gorm:"embedded"`
	Description  string `json:"description"`
}

// AccountEntry hold information about the value, accounts involved in the transfer of money
type AccountEntry struct {
	AbstractBase  `gorm:"embedded"`
	DebitAmount   decimal.Decimal `json:"dr_amount,omitempty"`
	CreditAmount  decimal.Decimal `json:"cr_amount,omitempty"`
	EffectiveDate *time.Time      `json:"effective_date"`
	AccountID     string          `json:"account_id"`
	Account       Account         `json:"account,omitempty" gorm:"foreginKey:AccountID"`
	TransactionID string          `json:"transaction_id"`
	Transaction   Transaction     `json:"transaction,omitempty" gorm:"foreginKey:TransactionID"`
}

// ValidateDebitAmount validates that a debit amount is non-negative
func (ae AccountEntry) ValidateDebitAmount() error {
	if ae.DebitAmount.IsNegative() {
		return fmt.Errorf("you can not debit a negative amount")
	}

	return nil
}

// ValidateDebitAmount validates that a credit amount is non-negative
func (ae AccountEntry) ValidateCreditAmount() error {
	if ae.CreditAmount.IsNegative() {
		return fmt.Errorf("you can not credit a negative amount")
	}

	return nil
}
