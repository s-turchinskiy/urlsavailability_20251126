// Package service Сервис
package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/mailru/easyjson"
	"github.com/signintech/gopdf"

	"github.com/s-turchinskiy/urlsavailability/internal/repository"
	"github.com/s-turchinskiy/urlsavailability/internal/service/availabilityworker"
	"github.com/s-turchinskiy/urlsavailability/internal/utils/errutil"
	"github.com/s-turchinskiy/urlsavailability/models"
)

var ErrServiceNotAvailable = fmt.Errorf("service not available")

type Servicer interface {
	Availability(ctx context.Context, urls []string) (result models.URLsKit, num uint64, err error)
	GetPDF(ctx context.Context, nums []uint64) ([]byte, error)
	LoadDataFromFile(ctx context.Context) error
	SaveDataToFile(ctx context.Context) error
}

type Service struct {
	rep             repository.Repository
	timeoutRequests time.Duration
	rateLimit       int
	fileStoragePath string
}

func New(rep repository.Repository, timeoutRequests time.Duration, rateLimit int, fileStoragePath string) *Service {
	return &Service{
		rep:             rep,
		timeoutRequests: timeoutRequests,
		rateLimit:       rateLimit,
		fileStoragePath: fileStoragePath,
	}
}

func (s *Service) Availability(ctx context.Context, urls []string) (models.URLsKit, uint64, error) {

	select {
	case <-ctx.Done():
		return nil, 0, ErrServiceNotAvailable
	default:

	}

	kit := availabilityworker.New(urls, s.rateLimit, s.timeoutRequests).Result()
	num, err := s.rep.Update(ctx, kit)
	if err != nil {
		return nil, 0, err
	}
	return kit, num, nil

}

func (s *Service) GetPDF(ctx context.Context, nums []uint64) ([]byte, error) {

	data, err := s.rep.GetDataWithFilter(ctx, nums)
	if err != nil {
		return nil, err
	}

	pdf := &gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: *gopdf.PageSizeA4})
	pdf.AddPage()

	err = pdf.AddTTFFont("font1", "./assets/Roboto-Black.ttf")
	if err != nil {
		return nil, err
	}
	err = pdf.SetFont("font1", "", 11)
	if err != nil {
		return nil, err
	}

	tableStartY := 10.0
	marginLeft := 10.0

	table := pdf.NewTableLayout(marginLeft, tableStartY, 20, len(data))

	table.AddColumn("URL", 300, "left")
	table.AddColumn("availability", 200, "left")

	viewedData := data.ConvertToReadableView()
	urls := make([]string, 0, len(viewedData))
	for k := range viewedData {
		urls = append(urls, k)
	}
	sort.Strings(urls)

	for _, url := range urls {
		table.AddRow([]string{url, viewedData[url]})
	}

	table.SetTableStyle(gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{
			Top:    true,
			Left:   true,
			Bottom: true,
			Right:  true,
			Width:  1.0,
		},
		FillColor: gopdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: gopdf.RGBColor{R: 0, G: 0, B: 0},
		FontSize:  10,
	})

	table.SetHeaderStyle(gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{
			Top:      true,
			Left:     true,
			Bottom:   true,
			Right:    true,
			Width:    2.0,
			RGBColor: gopdf.RGBColor{R: 100, G: 150, B: 255},
		},
		FillColor: gopdf.RGBColor{R: 255, G: 200, B: 200},
		TextColor: gopdf.RGBColor{R: 255, G: 100, B: 100},
		Font:      "font1",
		FontSize:  12,
	})

	table.SetCellStyle(gopdf.CellStyle{
		BorderStyle: gopdf.BorderStyle{
			Right:    true,
			Bottom:   true,
			Width:    0.5,
			RGBColor: gopdf.RGBColor{R: 0, G: 0, B: 0},
		},
		FillColor: gopdf.RGBColor{R: 255, G: 255, B: 255},
		TextColor: gopdf.RGBColor{R: 0, G: 0, B: 0},
		Font:      "font1",
		FontSize:  10,
	})

	err = table.DrawTable()
	if err != nil {
		return nil, err
	}

	return pdf.GetBytesPdf(), nil
}

// SaveDataToFile Сохранение данных в файл
func (s *Service) SaveDataToFile(ctx context.Context) error {

	data, err := s.rep.GetAllData(ctx)
	if err != nil {
		return err
	}

	bytes, err := easyjson.Marshal(data)
	if err != nil {
		return err
	}

	err = os.WriteFile(s.fileStoragePath, bytes, 0666)
	if err != nil {
		return err
	}

	log.Printf("data saved to file %s\n", s.fileStoragePath)

	return err

}

// LoadDataFromBytes Загрузка данных из массива байт
func (s *Service) LoadDataFromBytes(ctx context.Context, bytes []byte) error {

	data := &models.FileStore{}

	if err := easyjson.Unmarshal(bytes, data); err != nil {
		return err
	}

	err := s.rep.LoadAllData(ctx, *data)
	if err != nil {
		return err
	}

	fmt.Printf("data loaded from file %s\n", s.fileStoragePath)

	return err
}

// LoadDataFromFile Загрузка данных из файла
func (s *Service) LoadDataFromFile(ctx context.Context) error {

	bytes, err := os.ReadFile(s.fileStoragePath)

	if errors.Is(err, os.ErrNotExist) {
		dir, err2 := os.Getwd()
		if err2 != nil {
			return errutil.WrapError(fmt.Errorf("couldn't get the current directory, %w", err2))
		}

		fmt.Printf("data has not been loaded from file %s%s, file not exist\n", dir, s.fileStoragePath)
		return nil

	}

	if err != nil {
		return err
	}

	return s.LoadDataFromBytes(ctx, bytes)

}
