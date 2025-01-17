package cassandradb

import (
	"fmt"
	"time"

	"github.com/authorizerdev/authorizer/server/db/models"
	"github.com/google/uuid"
)

// AddSession to save session information in database
func (p *provider) AddSession(session models.Session) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	session.CreatedAt = time.Now().Unix()
	session.UpdatedAt = time.Now().Unix()

	insertSessionQuery := fmt.Sprintf("INSERT INTO %s (id, user_id, user_agent, ip, created_at, updated_at) VALUES ('%s', '%s', '%s', '%s', %d, %d)", KeySpace+"."+models.Collections.Session, session.ID, session.UserID, session.UserAgent, session.IP, session.CreatedAt, session.UpdatedAt)
	err := p.db.Query(insertSessionQuery).Exec()
	if err != nil {
		return err
	}
	return nil
}

// DeleteSession to delete session information from database
func (p *provider) DeleteSession(userId string) error {
	deleteSessionQuery := fmt.Sprintf("DELETE FROM %s WHERE user_id = '%s'", KeySpace+"."+models.Collections.Session, userId)
	err := p.db.Query(deleteSessionQuery).Exec()
	if err != nil {
		return err
	}
	return nil
}
