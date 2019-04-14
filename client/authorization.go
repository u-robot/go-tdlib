package client

import (
	"fmt"
	"time"
)

// AuthorizationStateHandler is interface declaring authorization state listener.
type AuthorizationStateHandler interface {
	Handle(client *Client, state AuthorizationState) error
	Close()
}

// Authorize starts authorization process for the client.
func Authorize(client *Client, authorizationStateHandler AuthorizationStateHandler) error {
	defer authorizationStateHandler.Close()

	for {
		state, err := client.GetAuthorizationState()
		if err != nil {
			return err
		}

		err = authorizationStateHandler.Handle(client, state)
		if err != nil {
			return err
		}

		if state.AuthorizationStateType() == TypeAuthorizationStateReady {
			// dirty hack for db flush after authorization
			time.Sleep(1 * time.Second)
			return nil
		}
	}
}

// Authorizer implements AuthorizationStateHandler interface.
type Authorizer struct {
	TdlibParameters chan *TdlibParameters
	PhoneNumber     chan string
	Code            chan string
	State           chan AuthorizationState
	FirstName       chan string
	LastName        chan string
	Password        chan string
}

// NewClientAuthorizer creates new instance of Authorizer.
func NewClientAuthorizer() *Authorizer {
	return &Authorizer{
		TdlibParameters: make(chan *TdlibParameters, 1),
		PhoneNumber:     make(chan string, 1),
		Code:            make(chan string, 1),
		State:           make(chan AuthorizationState, 10),
		FirstName:       make(chan string, 1),
		LastName:        make(chan string, 1),
		Password:        make(chan string, 1),
	}
}

// Handle is a function called during authorization.
func (a *Authorizer) Handle(client *Client, state AuthorizationState) error {
	a.State <- state
	switch state.AuthorizationStateType() {
	case TypeAuthorizationStateWaitTdlibParameters:
		_, err := client.SetTdlibParameters(&SetTdlibParametersRequest{
			Parameters: <-a.TdlibParameters,
		})
		return err
	case TypeAuthorizationStateWaitEncryptionKey:
		_, err := client.CheckDatabaseEncryptionKey(&CheckDatabaseEncryptionKeyRequest{})
		return err
	case TypeAuthorizationStateWaitPhoneNumber:
		_, err := client.SetAuthenticationPhoneNumber(&SetAuthenticationPhoneNumberRequest{
			PhoneNumber:          <-a.PhoneNumber,
			AllowFlashCall:       false,
			IsCurrentPhoneNumber: false,
		})
		return err
	case TypeAuthorizationStateWaitCode:
		_, err := client.CheckAuthenticationCode(&CheckAuthenticationCodeRequest{
			Code:      <-a.Code,
			FirstName: <-a.FirstName,
			LastName:  <-a.LastName,
		})
		return err
	case TypeAuthorizationStateWaitPassword:
		_, err := client.CheckAuthenticationPassword(&CheckAuthenticationPasswordRequest{
			Password: <-a.Password,
		})
		return err
	case TypeAuthorizationStateReady:
		return nil
	case TypeAuthorizationStateLoggingOut:
		return nil
	case TypeAuthorizationStateClosing:
		return nil
	case TypeAuthorizationStateClosed:
		return nil
	}
	return nil
}

// Close is a function called when authorization is done or cancelled.
func (a *Authorizer) Close() {
	close(a.TdlibParameters)
	close(a.PhoneNumber)
	close(a.Code)
	close(a.State)
	close(a.FirstName)
	close(a.LastName)
	close(a.Password)
}

// CliInteractor is a functions that implements simple authorization helper asking authorization data from terminal.
func CliInteractor(a *Authorizer) {
	for {
		select {
		case state, ok := <-a.State:
			if !ok {
				return
			}

			switch state.AuthorizationStateType() {
			case TypeAuthorizationStateWaitPhoneNumber:
				fmt.Println("Enter phone number: ")
				var phoneNumber string
				fmt.Scanln(&phoneNumber)

				a.PhoneNumber <- phoneNumber

			case TypeAuthorizationStateWaitCode:
				var code string
				var firstName string
				var lastName string

				fmt.Println("Enter code: ")
				fmt.Scanln(&code)

				if !state.(*AuthorizationStateWaitCode).IsRegistered {
					fmt.Println("Phone number is not registered.")

					fmt.Println("Enter first name: ")
					fmt.Scanln(&firstName)

					fmt.Println("Enter last name: ")
					fmt.Scanln(&lastName)
				}

				a.Code <- code
				a.FirstName <- firstName
				a.LastName <- lastName

			case TypeAuthorizationStateWaitPassword:
				fmt.Println("Enter password: ")
				var password string
				fmt.Scanln(&password)

				a.Password <- password

			case TypeAuthorizationStateReady:
				return
			}
		}
	}
}

// BotAuthorizer implements AuthorizationStateHandler interface.
type BotAuthorizer struct {
	TdlibParameters chan *TdlibParameters
	Token           chan string
	State           chan AuthorizationState
}

// NewBotAuthorizer creates new instance of BotAuthorizer.
func NewBotAuthorizer(token string) *BotAuthorizer {
	a := &BotAuthorizer{
		TdlibParameters: make(chan *TdlibParameters, 1),
		Token:           make(chan string, 1),
		State:           make(chan AuthorizationState, 10),
	}

	a.Token <- token

	return a
}

// Handle is a function called during authorization.
func (a *BotAuthorizer) Handle(client *Client, state AuthorizationState) error {
	a.State <- state

	switch state.AuthorizationStateType() {
	case TypeAuthorizationStateWaitTdlibParameters:
		_, err := client.SetTdlibParameters(&SetTdlibParametersRequest{
			Parameters: <-a.TdlibParameters,
		})
		return err

	case TypeAuthorizationStateWaitEncryptionKey:
		_, err := client.CheckDatabaseEncryptionKey(&CheckDatabaseEncryptionKeyRequest{})
		return err

	case TypeAuthorizationStateWaitPhoneNumber:
		_, err := client.CheckAuthenticationBotToken(&CheckAuthenticationBotTokenRequest{
			Token: <-a.Token,
		})
		return err

	case TypeAuthorizationStateWaitCode:
		return nil

	case TypeAuthorizationStateWaitPassword:
		return nil

	case TypeAuthorizationStateReady:
		return nil

	case TypeAuthorizationStateLoggingOut:
		return nil

	case TypeAuthorizationStateClosing:
		return nil

	case TypeAuthorizationStateClosed:
		return nil
	}

	return nil
}

// Close is a function called when authorization is done or cancelled.
func (a *BotAuthorizer) Close() {
	close(a.TdlibParameters)
	close(a.Token)
	close(a.State)
}
