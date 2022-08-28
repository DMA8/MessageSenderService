package ports

import "gitlab.com/g6834/team31/mailsender/internal/domain/models"

type MailSender interface {
	AcceptMail(email models.Email)
}
