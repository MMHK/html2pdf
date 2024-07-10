package lib

import (
	"encoding/json"
	"os"
	"strconv"
)

type CleanerConfig struct {
	CleanupPeriod int `json:"period"`
	FileAgeLimit  int `json:"expire"`
}

type Config struct {
	ChromePath string `json:"chrome_path"`
	Listen     string `json:"listen"`
	WebRoot    string `json:"web_root"`
	Worker     int    `json:"worker"`
	Timeout    int    `json:"timeout"`
	Cleaner    *CleanerConfig `json:"cleaner"`
	save_path string
}

func NewConfig(filename string) (err error, c *Config) {
	c = &Config{
		Cleaner: &CleanerConfig {
			CleanupPeriod: 1800, // 30 minutes
			FileAgeLimit : 86400, // 1 day
		},
	}
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
	if os.Getenv("CLEANER_PERIOD") != "" {
		this.Cleaner.CleanupPeriod, _ = strconv.Atoi(os.Getenv("CLEANER_PERIOD"))
	}
	if os.Getenv("CLEANER_FILE_AGE_LIMIT") != "" {
		this.Cleaner.FileAgeLimit, _ = strconv.Atoi(os.Getenv("CLEANER_FILE_AGE_LIMIT"))
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
