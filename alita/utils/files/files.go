package files

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/divideprojects/Alita_Robot/alita/utils/security"
)

// FileManager provides safe file operations with proper resource management
type FileManager struct {
	tempDir string
}

// NewFileManager creates a new file manager
func NewFileManager(tempDir string) *FileManager {
	if tempDir == "" {
		tempDir = os.TempDir()
	}
	return &FileManager{
		tempDir: tempDir,
	}
}

// SafeOpenFile opens a file with proper error handling and resource management
func (fm *FileManager) SafeOpenFile(filename string) (*os.File, error) {
	// Validate filename
	if err := security.ValidateFileUpload(filepath.Base(filename), 0); err != nil {
		return nil, fmt.Errorf("invalid filename: %w", err)
	}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return file, nil
}

// SafeCreateFile creates a file with proper permissions and error handling
func (fm *FileManager) SafeCreateFile(filename string) (*os.File, error) {
	// Validate filename
	if err := security.ValidateFileUpload(filepath.Base(filename), 0); err != nil {
		return nil, fmt.Errorf("invalid filename: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}

	return file, nil
}

// SafeWriteFile writes data to a file with proper resource management
func (fm *FileManager) SafeWriteFile(filename string, data []byte) error {
	file, err := fm.SafeCreateFile(filename)
	if err != nil {
		return err
	}
	defer SafeCloseFile(file)

	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// SafeReadFile reads a file with size limits and proper resource management
func (fm *FileManager) SafeReadFile(filename string, maxSize int64) ([]byte, error) {
	file, err := fm.SafeOpenFile(filename)
	if err != nil {
		return nil, err
	}
	defer SafeCloseFile(file)

	// Check file size
	stat, err := file.Stat()
	if err != nil {
		return nil, fmt.Errorf("failed to get file stats: %w", err)
	}

	if stat.Size() > maxSize {
		return nil, fmt.Errorf("file too large: %d bytes (max: %d)", stat.Size(), maxSize)
	}

	// Use LimitReader to prevent reading too much data
	limitedReader := io.LimitReader(file, maxSize)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// SafeCopyFile copies a file with proper resource management
func (fm *FileManager) SafeCopyFile(src, dst string) error {
	srcFile, err := fm.SafeOpenFile(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer SafeCloseFile(srcFile)

	dstFile, err := fm.SafeCreateFile(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer SafeCloseFile(dstFile)

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy file: %w", err)
	}

	return nil
}

// CreateTempFile creates a temporary file with proper cleanup
func (fm *FileManager) CreateTempFile(pattern string) (*TempFile, error) {
	file, err := os.CreateTemp(fm.tempDir, pattern)
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	return &TempFile{
		File: file,
		path: file.Name(),
	}, nil
}

// TempFile represents a temporary file with automatic cleanup
type TempFile struct {
	*os.File
	path string
}

// Close closes the temporary file and removes it
func (tf *TempFile) Close() error {
	var err error

	// Close the file first
	if tf.File != nil {
		err = tf.File.Close()
	}

	// Remove the file
	if removeErr := os.Remove(tf.path); removeErr != nil {
		log.WithError(removeErr).Errorf("Failed to remove temp file: %s", tf.path)
		if err == nil {
			err = removeErr
		}
	}

	return err
}

// Path returns the path of the temporary file
func (tf *TempFile) Path() string {
	return tf.path
}

// SafeCloseFile safely closes a file and logs any errors
func SafeCloseFile(file *os.File) {
	if file != nil {
		if err := file.Close(); err != nil {
			log.WithError(err).Error("Failed to close file")
		}
	}
}

// SafeRemoveFile safely removes a file and logs any errors
func SafeRemoveFile(filename string) {
	if err := os.Remove(filename); err != nil {
		log.WithError(err).Errorf("Failed to remove file: %s", filename)
	}
}

// CleanupOldFiles removes files older than the specified duration
func (fm *FileManager) CleanupOldFiles(dir string, maxAge time.Duration) error {
	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && time.Since(info.ModTime()) > maxAge {
			if removeErr := os.Remove(path); removeErr != nil {
				log.WithError(removeErr).Errorf("Failed to remove old file: %s", path)
			} else {
				log.Debugf("Removed old file: %s", path)
			}
		}

		return nil
	})
}

// WriteFileWithContext writes a file with context cancellation support
func (fm *FileManager) WriteFileWithContext(ctx context.Context, filename string, data []byte) error {
	// Check if context is already cancelled
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	file, err := fm.SafeCreateFile(filename)
	if err != nil {
		return err
	}
	defer SafeCloseFile(file)

	// Write in chunks to allow for context cancellation
	const chunkSize = 8192
	for i := 0; i < len(data); i += chunkSize {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		end := i + chunkSize
		if end > len(data) {
			end = len(data)
		}

		_, err = file.Write(data[i:end])
		if err != nil {
			return fmt.Errorf("failed to write chunk: %w", err)
		}
	}

	return nil
}

// GetFileSize returns the size of a file
func GetFileSize(filename string) (int64, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return 0, err
	}
	return stat.Size(), nil
}

// FileExists checks if a file exists
func FileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}
