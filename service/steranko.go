package service

import (
	"github.com/EmissarySocial/emissary/model"
	"github.com/benpate/derp"
	"github.com/benpate/steranko"
	"github.com/golang-jwt/jwt/v5"
)

// SterankoUserService is a wrapper/adapter that makes the User service compatable with Steranko.
type SterankoUserService struct {
	userService *User
	domainEmail *DomainEmail
}

// NewSterankoUserService returns a fully populated SterankoUserService.
func NewSterankoUserService(userService *User, domainEmail *DomainEmail) SterankoUserService {
	return SterankoUserService{
		userService: userService,
		domainEmail: domainEmail,
	}
}

// New creates a newly initialized User that is ready to use
func (service SterankoUserService) New() steranko.User {
	result := model.NewUser()
	return &result
}

// Load retrieves a single User from the database
func (service SterankoUserService) Load(username string, result steranko.User) error {

	user, ok := result.(*model.User)

	if !ok {
		return derp.NewInternalError("service.SterankoUserService.Load", "Invalid result provided.  This should never happen")
	}

	if err := service.userService.LoadByUsernameOrEmail(username, user); err != nil {
		return derp.Wrap(err, "service.SterankoUserService.Load", "Error loading user")
	}

	return nil
}

// Save inserts/updates a single User in the database
func (service SterankoUserService) Save(user steranko.User, comment string) error {

	if user, ok := user.(*model.User); ok {
		return service.userService.Save(user, comment)
	}

	return derp.NewInternalError("service.SterankoUserService.Save", "Steranko User is not a valid object.  This should never happen", user)
}

// Delete removes a single User from the database
func (service SterankoUserService) Delete(user steranko.User, comment string) error {

	if user, ok := user.(*model.User); ok {
		return service.userService.Delete(user, comment)
	}

	return derp.NewInternalError("service.SterankoUserService.Delete", "Steranko User is not a valid object.  This should never happen", user)
}

// RequestPasswordReset is not currently implemented in this service. (TODO)
func (service SterankoUserService) RequestPasswordReset(user steranko.User) error {

	if user, ok := user.(*model.User); ok {
		return service.domainEmail.SendPasswordReset(user)
	}

	return derp.NewInternalError("service.SterankoUserService.Save", "Steranko User is not a valid object.  This should never happen", user)
}

// NewClaims creates a new JWT claim object
func (service SterankoUserService) NewClaims() jwt.Claims {
	result := model.NewAuthorization()
	return &result
}

// Close is required to implement the steranko.UserService interface
func (service SterankoUserService) Close() {

}
