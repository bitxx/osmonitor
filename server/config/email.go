package config

type Email struct {
	Host          string
	Port          int
	Username      string
	Password      string
	FromEmail     string
	ContentType   string
	ToEmail       string
	SubjectPrefix string
	DelayTime     int
}

var EmailConfig = new(Email)
