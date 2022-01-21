package resolvers

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/authorizerdev/authorizer/server/constants"
	"github.com/authorizerdev/authorizer/server/db"
	"github.com/authorizerdev/authorizer/server/db/models"
	"github.com/authorizerdev/authorizer/server/email"
	"github.com/authorizerdev/authorizer/server/envstore"
	"github.com/authorizerdev/authorizer/server/graph/model"
	"github.com/authorizerdev/authorizer/server/session"
	"github.com/authorizerdev/authorizer/server/utils"
)

// SignupResolver is a resolver for signup mutation
func SignupResolver(ctx context.Context, params model.SignUpInput) (*model.AuthResponse, error) {
	gc, err := utils.GinContextFromContext(ctx)
	var res *model.AuthResponse
	if err != nil {
		return res, err
	}

	if envstore.EnvInMemoryStoreObj.GetBoolStoreEnvVariable(constants.EnvKeyDisableBasicAuthentication) {
		return res, fmt.Errorf(`basic authentication is disabled for this instance`)
	}
	if params.ConfirmPassword != params.Password {
		return res, fmt.Errorf(`password and confirm password does not match`)
	}

	params.Email = strings.ToLower(params.Email)

	if !utils.IsValidEmail(params.Email) {
		return res, fmt.Errorf(`invalid email address`)
	}

	// find user with email
	existingUser, err := db.Provider.GetUserByEmail(params.Email)
	if err != nil {
		log.Println("user with email " + params.Email + " not found")
	}

	if existingUser.EmailVerifiedAt != nil {
		// email is verified
		return res, fmt.Errorf(`%s has already signed up`, params.Email)
	} else if existingUser.ID != "" && existingUser.EmailVerifiedAt == nil {
		return res, fmt.Errorf("%s has already signed up. please complete the email verification process or reset the password", params.Email)
	}

	inputRoles := []string{}

	if len(params.Roles) > 0 {
		// check if roles exists
		if !utils.IsValidRoles(envstore.EnvInMemoryStoreObj.GetSliceStoreEnvVariable(constants.EnvKeyRoles), params.Roles) {
			return res, fmt.Errorf(`invalid roles`)
		} else {
			inputRoles = params.Roles
		}
	} else {
		inputRoles = envstore.EnvInMemoryStoreObj.GetSliceStoreEnvVariable(constants.EnvKeyDefaultRoles)
	}

	user := models.User{
		Email: params.Email,
	}

	user.Roles = strings.Join(inputRoles, ",")

	password, _ := utils.EncryptPassword(params.Password)
	user.Password = &password

	if params.GivenName != nil {
		user.GivenName = params.GivenName
	}

	if params.FamilyName != nil {
		user.FamilyName = params.FamilyName
	}

	if params.MiddleName != nil {
		user.MiddleName = params.MiddleName
	}

	if params.Nickname != nil {
		user.Nickname = params.Nickname
	}

	if params.Gender != nil {
		user.Gender = params.Gender
	}

	if params.Birthdate != nil {
		user.Birthdate = params.Birthdate
	}

	if params.PhoneNumber != nil {
		user.PhoneNumber = params.PhoneNumber
	}

	if params.Picture != nil {
		user.Picture = params.Picture
	}

	user.SignupMethods = constants.SignupMethodBasicAuth
	if envstore.EnvInMemoryStoreObj.GetBoolStoreEnvVariable(constants.EnvKeyDisableEmailVerification) {
		now := time.Now().Unix()
		user.EmailVerifiedAt = &now
	}
	user, err = db.Provider.AddUser(user)
	if err != nil {
		return res, err
	}
	userIdStr := fmt.Sprintf("%v", user.ID)
	roles := strings.Split(user.Roles, ",")
	userToReturn := utils.GetResponseUserData(user)

	if !envstore.EnvInMemoryStoreObj.GetBoolStoreEnvVariable(constants.EnvKeyDisableEmailVerification) {
		// insert verification request
		verificationType := constants.VerificationTypeBasicAuthSignup
		token, err := utils.CreateVerificationToken(params.Email, verificationType)
		if err != nil {
			log.Println(`error generating token`, err)
		}
		db.Provider.AddVerificationRequest(models.VerificationRequest{
			Token:      token,
			Identifier: verificationType,
			ExpiresAt:  time.Now().Add(time.Minute * 30).Unix(),
			Email:      params.Email,
		})

		// exec it as go routin so that we can reduce the api latency
		go func() {
			email.SendVerificationMail(params.Email, token)
		}()

		res = &model.AuthResponse{
			Message: `Verification email has been sent. Please check your inbox`,
			User:    userToReturn,
		}
	} else {

		refreshToken, _, _ := utils.CreateAuthToken(user, constants.TokenTypeRefreshToken, roles)

		accessToken, expiresAt, _ := utils.CreateAuthToken(user, constants.TokenTypeAccessToken, roles)

		session.SetUserSession(userIdStr, accessToken, refreshToken)
		utils.SaveSessionInDB(user.ID, gc)
		res = &model.AuthResponse{
			Message:     `Signed up successfully.`,
			AccessToken: &accessToken,
			ExpiresAt:   &expiresAt,
			User:        userToReturn,
		}

		utils.SetCookie(gc, accessToken)
	}

	return res, nil
}
