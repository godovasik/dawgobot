package ollama

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"

	"golang.org/x/image/draw"
)

var (
	ErrNotAnImage = errors.New("it's not an image")
)

func OpenImage(imagePath string) ([]byte, error) {
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("файл %s не найден", imagePath)
	}

	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		return nil, fmt.Errorf("cant decode this file: %w", err)
	}
	return imageBytes, nil
}

// ResizeImageBytes работает напрямую с байтами (более эффективно)
func ResizeImageBytes(imageBytes []byte) ([]byte, error) {
	// Создаем reader из байтов
	reader := bytes.NewReader(imageBytes)

	// Декодируем изображение
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	// Получаем размеры исходного изображения
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Проверяем, нужно ли сжимать
	maxSize := 1024
	if width <= maxSize && height <= maxSize {
		// Изображение уже подходящего размера, возвращаем как есть
		return imageBytes, nil
	}

	// Вычисляем новые размеры с сохранением пропорций
	var newWidth, newHeight int
	if width > height {
		newWidth = maxSize
		newHeight = int(float64(height) * float64(maxSize) / float64(width))
	} else {
		newHeight = maxSize
		newWidth = int(float64(width) * float64(maxSize) / float64(height))
	}

	// Создаем новое изображение
	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Масштабируем изображение
	draw.BiLinear.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Кодируем обратно в байты
	var buf bytes.Buffer

	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(&buf, dst)
	default:
		err = jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func CheckUrl(url string) (bool, error) {
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = "https://" + url
	}
	resp, err := http.Head(url)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	return strings.HasPrefix(contentType, "image/"), nil
}

func GetImage(url string) ([]byte, error) {
	ok, err := CheckUrl(url)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, ErrNotAnImage
	}
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
