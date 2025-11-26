// Package config Получение конфигурации
package config

import (
	"errors"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/s-turchinskiy/urlsavailability/internal/utils/errutil"

	"github.com/caarlos0/env/v11"
)

type netAddress struct {
	Host string
	Port int
}

type Config struct {
	Addr            *netAddress   `env:"ADDRESS"`           //Адрес запуска http сервера
	RateLimit       int           `env:"RATE_LIMIT"`        //Количество одновременно исходящих запросов на сервер
	URLTimeout      time.Duration `env:"URL_TIMEOUT"`       //Время ожидания доступности url
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT"`  //Максимальное время ожидания завершения процессов при выключении службы
	FileStoragePath string        `env:"FILE_STORAGE_PATH"` //Путь до файла, куда сохраняются (откуда загружаются) текущие значения
}

func LoadConfig() (*Config, error) {

	cfg := &Config{
		Addr:            &netAddress{Host: "localhost", Port: 8080},
		RateLimit:       runtime.NumCPU(),
		URLTimeout:      1 * time.Second,
		ShutdownTimeout: 20 * time.Second,
		FileStoragePath: "store.txt",
	}

	err := env.ParseWithOptions(cfg, env.Options{
		FuncMap: map[reflect.Type]env.ParserFunc{
			reflect.TypeOf(netAddress{}): func(incomingData string) (interface{}, error) {
				addr := netAddress{}
				err := addr.Set(incomingData)
				if err != nil {
					return nil, err
				}

				return addr, nil
			},
		},
	})

	if err != nil {
		return nil, errutil.WrapError(err)
	}

	return cfg, nil
}

func (a *netAddress) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func (a *netAddress) Set(s string) error {
	hp := strings.Split(s, ":")
	if len(hp) != 2 {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(hp[1])
	if err != nil {
		return err
	}
	a.Host = hp[0]
	a.Port = port
	return nil
}
