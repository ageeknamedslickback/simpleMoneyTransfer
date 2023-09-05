package postgresql

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgreSQL sets up the PostgreSQL databas layer with all the necessary dependencies
type PostgreSQL struct {
	ORM *gorm.DB
}

// CheckPreconditions ensures PostgreSQL's contract is adhered to
func (p PostgreSQL) CheckPreconditions() {
	if p.ORM == nil {
		log.Panicf("PostgreSQL's ORM driver has not been initialized")
	}
}

// NewPostgreSQLDatabase initializes a new PostgreSQL database instance
func NewPostgreSQLDatabase(gorm *gorm.DB) *PostgreSQL {
	db := &PostgreSQL{ORM: gorm}
	db.CheckPreconditions()
	return db
}

// ConnectToDatabase opens a connection to a given database
func ConnectToDatabase() (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s user=%s password=%s port=%s dbname=%s sslmode=disable TimeZone=Africa/Nairobi",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("server is unable to connect to the database: %v", err)
	}

	if err := Migrate(db); err != nil {
		return nil, err
	}

	return db, nil
}

func Migrate(db *gorm.DB) error {
	tables := []interface{}{
		&domain.Account{},
		&domain.Transaction{},
		&domain.AccountEntry{},
	}
	for _, table := range tables {
		if err := db.AutoMigrate(table); err != nil {
			return fmt.Errorf("server is unable to run database migrations: %v", err)
		}
	}

	return nil
}

// CreateAccount does a database call to create a account
func (p PostgreSQL) CreateAccount(
	account *domain.Account,
	depositAmount *decimal.Decimal,
) (*domain.Account, error) {
	if account == nil {
		return nil, fmt.Errorf("missing account creation information")
	}

	if depositAmount == nil {
		return nil, fmt.Errorf("a deposit amount should be provided for a new account")
	}

	if err := p.ORM.Create(&account).Error; err != nil {
		return nil, fmt.Errorf("unable to create account: %v", err)
	}

	effectiveDate := time.Now()
	drEntry := domain.AccountEntry{
		DebitAmount:   *depositAmount,
		EffectiveDate: &effectiveDate,
		AccountID:     "4a1c2699-3716-489e-903b-af0ebc4952bf", // get a system account to do the transaction
	}
	crEntry := domain.AccountEntry{
		CreditAmount:  *depositAmount,
		EffectiveDate: &effectiveDate,
		AccountID:     account.UUID,
	}

	description := "Account activation deposit"
	if _, err := p.CreateTransaction(description, &drEntry, &crEntry); err != nil {
		return nil, fmt.Errorf("unable to make a transaction: %v", err)
	}

	return account, nil
}

// CreateTransaction does a database call to create a transaction with account entries
func (p PostgreSQL) CreateTransaction(
	description string,
	drEntry *domain.AccountEntry,
	crEntry *domain.AccountEntry,
) (*domain.Transaction, error) {
	if drEntry == nil {
		return nil, fmt.Errorf("DR entry should be provided for a transaction")
	}

	if crEntry == nil {
		return nil, fmt.Errorf("CR entry should be provided for a transaction")
	}

	if err := drEntry.ValidateDebitAmount(); err != nil {
		return nil, err
	}

	if err := crEntry.ValidateCreditAmount(); err != nil {
		return nil, err
	}

	if drEntry.DebitAmount != crEntry.CreditAmount {
		return nil, fmt.Errorf("transaction does not observe double entry")
	}

	var transaction domain.Transaction
	if err := p.ORM.Transaction(func(tx *gorm.DB) error {
		transaction = domain.Transaction{Description: description}
		if err := tx.Create(&transaction).Error; err != nil {
			return fmt.Errorf("unable to create an accounting transaction: %v", err)
		}

		var entries []*domain.AccountEntry
		entries = append(entries, drEntry, crEntry)
		for _, entry := range entries {
			entry.TransactionID = transaction.UUID
			if err := tx.Create(&entry).Error; err != nil {
				return fmt.Errorf("unable to create an account entry: %v", err)
			}
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("unable to commit transaction: %v", err)
	}

	return &transaction, nil
}

// Account retrieves an account given it's ID(UUID)
func (p PostgreSQL) Account(accountID string) (*domain.Account, error) {
	var account domain.Account

	filter := domain.Account{
		AbstractBase: domain.AbstractBase{
			UUID: accountID,
		},
	}
	if err := p.ORM.Where(&filter).First(&account).Error; err != nil {
		return nil, fmt.Errorf("unable to get account %s: %v", accountID, err)
	}

	return &account, nil
}

// AccountDebitTotal aggregates all the debits done to an account
func (p PostgreSQL) AccountDebitTotal(account *domain.Account) (*decimal.Decimal, error) {
	accountID := account.UUID
	if _, err := uuid.Parse(accountID); err != nil {
		return nil, fmt.Errorf("%s is not a valid uuid: %v", accountID, err)
	}

	var total decimal.Decimal
	if err := p.ORM.Raw("SELECT SUM(debit_amount::float) FROM account_entries WHERE account_id = ?", accountID).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("unable to get the account's total debits: %v", err)
	}

	return &total, nil
}

// AccountCreditTotal aggregates all the credits done to an account
func (p PostgreSQL) AccountCreditTotal(account *domain.Account) (*decimal.Decimal, error) {
	accountID := account.UUID
	if _, err := uuid.Parse(accountID); err != nil {
		return nil, fmt.Errorf("%s is not a valid uuid: %v", accountID, err)
	}

	var total decimal.Decimal
	if err := p.ORM.Raw("SELECT SUM(credit_amount::float) FROM account_entries WHERE account_id = ?", accountID).Scan(&total).Error; err != nil {
		return nil, fmt.Errorf("unable to get the account's total credits: %v", err)
	}

	return &total, nil
}

// AccountBalance computes the balance of an account from it's entries
func (p PostgreSQL) AccountBalance(account *domain.Account) (*decimal.Decimal, error) {
	if account == nil {
		return nil, fmt.Errorf("account has not been supplied")
	}

	debits, err := p.AccountDebitTotal(account)
	if err != nil {
		return nil, err
	}

	credits, err := p.AccountCreditTotal(account)
	if err != nil {
		return nil, err
	}

	var balance decimal.Decimal
	if account.BalanceType == domain.Debit {
		balance = debits.Sub(*credits)
	} else {
		balance = credits.Sub(*debits)
	}

	return &balance, nil
}
