package usecases_test

import (
	"log"
	"testing"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/application"
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
	currency := domain.Kenyan
	accountInput := application.AccountCreationInput{
		CustomerName: "John Doe",
		Amount:       &amount,
		Currency:     &currency,
		Header:       domain.Deposit,
	}

	accountInputWithoutAmount := application.AccountCreationInput{
		CustomerName: "John Doe",
		Currency:     &currency,
		Header:       domain.Deposit,
	}

	negativeAmount := decimal.NewFromInt(-100)
	accountInputWithoutNegativeAmount := application.AccountCreationInput{
		CustomerName: "John Doe",
		Currency:     &currency,
		Header:       domain.Deposit,
		Amount:       &negativeAmount,
	}

	zeroAmount := decimal.Zero
	accountInputWithoutZeroAmount := application.AccountCreationInput{
		CustomerName: "John Doe",
		Currency:     &currency,
		Header:       domain.Deposit,
		Amount:       &zeroAmount,
	}
	type args struct {
		accountInput application.AccountCreationInput
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
				accountInput: accountInput,
			},
			wantErr: false,
		},
		{
			name: "sad case - no deposit",
			args: args{
				accountInput: accountInputWithoutAmount,
			},
			wantErr: true,
		},
		{
			name:    "sad case - no account information",
			args:    args{},
			wantErr: true,
		},
		{
			name: "sad case - negative deposit",
			args: args{
				accountInput: accountInputWithoutNegativeAmount,
			},
			wantErr: true,
		},
		{
			name: "sad case - deposit of 0",
			args: args{
				accountInput: accountInputWithoutZeroAmount,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mt := newTestMoneyTransferUsecases()
			// Create system accounts
			if err := mt.Create.CreateSystemAccount(); err != nil {
				log.Panicf("system error, unable to create default account(s): %v", err)
			}

			account, err := mt.CreateCustomerAccount(tt.args.accountInput)
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

				account, err := mt.Get.Account(account.UUID)
				if err != nil {
					t.Errorf(err.Error())
					return
				}
				if !account.Balance.Equal(amount) {
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

	amount := decimal.NewFromInt(100)
	currency := domain.Kenyan
	accountInput := application.AccountCreationInput{
		CustomerName: "John Doe",
		Amount:       &amount,
		Currency:     &currency,
		Header:       domain.Deposit,
	}

	srcAccount, err := mt.CreateCustomerAccount(accountInput)
	if err != nil {
		t.Errorf("unable to create test src account: %v", err)
	}

	destAccount, err := mt.CreateCustomerAccount(accountInput)
	if err != nil {
		t.Errorf("unable to create test dest account: %v", err)
	}

	transferAmount := decimal.NewFromInt(80)
	transferInput := application.TransferInput{
		SourceAccount:      srcAccount,
		DestinationAccount: destAccount,
		Amount:             &transferAmount,
	}

	noSrcTransferInput := application.TransferInput{
		DestinationAccount: destAccount,
		Amount:             &transferAmount,
	}

	noDestTransferInput := application.TransferInput{
		SourceAccount: srcAccount,
		Amount:        &transferAmount,
	}

	largeAmount := decimal.NewFromInt(800)
	invalidAmountInput := application.TransferInput{
		SourceAccount:      srcAccount,
		DestinationAccount: destAccount,
		Amount:             &largeAmount,
	}

	type args struct {
		transferInput application.TransferInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				transferInput: transferInput,
			},
			wantErr: false,
		},
		{
			name: "sad case - non-existent account(dest)",
			args: args{
				transferInput: noDestTransferInput,
			},
			wantErr: true,
		},
		{
			name: "sad case - non-existent account(src)",
			args: args{
				transferInput: noSrcTransferInput,
			},
			wantErr: true,
		},
		{
			name: "sad case - transfer more than current balance",
			args: args{
				transferInput: invalidAmountInput,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transaction, err := mt.Transfer(tt.args.transferInput)
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

	amount := decimal.NewFromInt(100)
	currency := domain.Kenyan
	accountInput := application.AccountCreationInput{
		CustomerName: "John Doe",
		Amount:       &amount,
		Currency:     &currency,
		Header:       domain.Deposit,
	}

	srcAccount, err := mt.CreateCustomerAccount(accountInput)
	if err != nil {
		t.Errorf("unable to create test src account: %v", err)
		return
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
