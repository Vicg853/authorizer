package models

import "github.com/authorizerdev/authorizer/server/graph/model"

// Note: any change here should be reflected in providers/casandra/provider.go as it does not have model support in collection creation

// VerificationRequest model for db
type VerificationRequest struct {
	Key         string `json:"_key,omitempty" bson:"_key" cql:"_key,omitempty"` // for arangodb
	ID          string `gorm:"primaryKey;type:char(36)" json:"_id" bson:"_id" cql:"id"`
	Token       string `gorm:"type:text" json:"token" bson:"token" cql:"jwt_token"` // token is reserved keyword in cassandra
	Identifier  string `gorm:"uniqueIndex:idx_email_identifier;type:varchar(64)" json:"identifier" bson:"identifier" cql:"identifier"`
	ExpiresAt   int64  `json:"expires_at" bson:"expires_at" cql:"expires_at"`
	Email       string `gorm:"uniqueIndex:idx_email_identifier;type:varchar(256)" json:"email" bson:"email" cql:"email"`
	Nonce       string `gorm:"type:text" json:"nonce" bson:"nonce" cql:"nonce"`
	RedirectURI string `gorm:"type:text" json:"redirect_uri" bson:"redirect_uri" cql:"redirect_uri"`
	CreatedAt   int64  `json:"created_at" bson:"created_at" cql:"created_at"`
	UpdatedAt   int64  `json:"updated_at" bson:"updated_at" cql:"updated_at"`
}

func (v *VerificationRequest) AsAPIVerificationRequest() *model.VerificationRequest {
	token := v.Token
	createdAt := v.CreatedAt
	updatedAt := v.UpdatedAt
	email := v.Email
	nonce := v.Nonce
	redirectURI := v.RedirectURI
	expires := v.ExpiresAt
	identifier := v.Identifier
	return &model.VerificationRequest{
		ID:          v.ID,
		Token:       &token,
		Identifier:  &identifier,
		Expires:     &expires,
		Email:       &email,
		Nonce:       &nonce,
		RedirectURI: &redirectURI,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}
}
