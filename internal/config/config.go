package config

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type file struct {
	Profiles map[string]profile
	Ollama   ollama
}

type ollama struct {
	Port       string
	Keep_alive string
}

type profile struct {
	Model  string
	System []string
}

type Config struct {
	Profile profile
	Ollama  ollama
}

func defaultOllama() ollama {
	return ollama{
		Port: "11434",
	}
}

func defaultFile() file {
	return file{
		Profiles: map[string]profile{
			"default": {
				Model:  "gemma2:27b",
				System: []string{},
			},
		},
		Ollama: defaultOllama(),
	}
}

func readf() file {
	config := defaultFile()

	home, err := os.UserHomeDir()
	if err != nil {
		log.Println("can't find home folder, using default config")
		return config
	}

	configFile, err := os.ReadFile(home + "/.config/teja/teja.yaml")
	if err != nil {
		fmt.Println("con't find config file, using defaults")
		return config
	}

	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		log.Fatalln(err)
	}

	return config
}

func Read() Config {
	file := readf()

	selected := "default"
	if len(os.Args) > 1 {
		selected = os.Args[1]
	}

	profile, ok := file.Profiles[selected]
	if !ok {
		log.Fatalln("profile not in the config")
	}

	return Config{
		Profile: profile,
		Ollama:  file.Ollama,
	}
}
