package mailer

import "github.com/google/uuid"

//go:generate mockery --name=Mailer --inpackage --case=snake

type Mailer interface {
	SendVerificationEmail(to string, token uuid.UUID) error
}
