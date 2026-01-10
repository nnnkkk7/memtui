package viewer

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"io"

	"github.com/klauspost/compress/zstd"
)

// IsGzipCompressed checks if the data starts with gzip magic bytes (0x1f, 0x8b)
func IsGzipCompressed(data []byte) bool {
	return len(data) >= 2 && data[0] == 0x1f && data[1] == 0x8b
}

// IsZlibCompressed checks if the data starts with zlib magic bytes (0x78)
// followed by a valid compression level byte (0x01, 0x5e, 0x9c, or 0xda)
func IsZlibCompressed(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	// Zlib starts with 0x78 followed by compression level byte
	// 0x01 = no compression
	// 0x5e = best speed (level 1-5)
	// 0x9c = default compression (level 6)
	// 0xda = best compression (level 7-9)
	return data[0] == 0x78 && (data[1] == 0x01 || data[1] == 0x5e || data[1] == 0x9c || data[1] == 0xda)
}

// IsZstdCompressed checks if the data starts with zstd magic bytes (0x28, 0xb5, 0x2f, 0xfd)
func IsZstdCompressed(data []byte) bool {
	return len(data) >= 4 && data[0] == 0x28 && data[1] == 0xb5 && data[2] == 0x2f && data[3] == 0xfd
}

// Decompress automatically detects compression format and decompresses the data.
// Supports gzip, zlib, and zstd compression formats.
// If the data is not compressed, it returns the original data unchanged.
func Decompress(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return data, nil
	}

	// Check for gzip compression first
	if IsGzipCompressed(data) {
		return decompressGzip(data)
	}

	// Check for zlib compression
	if IsZlibCompressed(data) {
		return decompressZlib(data)
	}

	// Check for zstd compression
	if IsZstdCompressed(data) {
		return decompressZstd(data)
	}

	// Not compressed, return original data
	return data, nil
}

// decompressGzip decompresses gzip-compressed data
func decompressGzip(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

// decompressZlib decompresses zlib-compressed data
func decompressZlib(data []byte) ([]byte, error) {
	reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer func() { _ = reader.Close() }()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

// decompressZstd decompresses zstd-compressed data
func decompressZstd(data []byte) ([]byte, error) {
	reader, err := zstd.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}
