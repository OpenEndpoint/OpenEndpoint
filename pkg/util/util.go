package util

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// GenerateID generates a unique ID
func GenerateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// GenerateETag generates an ETag
func GenerateETag() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("\"%s\"", hex.EncodeToString(b))
}

// GenerateUploadID generates a multipart upload ID
func GenerateUploadID() string {
	return fmt.Sprintf("%d-%s", time.Now().UnixNano(), GenerateID())
}

// CopyReader copies data from reader to writer with progress
func CopyReader(dst io.Writer, src io.Reader, progress func(n int64)) (written int64, err error) {
	buf := make([]byte, 32*1024) // 32KB buffer
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				if progress != nil {
					progress(written)
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}

// RoundToBlock rounds size to nearest block size
func RoundToBlock(size int64, blockSize int64) int64 {
	if size%blockSize == 0 {
		return size
	}
	return ((size / blockSize) + 1) * blockSize
}

// FormatBytes formats bytes as human readable string
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration formats duration as human readable string
func FormatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%.1fm", d.Minutes())
	}
	return fmt.Sprintf("%.1fh", d.Hours())
}

// Contains checks if slice contains value
func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}

// Unique removes duplicates from slice
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]bool)
	result := make([]T, 0)
	for _, v := range slice {
		if !seen[v] {
			seen[v] = true
			result = append(result, v)
		}
	}
	return result
}

// ChunkSlice slices a slice into chunks
func ChunkSlice[T any](slice []T, chunkSize int) [][]T {
	var chunks [][]T
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Clamp clamps value between min and max
func Clamp(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
