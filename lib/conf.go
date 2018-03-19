package lib

import (
	"encoding/json"
	"os"
)

type Config struct {
	Listen     string   `json:"listen"`
	TempPath   string   `json:"tmp_path"`
	WebRoot    string   `json:"web_root"`
	WebKitBin  string   `json:"webkit_bin"`
	WebKitArgs []string `json:"webkit_args"`
	PDFTK      string   `json:"pdftk_bin"`
	Worker     int      `json:"worker"`
	Timeout    int      `json:timeout`
	save_path  string
}

func NewConfig(filename string) (err error, c *Config) {
	c = &Config{}
	c.save_path = filename
	err = c.load(filename)
	return
}

func (c *Config) load(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(c)
	if err != nil {
		ErrLogger.Println(err)
	}
	return err
}

func (c *Config) Save() error {
	file, err := os.Create(c.save_path)
	if err != nil {
		ErrLogger.Println(err)
		return err
	}
	defer file.Close()
	data, err2 := json.MarshalIndent(c, "", "    ")
	if err2 != nil {
		ErrLogger.Println(err2)
		return err2
	}
	_, err3 := file.Write(data)
	if err3 != nil {
		ErrLogger.Println(err3)
	}
	return err3
}
