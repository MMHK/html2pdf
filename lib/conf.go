package lib

import (
	"encoding/json"
	"os"
	"strconv"
)

type Config struct {
	ChromePath string `json:"chrome_path"`
	Listen     string `json:"listen"`
	WebRoot    string `json:"web_root"`
	Worker     int    `json:"worker"`
	Timeout    int    `json:"timeout"`
	save_path  string
}

func NewConfig(filename string) (err error, c *Config) {
	c = &Config{}
	c.save_path = filename
	err = c.load(filename)
	return
}

func (this *Config) LoadWithENV() *Config {
	if os.Getenv("LISTEN") != "" {
		this.Listen = os.Getenv("LISTEN")
	}
	if os.Getenv("WEB_ROOT") != "" {
		this.WebRoot = os.Getenv("WEB_ROOT")
	}
	if os.Getenv("WORKER") != "" {
		this.Worker, _ = strconv.Atoi(os.Getenv("WORKER"))
	}
	if os.Getenv("TIMEOUT") != "" {
		this.Timeout, _ = strconv.Atoi(os.Getenv("TIMEOUT"))
	}
	if os.Getenv("CHROME_PATH") != "" {
		this.ChromePath = os.Getenv("CHROME_PATH")
	}

	return this
}

func (c *Config) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		Log.Error(err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		Log.Error(err)
	}
	return err
}

func (c *Config) Save() error {
	file, err := os.Create(c.save_path)
	if err != nil {
		Log.Error(err)
		return err
	}
	defer file.Close()
	data, err := json.MarshalIndent(c, "", "    ")
	if err != nil {
		Log.Error(err)
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		Log.Error(err)
	}
	return err
}
