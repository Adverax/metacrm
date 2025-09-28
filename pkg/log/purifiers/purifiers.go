package purifiers

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"sync"
	"time"
)

type Purifier interface {
	Purify(original, derivative string) string
}

type ChunkStorage interface {
	Save(data string) string
}

type fileInfo struct {
	path    string
	modTime time.Time
}

type ChunkManagerOptions struct {
	MaxCount int
	MaxAge   int
	Folder   string
}

type ChunkManager struct {
	options   ChunkManagerOptions
	files     []*fileInfo
	mu        sync.Mutex
	cleanupCh chan struct{}
}

func NewChunkManager(options ChunkManagerOptions) *ChunkManager {
	_ = os.MkdirAll(options.Folder, 0755)

	cb := &ChunkManager{
		options:   options,
		cleanupCh: make(chan struct{}, 1),
	}

	go cb.cleanupWorker()

	return cb
}

func (that *ChunkManager) Save(data string) string {
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), that.getExtension(data))
	filePath := filepath.Join(that.options.Folder, filename)
	err := os.WriteFile(filePath, []byte(data), 0644)
	if err != nil {
		return err.Error()
	}
	defer os.Remove(filePath)

	compressedFilePath, err := that.compressFile(filePath)
	if err != nil {
		return err.Error()
	}

	that.addFile(&fileInfo{path: compressedFilePath, modTime: time.Now()})

	select {
	case that.cleanupCh <- struct{}{}:
	default:
	}

	return fmt.Sprintf("CHUNK: %s", filepath.Base(compressedFilePath))
}

func (that *ChunkManager) compressFile(filePath string) (string, error) {
	compressedFilePath := filePath + ".gz"
	inFile, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer inFile.Close()

	outFile, err := os.Create(compressedFilePath)
	if err != nil {
		return "", err
	}
	defer outFile.Close()

	gzipWriter := gzip.NewWriter(outFile)
	defer gzipWriter.Close()

	_, err = io.Copy(gzipWriter, inFile)
	if err != nil {
		return "", err
	}

	return compressedFilePath, nil
}

func (that *ChunkManager) addFile(file *fileInfo) {
	that.mu.Lock()
	defer that.mu.Unlock()

	that.files = append(that.files, file)
}

func (that *ChunkManager) cleanupWorker() {
	for range that.cleanupCh {
		that.cleanupOldFiles()
	}
}

func (that *ChunkManager) cleanupOldFiles() {
	that.mu.Lock()
	defer that.mu.Unlock()

	// Sort files by modification time
	sort.Slice(that.files, func(i, j int) bool {
		return that.files[i].modTime.Before(that.files[j].modTime)
	})

	// Remove oldest files if count exceeds maxCount
	if len(that.files) > that.options.MaxCount {
		for _, file := range that.files[:len(that.files)-that.options.MaxCount] {
			os.Remove(file.path)
		}
		that.files = that.files[len(that.files)-that.options.MaxCount:]
	}

	// Remove files older than maxAge
	cutoff := time.Now().Add(-time.Duration(that.options.MaxAge) * time.Second)
	for i := 0; i < len(that.files); {
		if that.files[i].modTime.Before(cutoff) {
			os.Remove(that.files[i].path)
			that.files = append(that.files[:i], that.files[i+1:]...)
		} else {
			i++
		}
	}
}

func (that *ChunkManager) getExtension(s string) string {
	if isPNG(s) {
		return ".png"
	}

	if isJSON(s) {
		return ".json"
	}

	return ".txt"
}

func isPNG(s string) bool {
	return len(s) >= 4 && s[0] == 0x89 && s[1] == 0x50 && s[2] == 0x4E && s[3] == 0x47
}

var canBeJson = regexp.MustCompile(`^\s*[\[{"0-9]`)

func isJSON(s string) bool {
	if !canBeJson.MatchString(s) {
		return false
	}

	decoder := json.NewDecoder(bytes.NewReader([]byte(s)))
	for {
		_, err := decoder.Token()
		if err != nil {
			return err == nil
		}
	}
}
