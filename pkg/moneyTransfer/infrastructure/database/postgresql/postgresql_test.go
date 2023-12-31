package postgresql_test

import (
	"log"
	"testing"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/infrastructure/database/postgresql"
	"github.com/brianvoe/gofakeit"
	"github.com/google/uuid"
)

func newTestPostgreSQL() *postgresql.PostgreSQL {
	db, err := postgresql.ConnectToDatabase()
	if err != nil {
		log.Panicf("error connecting to the testing database: %v", err)
	}

	return postgresql.NewPostgreSQLDatabase((db))
}

func TestPostgreSQL_CreateAccount(t *testing.T) {
	p := newTestPostgreSQL()

	type args struct {
		account *domain.Account
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				account: &domain.Account{
					Name:        gofakeit.Name(),
					Description: "Customer's deposit account",
					BalanceType: domain.Credit,
				},
			},
			wantErr: false,
		},
		{
			name:    "sad case - no account information",
			args:    args{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := p.CreateAccount(tt.args.account)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgreSQL.CreateAccount() error = %v, wantErr %v", err, tt.wantErr)
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
			}
			if tt.wantErr && account != nil {
				t.Errorf("did not expect an account to be created")
				return
			}
		})
	}
}

func TestPostgreSQL_Account(t *testing.T) {
	p := newTestPostgreSQL()

	newAccount, err := p.CreateAccount(&domain.Account{
		Name:        gofakeit.Name(),
		Description: "Customer's deposit account",
		BalanceType: domain.Credit,
		Currency:    "UGX",
	})
	if err != nil {
		t.Errorf("unable to create test account: %v", err)
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
				accountID: newAccount.UUID,
			},
			wantErr: false,
		},
		{
			name: "sad case - nonexistent account",
			args: args{
				accountID: uuid.New().String(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			account, err := p.Account(tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("PostgreSQL.Account() error = %v, wantErr %v", err, tt.wantErr)
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
				if account.Currency != domain.Ugandan {
					t.Errorf("expected a default KSH account")
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
