package config

type App struct {
	Name      string
	Version   string
	Secret    string
	ServerUrl string
	ProcNames string
	IsPing    bool
	DelayTime uint
}

var AppConfig = new(App)
