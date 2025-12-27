package email

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"html/template"
	"net/smtp"
	"strings"
	"sync"
)

// Config holds email configuration
type Config struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
	FromName string
	UseTLS   bool
}

// Service provides email sending functionality
type Service struct {
	config    *Config
	templates map[string]*template.Template
	mu        sync.RWMutex
}

// New creates a new email service
func New(config *Config) *Service {
	s := &Service{
		config:    config,
		templates: make(map[string]*template.Template),
	}
	s.loadDefaultTemplates()
	return s
}

// Message represents an email message
type Message struct {
	To       []string
	Subject  string
	Body     string
	IsHTML   bool
	Template string
	Data     interface{}
}

// Send sends an email
func (s *Service) Send(msg *Message) error {
	if msg.Template != "" {
		body, err := s.RenderTemplate(msg.Template, msg.Data)
		if err != nil {
			return fmt.Errorf("failed to render template: %w", err)
		}
		msg.Body = body
		msg.IsHTML = true
	}

	return s.sendMail(msg)
}

// sendMail sends an email via SMTP
func (s *Service) sendMail(msg *Message) error {
	// Build message
	var msgBuilder strings.Builder

	msgBuilder.WriteString(fmt.Sprintf("From: %s <%s>\r\n", s.config.FromName, s.config.From))
	msgBuilder.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(msg.To, ", ")))
	msgBuilder.WriteString(fmt.Sprintf("Subject: %s\r\n", msg.Subject))
	msgBuilder.WriteString("MIME-Version: 1.0\r\n")

	if msg.IsHTML {
		msgBuilder.WriteString("Content-Type: text/html; charset=UTF-8\r\n")
	} else {
		msgBuilder.WriteString("Content-Type: text/plain; charset=UTF-8\r\n")
	}

	msgBuilder.WriteString("\r\n")
	msgBuilder.WriteString(msg.Body)

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	auth := smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)

	if s.config.UseTLS {
		return s.sendWithTLS(addr, auth, msg.To, []byte(msgBuilder.String()))
	}

	return smtp.SendMail(addr, auth, s.config.From, msg.To, []byte(msgBuilder.String()))
}

// sendWithTLS sends email with TLS connection
func (s *Service) sendWithTLS(addr string, auth smtp.Auth, to []string, msg []byte) error {
	tlsConfig := &tls.Config{
		ServerName: s.config.Host,
	}

	conn, err := tls.Dial("tcp", addr, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer conn.Close()

	client, err := smtp.NewClient(conn, s.config.Host)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %w", err)
	}
	defer client.Close()

	if err := client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate: %w", err)
	}

	if err := client.Mail(s.config.From); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	for _, addr := range to {
		if err := client.Rcpt(addr); err != nil {
			return fmt.Errorf("failed to set recipient %s: %w", addr, err)
		}
	}

	w, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to open data writer: %w", err)
	}

	_, err = w.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	if err := w.Close(); err != nil {
		return fmt.Errorf("failed to close data writer: %w", err)
	}

	return client.Quit()
}

// RegisterTemplate registers a new email template
func (s *Service) RegisterTemplate(name, content string) error {
	tmpl, err := template.New(name).Parse(content)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates[name] = tmpl

	return nil
}

// RenderTemplate renders a template with data
func (s *Service) RenderTemplate(name string, data interface{}) (string, error) {
	s.mu.RLock()
	tmpl, ok := s.templates[name]
	s.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("template %s not found", name)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}

// Template names
const (
	TemplateVerifyEmail     = "verify_email"
	TemplatePasswordReset   = "password_reset"
	TemplatePhotoApproved   = "photo_approved"
	TemplatePhotoRejected   = "photo_rejected"
	TemplateWelcome         = "welcome"
)

// loadDefaultTemplates loads default email templates
func (s *Service) loadDefaultTemplates() {
	// Welcome email template
	s.RegisterTemplate(TemplateWelcome, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>欢迎加入 QuanPhotos</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c3e50;">欢迎加入 QuanPhotos!</h1>
        <p>亲爱的 {{.Username}},</p>
        <p>感谢您注册 QuanPhotos 航空摄影社区！</p>
        <p>现在您可以：</p>
        <ul>
            <li>上传您的航空摄影作品</li>
            <li>浏览和收藏精彩的航空照片</li>
            <li>与其他航空摄影爱好者交流</li>
        </ul>
        <p>如有任何问题，欢迎联系我们。</p>
        <p>祝您使用愉快！</p>
        <p>QuanPhotos 团队</p>
    </div>
</body>
</html>
`)

	// Email verification template
	s.RegisterTemplate(TemplateVerifyEmail, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>验证您的邮箱</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c3e50;">验证您的邮箱</h1>
        <p>亲爱的 {{.Username}},</p>
        <p>请点击下面的按钮验证您的邮箱地址：</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="{{.VerifyURL}}" style="background-color: #3498db; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">验证邮箱</a>
        </p>
        <p>或者复制以下链接到浏览器：</p>
        <p style="word-break: break-all; color: #666;">{{.VerifyURL}}</p>
        <p>此链接将在 {{.ExpireHours}} 小时后失效。</p>
        <p>如果您没有注册 QuanPhotos 账号，请忽略此邮件。</p>
        <p>QuanPhotos 团队</p>
    </div>
</body>
</html>
`)

	// Password reset template
	s.RegisterTemplate(TemplatePasswordReset, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>重置密码</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #2c3e50;">重置您的密码</h1>
        <p>亲爱的 {{.Username}},</p>
        <p>我们收到了重置您密码的请求。点击下面的按钮设置新密码：</p>
        <p style="text-align: center; margin: 30px 0;">
            <a href="{{.ResetURL}}" style="background-color: #e74c3c; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">重置密码</a>
        </p>
        <p>或者复制以下链接到浏览器：</p>
        <p style="word-break: break-all; color: #666;">{{.ResetURL}}</p>
        <p>此链接将在 {{.ExpireHours}} 小时后失效。</p>
        <p>如果您没有请求重置密码，请忽略此邮件，您的密码不会改变。</p>
        <p>QuanPhotos 团队</p>
    </div>
</body>
</html>
`)

	// Photo approved template
	s.RegisterTemplate(TemplatePhotoApproved, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>您的照片已通过审核</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #27ae60;">恭喜！您的照片已通过审核</h1>
        <p>亲爱的 {{.Username}},</p>
        <p>您上传的照片 <strong>{{.PhotoTitle}}</strong> 已通过审核并发布成功！</p>
        <p style="text-align: center; margin: 20px 0;">
            <a href="{{.PhotoURL}}" style="background-color: #27ae60; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">查看照片</a>
        </p>
        <p>感谢您为 QuanPhotos 社区贡献优质内容！</p>
        <p>QuanPhotos 团队</p>
    </div>
</body>
</html>
`)

	// Photo rejected template
	s.RegisterTemplate(TemplatePhotoRejected, `
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>您的照片未通过审核</title>
</head>
<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
    <div style="max-width: 600px; margin: 0 auto; padding: 20px;">
        <h1 style="color: #e74c3c;">您的照片未通过审核</h1>
        <p>亲爱的 {{.Username}},</p>
        <p>很抱歉，您上传的照片 <strong>{{.PhotoTitle}}</strong> 未能通过审核。</p>
        {{if .Reason}}
        <p><strong>原因：</strong>{{.Reason}}</p>
        {{end}}
        <p>您可以根据反馈修改后重新上传，或者提交工单进行申诉。</p>
        <p style="text-align: center; margin: 20px 0;">
            <a href="{{.AppealURL}}" style="background-color: #3498db; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">提交申诉</a>
        </p>
        <p>如有疑问，欢迎联系我们。</p>
        <p>QuanPhotos 团队</p>
    </div>
</body>
</html>
`)
}
