package config

import(
	"os"
)

const configFileName = ".gatorconfig.json"

type Config struct{
	DBUrl string `json:"db_url"`
	CurrentUsername string `json:"current_user_name"`
}


func Read() Config {

}


func (cfg Config) SetUser