package services

import (
	"fmt"
	"net/smtp"
)

type EmailService struct {
	host string
	port string
}

func NewEmailService(host, port string) *EmailService {
	return &EmailService{
		host: host,
		port: port,
	}
}

// SendOTP envia um e-mail simples em texto/HTML com o código para o usuário.
func (s *EmailService) SendOTP(to string, code string) error {
	from := "noreply@nayz.tech"
	
	// Construindo o corpo do e-mail no formato MIME simples
	subject := "Subject: Seu Código de Acesso\r\n"
	mime := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	body := fmt.Sprintf(`
		<h2>Olá!</h2>
		<p>O seu código de acesso temporário é: <strong>%s</strong></p>
		<p>Este código expira em 5 minutos.</p>
		<p>Se você não solicitou este acesso, por favor ignore este e-mail.</p>
	`, code)

	msg := []byte(subject + mime + body)

	// Como estamos usando Mailpit (porta 1025 local sem SSL/TLS forçado), 
	// passamos 'nil' para a autenticação. Se fosse SendGrid, usaríamos smtp.PlainAuth()
	addr := s.host + ":" + s.port
	
	// A biblioteca nativa envia a mensagem em milissegundos
	err := smtp.SendMail(addr, nil, from, []string{to}, msg)
	if err != nil {
		return fmt.Errorf("falha ao enviar o e-mail: %v", err)
	}
	return nil
}
