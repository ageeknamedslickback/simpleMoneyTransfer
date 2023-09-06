package data

import "github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/domain"

var SYSTEM_CASH_ACCOUNT = "ddff1ec2-edb2-4d8e-90f0-115766cace6b"

// SystemAccounts creates system related control accounts
func SystemAccounts() []*domain.Account {
	return []*domain.Account{
		{
			AbstractBase: domain.AbstractBase{
				UUID: SYSTEM_CASH_ACCOUNT,
			},
			Name:            "Default System's Payment Method account",
			Description:     "Default System's Payment Method account",
			Number:          "AC-0123456789",
			Currency:        domain.Kenyan,
			BalanceType:     domain.Debit,
			Header:          domain.Cash,
			IsSystemAccount: true,
		},
	}
}
