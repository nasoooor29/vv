package services

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"visory/internal/utils"

	"github.com/labstack/echo/v4"
)

type ISOService struct {
	Dispatcher *utils.Dispatcher
	Logger     *slog.Logger
	FS         *utils.FS
}

// NewISOService creates a new ISOService with dependency injection
func NewISOService(dispatcher *utils.Dispatcher, fs *utils.FS, logger *slog.Logger) *ISOService {
	return &ISOService{
		Dispatcher: dispatcher.WithGroup("iso"),
		Logger:     logger.WithGroup("iso"),
		FS:         fs,
	}
}

// @Summary      List ISO files
// @Description  Get a list of all available ISO files
// @Tags         iso
// @Produce      json
// @Success      200  {array}   map[string]interface{}
// @Failure      401  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPError
// @Router       /iso [get]
//
// ListISOs returns list of available ISO files
func (s *ISOService) ListISOs(c echo.Context) error {
	entries, err := os.ReadDir(s.FS.ISOs)
	if err != nil {
		s.Logger.Error("Failed to list ISO directory", "error", err, "path", s.FS.ISOs)
		return s.Dispatcher.NewInternalServerError("Failed to list ISO files", err)
	}

	isos := make([]map[string]interface{}, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			info, err := entry.Info()
			if err != nil {
				s.Logger.Warn("Failed to get file info", "name", entry.Name(), "error", err)
				continue
			}

			isos = append(isos, map[string]interface{}{
				"name":         entry.Name(),
				"size":         info.Size(),
				"modified":     info.ModTime(),
				"is_directory": entry.IsDir(),
			})
		}
	}

	return c.JSON(http.StatusOK, isos)
}

// @Summary      Get ISO file info
// @Description  Get detailed information about a specific ISO file
// @Tags         iso
// @Param        filename  path  string  true  "ISO filename"
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      404  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPError
// @Router       /iso/{filename} [get]
//
// GetISOInfo returns information about a specific ISO file
func (s *ISOService) GetISOInfo(c echo.Context) error {
	filename := c.Param("filename")
	if filename == "" {
		return s.Dispatcher.NewBadRequest("Filename is required", nil)
	}

	filePath := filepath.Join(s.FS.ISOs, filename)

	// Validate path to prevent directory traversal
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return s.Dispatcher.NewBadRequest("Invalid filename", err)
	}

	absISODir, err := filepath.Abs(s.FS.ISOs)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to validate path", err)
	}

	if !filepath.HasPrefix(absPath, absISODir) {
		return s.Dispatcher.NewBadRequest("Invalid filename", nil)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.Dispatcher.NewNotFound("ISO file not found", err)
		}
		s.Logger.Error("Failed to get file info", "filename", filename, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to get ISO file info", err)
	}

	if info.IsDir() {
		return s.Dispatcher.NewBadRequest("Path is a directory, not a file", nil)
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"name":         info.Name(),
		"size":         info.Size(),
		"modified":     info.ModTime(),
		"is_directory": info.IsDir(),
	})
}

// @Summary      Upload ISO file
// @Description  Upload a new ISO file to the server
// @Tags         iso
// @Accept       multipart/form-data
// @Param        file  formData  file  true  "ISO file to upload"
// @Produce      json
// @Success      201  {object}  map[string]interface{}
// @Failure      400  {object}  models.HTTPError
// @Failure      401  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPError
// @Router       /iso [post]
//
// UploadISO handles ISO file uploads
func (s *ISOService) UploadISO(c echo.Context) error {
	file, err := c.FormFile("file")
	if err != nil {
		return s.Dispatcher.NewBadRequest("File upload failed", err)
	}

	// Validate file size (max 5GB)
	const maxSize = 5 * 1024 * 1024 * 1024 // 5GB
	if file.Size > maxSize {
		return s.Dispatcher.NewBadRequest("File size exceeds maximum limit of 5GB", nil)
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		s.Logger.Error("Failed to open uploaded file", "filename", file.Filename, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to process uploaded file", err)
	}
	defer src.Close()

	// Create destination file
	dstPath := filepath.Join(s.FS.ISOs, filepath.Base(file.Filename))

	// Validate destination path to prevent directory traversal
	absDstPath, err := filepath.Abs(dstPath)
	if err != nil {
		return s.Dispatcher.NewBadRequest("Invalid filename", err)
	}

	absISODir, err := filepath.Abs(s.FS.ISOs)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to validate path", err)
	}

	if !filepath.HasPrefix(absDstPath, absISODir) {
		return s.Dispatcher.NewBadRequest("Invalid filename", nil)
	}

	dst, err := os.Create(dstPath)
	if err != nil {
		s.Logger.Error("Failed to create destination file", "filename", file.Filename, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to save ISO file", err)
	}
	defer dst.Close()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		s.Logger.Error("Failed to copy file content", "filename", file.Filename, "error", err)
		os.Remove(dstPath) // Clean up partial file
		return s.Dispatcher.NewInternalServerError("Failed to save ISO file", err)
	}

	info, err := os.Stat(dstPath)
	if err != nil {
		s.Logger.Error("Failed to get uploaded file info", "filename", file.Filename, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to verify ISO file", err)
	}

	s.Logger.Info("ISO file uploaded successfully", "filename", file.Filename, "size", info.Size())

	return c.JSON(http.StatusCreated, map[string]interface{}{
		"name":     info.Name(),
		"size":     info.Size(),
		"modified": info.ModTime(),
		"message":  fmt.Sprintf("ISO file '%s' uploaded successfully", file.Filename),
	})
}

// @Summary      Delete ISO file
// @Description  Delete an ISO file from the server
// @Tags         iso
// @Param        filename  path  string  true  "ISO filename to delete"
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      404  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPError
// @Router       /iso/{filename} [delete]
//
// DeleteISO deletes an ISO file
func (s *ISOService) DeleteISO(c echo.Context) error {
	filename := c.Param("filename")
	if filename == "" {
		return s.Dispatcher.NewBadRequest("Filename is required", nil)
	}

	filePath := filepath.Join(s.FS.ISOs, filename)

	// Validate path to prevent directory traversal
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return s.Dispatcher.NewBadRequest("Invalid filename", err)
	}

	absISODir, err := filepath.Abs(s.FS.ISOs)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to validate path", err)
	}

	if !filepath.HasPrefix(absPath, absISODir) {
		return s.Dispatcher.NewBadRequest("Invalid filename", nil)
	}

	// Check if file exists and is a file (not directory)
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.Dispatcher.NewNotFound("ISO file not found", err)
		}
		s.Logger.Error("Failed to check file", "filename", filename, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to delete ISO file", err)
	}

	if info.IsDir() {
		return s.Dispatcher.NewBadRequest("Cannot delete a directory", nil)
	}

	if err := os.Remove(filePath); err != nil {
		s.Logger.Error("Failed to delete file", "filename", filename, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to delete ISO file", err)
	}

	s.Logger.Info("ISO file deleted successfully", "filename", filename)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"message": fmt.Sprintf("ISO file '%s' deleted successfully", filename),
	})
}

// @Summary      Download ISO file
// @Description  Download an ISO file
// @Tags         iso
// @Param        filename  path  string  true  "ISO filename to download"
// @Produce      octet-stream
// @Success      200
// @Failure      401  {object}  models.HTTPError
// @Failure      403  {object}  models.HTTPError
// @Failure      404  {object}  models.HTTPError
// @Failure      500  {object}  models.HTTPError
// @Router       /iso/{filename}/download [get]
//
// DownloadISO downloads an ISO file
func (s *ISOService) DownloadISO(c echo.Context) error {
	filename := c.Param("filename")
	if filename == "" {
		return s.Dispatcher.NewBadRequest("Filename is required", nil)
	}

	filePath := filepath.Join(s.FS.ISOs, filename)

	// Validate path to prevent directory traversal
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return s.Dispatcher.NewBadRequest("Invalid filename", err)
	}

	absISODir, err := filepath.Abs(s.FS.ISOs)
	if err != nil {
		return s.Dispatcher.NewInternalServerError("Failed to validate path", err)
	}

	if !filepath.HasPrefix(absPath, absISODir) {
		return s.Dispatcher.NewBadRequest("Invalid filename", nil)
	}

	// Check if file exists and is a file (not directory)
	info, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return s.Dispatcher.NewNotFound("ISO file not found", err)
		}
		s.Logger.Error("Failed to check file", "filename", filename, "error", err)
		return s.Dispatcher.NewInternalServerError("Failed to download ISO file", err)
	}

	if info.IsDir() {
		return s.Dispatcher.NewBadRequest("Cannot download a directory", nil)
	}

	s.Logger.Info("ISO file download started", "filename", filename)

	return c.File(filePath)
}
