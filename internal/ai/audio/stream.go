package audio

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

// 400 строк пиздец
// вайбкод пацаны

// Конфигурация
type Config struct {
	AssemblyAIKey string
	TwitchChannel string
	SaveAudio     bool
	OutputDir     string
}

// Сообщения WebSocket
type WSMessage struct {
	MessageType string `json:"message_type"`
}

type SessionBegins struct {
	MessageType string `json:"message_type"`
	SessionID   string `json:"session_id"`
	ExpiresAt   string `json:"expires_at"`
}

type PartialTranscript struct {
	MessageType string  `json:"message_type"`
	AudioStart  int     `json:"audio_start"`
	AudioEnd    int     `json:"audio_end"`
	Confidence  float64 `json:"confidence"`
	Text        string  `json:"text"`
	Words       []Word  `json:"words"`
}

type FinalTranscript struct {
	MessageType string  `json:"message_type"`
	AudioStart  int     `json:"audio_start"`
	AudioEnd    int     `json:"audio_end"`
	Confidence  float64 `json:"confidence"`
	Text        string  `json:"text"`
	Words       []Word  `json:"words"`
}

type Word struct {
	Start      int     `json:"start"`
	End        int     `json:"end"`
	Confidence float64 `json:"confidence"`
	Text       string  `json:"text"`
}

type AudioData struct {
	AudioData string `json:"audio_data"`
}

type Transcriber struct {
	config      Config
	conn        *websocket.Conn
	audioCmd    *exec.Cmd
	mu          sync.Mutex
	isRunning   bool
	transcripts []string
}

func NewTranscriber(config Config) *Transcriber {
	return &Transcriber{
		config:      config,
		transcripts: make([]string, 0),
	}
}

func (t *Transcriber) Start() error {
	// Подключение к AssemblyAI WebSocket
	if err := t.connectWebSocket(); err != nil {
		return fmt.Errorf("failed to connect WebSocket: %v", err)
	}

	// Запуск аудио стрима
	if err := t.startAudioStream(); err != nil {
		return fmt.Errorf("failed to start audio stream: %v", err)
	}

	t.isRunning = true
	return nil
}

func (t *Transcriber) connectWebSocket() error {
	header := http.Header{}
	header.Set("Authorization", t.config.AssemblyAIKey)

	conn, _, err := websocket.DefaultDialer.Dial("wss://api.assemblyai.com/v2/realtime/ws?sample_rate=16000", header)
	if err != nil {
		return err
	}

	t.conn = conn

	// Запуск горутины для чтения сообщений
	go t.readMessages()

	return nil
}

func (t *Transcriber) readMessages() {
	defer t.conn.Close()

	for {
		_, message, err := t.conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error: %v", err)
			return
		}

		t.handleMessage(message)
	}
}

func (t *Transcriber) handleMessage(message []byte) {
	var baseMsg WSMessage
	if err := json.Unmarshal(message, &baseMsg); err != nil {
		log.Printf("Failed to parse message: %v", err)
		return
	}

	switch baseMsg.MessageType {
	case "session_begins":
		var msg SessionBegins
		json.Unmarshal(message, &msg)
		fmt.Printf("🟢 Session started: %s\n", msg.SessionID)

	case "partial_transcript":
		var msg PartialTranscript
		json.Unmarshal(message, &msg)
		if strings.TrimSpace(msg.Text) != "" {
			fmt.Printf("📝 Partial: %s\n", msg.Text)
		}

	case "final_transcript":
		var msg FinalTranscript
		json.Unmarshal(message, &msg)
		if strings.TrimSpace(msg.Text) != "" {
			timestamp := time.Now().Format("15:04:05")
			finalText := fmt.Sprintf("[%s] %s", timestamp, msg.Text)
			fmt.Printf("✅ Final: %s\n", finalText)

			t.mu.Lock()
			t.transcripts = append(t.transcripts, finalText)
			t.mu.Unlock()
		}

	case "session_terminated":
		fmt.Println("🔴 Session terminated")

	default:
		log.Printf("Unknown message type: %s", baseMsg.MessageType)
	}
}

func (t *Transcriber) startAudioStream() error {
	// Команда для получения аудио с Twitch через streamlink и конвертации в PCM
	args := []string{
		"-f", "lavfi",
		"-i", fmt.Sprintf("amovie=pipe\\:0,aresample=16000:resampler=swr[out0]"),
		"-f", "s16le",
		"-acodec", "pcm_s16le",
		"-ac", "1",
		"-ar", "16000",
		"pipe:1",
	}

	// Запуск streamlink для получения аудио
	streamCmd := exec.Command("streamlink",
		"--stdout",
		"--player-external-http",
		fmt.Sprintf("twitch.tv/%s", t.config.TwitchChannel),
		"audio_only")

	// Запуск ffmpeg для конвертации
	ffmpegCmd := exec.Command("ffmpeg", args...)

	// Соединяем команды через pipe
	streamOut, err := streamCmd.StdoutPipe()
	if err != nil {
		return err
	}

	ffmpegCmd.Stdin = streamOut
	ffmpegOut, err := ffmpegCmd.StdoutPipe()
	if err != nil {
		return err
	}

	// Запускаем команды
	if err := streamCmd.Start(); err != nil {
		return fmt.Errorf("failed to start streamlink: %v", err)
	}

	if err := ffmpegCmd.Start(); err != nil {
		streamCmd.Process.Kill()
		return fmt.Errorf("failed to start ffmpeg: %v", err)
	}

	t.audioCmd = ffmpegCmd

	// Опционально сохраняем аудио
	var audioWriter io.Writer
	var audioFile *os.File

	if t.config.SaveAudio {
		filename := fmt.Sprintf("%s/twitch_%s_%s.raw",
			t.config.OutputDir,
			t.config.TwitchChannel,
			time.Now().Format("20060102_150405"))

		audioFile, err = os.Create(filename)
		if err != nil {
			log.Printf("Failed to create audio file: %v", err)
		} else {
			fmt.Printf("💾 Saving audio to: %s\n", filename)
			audioWriter = io.MultiWriter(audioFile)
		}
	}

	// Запуск горутины для обработки аудио
	go t.processAudioStream(ffmpegOut, audioWriter, audioFile)

	return nil
}

func (t *Transcriber) processAudioStream(reader io.Reader, audioWriter io.Writer, audioFile *os.File) {
	defer func() {
		if audioFile != nil {
			audioFile.Close()
		}
	}()

	buffer := make([]byte, 3200) // 100ms при 16kHz, 16-bit mono

	for t.isRunning {
		n, err := reader.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("Audio read error: %v", err)
			}
			return
		}

		if n > 0 {
			// Сохраняем аудио если нужно
			if audioWriter != nil {
				audioWriter.Write(buffer[:n])
			}

			// Отправляем в AssemblyAI
			t.sendAudioData(buffer[:n])
		}
	}
}

func (t *Transcriber) sendAudioData(data []byte) {
	if t.conn == nil {
		return
	}

	audioMsg := AudioData{
		AudioData: base64.StdEncoding.EncodeToString(data),
	}

	if err := t.conn.WriteJSON(audioMsg); err != nil {
		log.Printf("Failed to send audio data: %v", err)
	}
}

func (t *Transcriber) Stop() {
	t.mu.Lock()
	t.isRunning = false
	t.mu.Unlock()

	// Останавливаем аудио команду
	if t.audioCmd != nil && t.audioCmd.Process != nil {
		t.audioCmd.Process.Kill()
	}

	// Закрываем WebSocket
	if t.conn != nil {
		// Отправляем сигнал завершения
		terminateMsg := map[string]string{"terminate_session": "true"}
		t.conn.WriteJSON(terminateMsg)
		t.conn.Close()
	}
}

func (t *Transcriber) SaveTranscripts() error {
	if len(t.transcripts) == 0 {
		return nil
	}

	filename := fmt.Sprintf("%s/transcript_%s_%s.txt",
		t.config.OutputDir,
		t.config.TwitchChannel,
		time.Now().Format("20060102_150405"))

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	for _, transcript := range t.transcripts {
		writer.WriteString(transcript + "\n")
	}
	writer.Flush()

	fmt.Printf("📄 Transcript saved to: %s\n", filename)
	return nil
}

func HolyFuck() {
	// Проверка наличия необходимых программ
	if err := checkDependencies(); err != nil {
		log.Fatal(err)
	}

	// Конфигурация
	config := Config{
		AssemblyAIKey: os.Getenv("ASSEMBLY_TOKEN"),
		TwitchChannel: "lagoda1337", // Замените на нужный канал
		SaveAudio:     true,
		OutputDir:     "./output",
	}

	if config.AssemblyAIKey == "" {
		log.Fatal("ASSEMBLYAI_API_KEY environment variable is required")
	}

	// Создаем директорию для выходных файлов
	os.MkdirAll(config.OutputDir, os.ModePerm)

	// Создаем транскрайбер
	transcriber := NewTranscriber(config)

	// Обработка сигналов
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Printf("🎬 Starting transcription for channel: %s\n", config.TwitchChannel)
	fmt.Println("Press Ctrl+C to stop...")

	// Запускаем транскрибер
	if err := transcriber.Start(); err != nil {
		log.Fatal(err)
	}

	// Ожидаем сигнал остановки
	<-sigChan
	fmt.Println("\n🛑 Stopping transcriber...")

	// Останавливаем транскрибер
	transcriber.Stop()

	// Сохраняем транскрипты
	if err := transcriber.SaveTranscripts(); err != nil {
		log.Printf("Failed to save transcripts: %v", err)
	}

	fmt.Println("✅ Done!")
}

func checkDependencies() error {
	// Проверяем streamlink
	if _, err := exec.LookPath("streamlink"); err != nil {
		return fmt.Errorf("streamlink not found. Install it: pip install streamlink")
	}

	// Проверяем ffmpeg
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return fmt.Errorf("ffmpeg not found. Install it from https://ffmpeg.org/")
	}

	return nil
}
