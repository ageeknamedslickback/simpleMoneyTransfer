package usecases_test

import (
	"log"
	"testing"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/infrastructure/database/postgresql"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/usecases"
	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func newTestMoneyTransferUsecases() *usecases.MoneyTransfer {
	db, err := postgresql.ConnectToDatabase()
	if err != nil {
		log.Panicf("error connecting to the testing database: %v", err)
	}
	create := postgresql.NewPostgreSQLDatabase(db)
	get := postgresql.NewPostgreSQLDatabase(db)

	return usecases.NewMoneyTransferUsecases(create, get)
}
func TestMoneyTransfer_CreateCustomerAccount(t *testing.T) {
	amount := decimal.NewFromInt(100)
	type args struct {
		customerName string
		startBalance *decimal.Decimal
		currency     *domain.CurrencyType
		header       domain.HeaderType
	}
	tests := []struct {
		name    string
		args    args
		want    *domain.Account
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				customerName: "John Doe",
				startBalance: &amount,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := newTestMoneyTransferUsecases()
			account, err := mt.CreateCustomerAccount(tt.args.customerName, tt.args.startBalance, tt.args.currency, tt.args.header)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoneyTransfer.CreateCustomerAccount() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if account == nil {
					t.Errorf("expected account to be created")
					return
				}
				if account.Number == "" {
					t.Errorf("expected account to have a number")
					return
				}
				if account.Currency != domain.Kenyan {
					t.Errorf("expected a default KSH account")
					return
				}

				balance, err := mt.Get.AccountBalance(account)
				if err != nil {
					t.Errorf(err.Error())
					return
				}
				if !balance.Equal(amount) {
					t.Errorf("expected the account to have a starting balance of 100")
					return
				}
			}
			if tt.wantErr && account != nil {
				t.Errorf("did not expect an account to be created")
				return
			}
		})
	}
}

func TestMoneyTransfer_Transfer(t *testing.T) {
	mt := newTestMoneyTransferUsecases()

	customerName := "John Doe"
	startBalance := decimal.NewFromInt(100)

	srcAccount, err := mt.CreateCustomerAccount(customerName, &startBalance, nil, domain.Deposit)
	if err != nil {
		t.Errorf("unable to create test src account: %v", err)
	}

	destAccount, err := mt.CreateCustomerAccount(customerName, &startBalance, nil, domain.Deposit)
	if err != nil {
		t.Errorf("unable to create test dest account: %v", err)
	}

	type args struct {
		sourceAccountID      string
		destinationAccountID string
		amount               decimal.Decimal
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				sourceAccountID:      srcAccount.UUID,
				destinationAccountID: destAccount.UUID,
				amount:               decimal.NewFromInt(80),
			},
			wantErr: false,
		},
		{
			name: "sad case - non-existent account(src)",
			args: args{
				sourceAccountID:      uuid.NewString(),
				destinationAccountID: destAccount.UUID,
				amount:               decimal.NewFromInt(80),
			},
			wantErr: true,
		},
		{
			name: "sad case - non-existent account(dest)",
			args: args{
				sourceAccountID:      srcAccount.UUID,
				destinationAccountID: uuid.NewString(),
				amount:               decimal.NewFromInt(80),
			},
			wantErr: true,
		},
		{
			name: "sad case - transfer more than current balance",
			args: args{
				sourceAccountID:      srcAccount.UUID,
				destinationAccountID: destAccount.UUID,
				amount:               decimal.NewFromInt(800),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := mt.Transfer(tt.args.sourceAccountID, tt.args.destinationAccountID, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoneyTransfer.Transfer() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && transaction != nil {
				t.Errorf("did not expect a transaction to happen")
				return
			}
			if !tt.wantErr && transaction == nil {
				t.Errorf("expected a transaction to happen")
				return
			}
		})
	}
}

func TestMoneyTransfer_Account(t *testing.T) {
	mt := newTestMoneyTransferUsecases()

	customerName := "John Doe"
	startBalance := decimal.NewFromInt(100)

	srcAccount, err := mt.CreateCustomerAccount(customerName, &startBalance, nil, domain.Deposit)
	if err != nil {
		t.Errorf("unable to create test src account: %v", err)
	}

	type args struct {
		accountID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				accountID: srcAccount.UUID,
			},
			wantErr: false,
		},
		{
			name: "sad case",
			args: args{
				accountID: uuid.NewString(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := mt.Account(tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("MoneyTransfer.Account() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && account != nil {
				t.Errorf("did not expect an account to happen")
				return
			}
			if !tt.wantErr && account == nil {
				t.Errorf("expected an account to happen")
				return
			}
		})
	}
}
