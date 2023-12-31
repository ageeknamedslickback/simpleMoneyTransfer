package postgresql

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/application"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain/data"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/GoogleCloudPlatform/cloudsql-proxy/proxy/dialers/postgres"
)

var DUPLICATE_KEY_MSG = "duplicate key value violates unique constraint"

// PostgreSQL sets up the PostgreSQL database layer with all the necessary dependencies
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
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Africa/Nairobi",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASS"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_PORT"),
	)

	if os.Getenv("DB_CONNECTION") != "cloud" {
		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("can't open connection to the local database: %v", err)
		}
		if err := Migrate(db); err != nil {
			return nil, err
		}

		return db, nil
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		DriverName: "cloudsqlpostgres",
		DSN:        dsn,
	}))

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

// CreateSystemAccount created default system accounts
func (p PostgreSQL) CreateSystemAccount() error {
	for _, account := range data.SystemAccounts() {
		err := p.ORM.Create(&account).Error
		if err != nil {
			if strings.Contains(err.Error(), DUPLICATE_KEY_MSG) {
				continue
			} else {
				return err
			}
		}
	}
	return nil
}

// CreateAccount does a database call to create a account
func (p PostgreSQL) CreateAccount(account *domain.Account) (*application.AccountInformationOutput, error) {
	if account == nil {
		return nil, fmt.Errorf("missing account creation information")
	}

	if err := p.ORM.Create(&account).Error; err != nil {
		return nil, fmt.Errorf("unable to create account: %v", err)
	}

	return p.Account(account.UUID)
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
func (p PostgreSQL) Account(accountID string) (*application.AccountInformationOutput, error) {
	var account domain.Account

	filter := domain.Account{
		AbstractBase: domain.AbstractBase{
			UUID: accountID,
		},
	}
	if err := p.ORM.Where(&filter).First(&account).Error; err != nil {
		return nil, fmt.Errorf("unable to get account %s: %v", accountID, err)
	}

	balance, err := p.AccountBalance(&account)
	if err != nil {
		return nil, fmt.Errorf("unable to get account's balance: %v", err)
	}

	effectiveDate := time.Now()
	accountOutput := application.AccountInformationOutput{
		UUID:            account.UUID,
		Active:          account.Active,
		CreatedAt:       account.CreatedAt,
		UpdatedAt:       account.UpdatedAt,
		Name:            account.Name,
		Description:     account.Description,
		Currency:        account.Currency,
		BalanceType:     account.BalanceType,
		Header:          account.Header,
		IsSystemAccount: account.IsSystemAccount,
		Number:          account.Number,
		Balance:         balance,
		BalanceAsOf:     &effectiveDate,
	}

	return &accountOutput, nil
}

// AccountDebitTotal aggregates all the debits done to an account
func (p PostgreSQL) AccountDebitTotal(account *domain.Account) (*decimal.Decimal, error) {
	accountID := account.UUID
	if _, err := uuid.Parse(accountID); err != nil {
		return nil, fmt.Errorf("%s is not a valid uuid: %v", accountID, err)
	}

	var total decimal.Decimal
	if err := p.ORM.Raw("SELECT COALESCE(SUM(debit_amount::float), 0) AS totalDebit FROM account_entries WHERE account_id = ?", accountID).Scan(&total).Error; err != nil {
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
	if err := p.ORM.Raw("SELECT COALESCE(SUM(credit_amount::float), 0) AS totalCredit FROM account_entries WHERE account_id = ?", accountID).Scan(&total).Error; err != nil {
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
	if account.BalanceType == domain.Credit {
		balance = debits.Sub(*credits)
	} else {
		balance = credits.Sub(*debits)
	}

	return &balance, nil
}
