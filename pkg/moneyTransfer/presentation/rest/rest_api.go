package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/application"
	"github.com/ageeknamedslickback/simpleMoneyTransfer/pkg/moneyTransfer/usecases"
	"github.com/gin-gonic/gin"
)

// RestHandlers defines a contract the money transfer rest presentation adheres to
type RestHandlers interface {
	Authenticate(c *gin.Context)
	CreateAccount(c *gin.Context)
	Account(c *gin.Context)
	Transfer(c *gin.Context)
}

// Rest sets up REST presentation layer with all it's dependencies
type Rest struct {
	Uc usecases.MoneyTransferUsecases
}

// CheckPreconditions ensures a correct Rest struct is initialized
func (r Rest) CheckPreconditions() {
	if r.Uc == nil {
		log.Panic("rest presentation layer has not initialized the business logic")
	}
}

// NewRestHandlers initializes a new Rest API endpoints handler
func NewRestHandlers(uc usecases.MoneyTransferUsecases) *Rest {
	rst := &Rest{
		Uc: uc,
	}
	rst.CheckPreconditions()
	return rst
}

func jsonErrorResponse(c *gin.Context, statusCode int, err string) {
	c.JSON(statusCode, gin.H{"error": err})
}

// CreateAccount is account creation handler
func (r Rest) CreateAccount(c *gin.Context) {
	var accountCreationInput application.AccountCreationInput
	if err := c.ShouldBindJSON(&accountCreationInput); err != nil {
		jsonErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	account, err := r.Uc.CreateCustomerAccount(accountCreationInput)
	if err != nil {
		jsonErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"account": account})
}

// Account implements a get account endpoint handler
func (r Rest) Account(c *gin.Context) {
	accountID := c.Param("id")
	if accountID == "" {
		err := "account ID has not bee provided"
		jsonErrorResponse(c, http.StatusBadRequest, err)
		return
	}

	account, err := r.Uc.Account(accountID)
	if err != nil {
		jsonErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"account": account})
}

// Transfer implements an account transaction handler
func (r Rest) Transfer(c *gin.Context) {
	var payload application.TransferPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		jsonErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	sourceAccount, err := r.Uc.Account(payload.SourceAccountID)
	if err != nil {
		jsonErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	destinationAccount, err := r.Uc.Account(payload.DestinationAccountID)
	if err != nil {
		jsonErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	transferInput := application.TransferInput{
		SourceAccount:      sourceAccount,
		DestinationAccount: destinationAccount,
		Amount:             payload.Amount,
	}
	transaction, err := r.Uc.Transfer(transferInput)
	if err != nil {
		jsonErrorResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"transaction": transaction})
}

// Authenticate provides an authentication endpoint that returns an access token
// to interact with the other APIs
func (r Rest) Authenticate(c *gin.Context) {
	params := url.Values{}
	params.Add("grant_type", os.Getenv("AUTH0_GRANT_TYPE"))
	params.Add("client_id", os.Getenv("AUTH0_CLIENT_ID"))
	params.Add("client_secret", os.Getenv("AUTH0_CLIENT_SECRET"))
	params.Add("audience", os.Getenv("AUTH0_AUDIENCE"))
	payload := strings.NewReader(params.Encode())

	URL := fmt.Sprintf("https://%s/oauth/token", os.Getenv("AUTH0_DOMAIN"))
	req, err := http.NewRequest(http.MethodPost, URL, payload)
	if err != nil {
		jsonErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	req.Header.Set("content-type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		jsonErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		jsonErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var accessToken application.AccessToken
	if err := json.Unmarshal(body, &accessToken); err != nil {
		jsonErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{"response": accessToken})
}
