package api

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	ctx "github.com/gophish/gophish/context"
	log "github.com/gophish/gophish/logger"
	"github.com/gophish/gophish/models"
	"github.com/jinzhu/gorm"
)

const (
	pageUploadMemoryLimit = 64 << 20
	pageUploadMaxFiles    = 500
	pageUploadMaxFileSize = 100 << 20
	pageUploadMaxHTML     = 20 << 20
	pageAssetsURLPrefix   = "/static/page-assets"
	pageAssetsRoot        = "page-assets"
)

var pageUploadAllowedExtensions = map[string]bool{
	".html": true, ".htm": true,
	".css": true, ".js": true,
	".png": true, ".jpg": true, ".jpeg": true,
	".gif": true, ".webp": true, ".svg": true,
	".ico": true, ".woff": true, ".woff2": true,
	".ttf": true, ".eot": true, ".otf": true,
	".map": true, ".json": true,
}

var pageResourceReferencePattern = regexp.MustCompile(`(?i)(?:\b(?:src|href|poster)\s*=\s*|\burl\s*\()(["'])?([^"'\s>]+)`)

// UploadPageFolder imports a landing page from a browser-selected directory.
// The browser's directory picker sends each file with its relative path in the
// part filename. Go's multipart parser keeps only the base name in Part.FileName,
// so the frontend sends the original relative path in a companion paths field.
func (as *Server) UploadPageFolder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		JSONResponse(w, models.Response{Success: false, Message: "Method not allowed"}, http.StatusMethodNotAllowed)
		return
	}
	userID, ok := ctx.Get(r, "user_id").(int64)
	if !ok {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid user context"}, http.StatusUnauthorized)
		return
	}
	if err := r.ParseMultipartForm(pageUploadMemoryLimit); err != nil {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid multipart upload"}, http.StatusBadRequest)
		return
	}
	defer r.MultipartForm.RemoveAll()

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		JSONResponse(w, models.Response{Success: false, Message: "No page name specified"}, http.StatusBadRequest)
		return
	}
	if len(r.MultipartForm.File) != 1 || len(r.MultipartForm.File["files"]) == 0 {
		JSONResponse(w, models.Response{Success: false, Message: "No folder files uploaded"}, http.StatusBadRequest)
		return
	}
	if len(r.MultipartForm.File["files"]) > pageUploadMaxFiles {
		JSONResponse(w, models.Response{Success: false, Message: "Too many files uploaded"}, http.StatusBadRequest)
		return
	}
	if _, err := models.GetPageByName(name, userID); err != gorm.ErrRecordNotFound {
		JSONResponse(w, models.Response{Success: false, Message: "Page name already in use"}, http.StatusConflict)
		return
	}

	paths := r.Form["paths"]
	files := r.MultipartForm.File["files"]
	if len(paths) != len(files) {
		JSONResponse(w, models.Response{Success: false, Message: "Invalid file path metadata"}, http.StatusBadRequest)
		return
	}
	for _, fileHeader := range files {
		if fileHeader.Size > pageUploadMaxFileSize {
			JSONResponse(w, models.Response{Success: false, Message: "A file exceeds the maximum size"}, http.StatusBadRequest)
			return
		}
	}

	normalized := make([]normalizedPageFile, 0, len(files))
	for i, fileHeader := range files {
		normalizedFile, err := normalizeUploadPath(fileHeader, paths[i])
		if err != nil {
			JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
			return
		}
		if normalizedFile == nil {
			continue
		}
		for _, existing := range normalized {
			if strings.EqualFold(existing.relativePath, normalizedFile.relativePath) {
				JSONResponse(w, models.Response{Success: false, Message: "Duplicate file path uploaded"}, http.StatusBadRequest)
				return
			}
		}
		normalized = append(normalized, *normalizedFile)
	}
	entryHTML, err := selectPageHTML(normalized)
	if err != nil {
		JSONResponse(w, models.Response{Success: false, Message: err.Error()}, http.StatusBadRequest)
		return
	}

	pageDir := fmt.Sprintf("page-%d-%d", userID, time.Now().UTC().UnixNano())
	uploaded, err := persistPageFiles(pageDir, normalized, entryHTML.relativePath)
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Error storing uploaded files"}, http.StatusInternalServerError)
		return
	}

	pageHTML, err := io.ReadAll(uploaded[entryHTML.relativePath])
	if err != nil {
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Error reading uploaded HTML"}, http.StatusInternalServerError)
		return
	}
	if len(pageHTML) > pageUploadMaxHTML {
		JSONResponse(w, models.Response{Success: false, Message: "HTML file is too large"}, http.StatusBadRequest)
		return
	}
	rewrittenHTML := rewritePageAssetReferences(string(pageHTML), pageDir, entryHTML.relativePath)
	page := models.Page{
		Name:         name,
		HTML:         rewrittenHTML,
		UserId:       userID,
		ModifiedDate: time.Now().UTC(),
	}
	closeUploadedPageFiles(uploaded)
	if err := models.PostPage(&page); err != nil {
		removePageAssets(pageDir)
		log.Error(err)
		JSONResponse(w, models.Response{Success: false, Message: "Error creating page: " + err.Error()}, http.StatusInternalServerError)
		return
	}
	JSONResponse(w, page, http.StatusCreated)
}

func closeUploadedPageFiles(files map[string]io.ReadSeekCloser) {
	for _, file := range files {
		if err := file.Close(); err != nil {
			log.Error(err)
		}
	}
}

type normalizedPageFile struct {
	header       *multipart.FileHeader
	relativePath string
}

func normalizeUploadPath(fileHeader *multipart.FileHeader, browserPath string) (*normalizedPageFile, error) {
	cleanPath := browserPath
	if cleanPath == "" {
		cleanPath = fileHeader.Filename
	}
	cleanPath = strings.ReplaceAll(cleanPath, "\\", "/")
	cleanPath = strings.TrimPrefix(cleanPath, "/")
	if cleanPath == "" || strings.HasPrefix(cleanPath, ".") || strings.Contains(cleanPath, "/.") {
		return nil, nil
	}
	cleanPath = path.Clean(cleanPath)
	if cleanPath == "." || cleanPath == ".." || strings.HasPrefix(cleanPath, "../") {
		return nil, errors.New("Invalid file path")
	}
	for _, segment := range strings.Split(cleanPath, "/") {
		if segment == "" || segment == "." || segment == ".." {
			return nil, errors.New("Invalid file path")
		}
	}
	if !pageUploadAllowedExtensions[strings.ToLower(path.Ext(cleanPath))] {
		return nil, nil
	}
	return &normalizedPageFile{header: fileHeader, relativePath: cleanPath}, nil
}

func selectPageHTML(files []normalizedPageFile) (*normalizedPageFile, error) {
	sort.Slice(files, func(i, j int) bool {
		return strings.ToLower(files[i].relativePath) < strings.ToLower(files[j].relativePath)
	})
	for i := range files {
		if strings.EqualFold(path.Base(files[i].relativePath), "index.html") {
			return &files[i], nil
		}
	}
	for i := range files {
		if strings.EqualFold(path.Ext(files[i].relativePath), ".htm") {
			return &files[i], nil
		}
	}
	return nil, errors.New("No HTML file found in uploaded folder")
}

func persistPageFiles(pageDir string, files []normalizedPageFile, entryHTML string) (map[string]io.ReadSeekCloser, error) {
	if err := os.MkdirAll(filepath.Join("static", "endpoint", pageAssetsRoot, pageDir), 0750); err != nil {
		return nil, err
	}
	uploaded := make(map[string]io.ReadSeekCloser)
	for _, file := range files {
		if file.relativePath == entryHTML {
			src, err := file.header.Open()
			if err != nil {
				removePageAssets(pageDir)
				return nil, err
			}
			uploaded[file.relativePath] = src
			continue
		}
		src, err := file.header.Open()
		if err != nil {
			removePageAssets(pageDir)
			return nil, err
		}
		destinationPath := filepath.Join("static", "endpoint", pageAssetsRoot, pageDir, filepath.FromSlash(file.relativePath))
		if err := os.MkdirAll(filepath.Dir(destinationPath), 0750); err != nil {
			src.Close()
			removePageAssets(pageDir)
			return nil, err
		}
		destination, err := os.OpenFile(destinationPath, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0640)
		if err != nil {
			src.Close()
			removePageAssets(pageDir)
			return nil, err
		}
		_, copyErr := io.Copy(destination, src)
		closeErr := destination.Close()
		src.Close()
		if copyErr != nil {
			removePageAssets(pageDir)
			return nil, copyErr
		}
		if closeErr != nil {
			removePageAssets(pageDir)
			return nil, closeErr
		}
	}
	return uploaded, nil
}

func removePageAssets(pageDir string) {
	if err := os.RemoveAll(filepath.Join("static", "endpoint", pageAssetsRoot, pageDir)); err != nil {
		log.Error(err)
	}
}

func rewritePageAssetReferences(html string, pageDir string, entryPath string) string {
	entryDir := path.Dir(entryPath)
	if entryDir == "." {
		entryDir = ""
	}
	return pageResourceReferencePattern.ReplaceAllStringFunc(html, func(match string) string {
		submatches := pageResourceReferencePattern.FindStringSubmatch(match)
		if len(submatches) < 3 {
			return match
		}
		quote := submatches[1]
		rawValue := submatches[2]
		unquoted, err := url.PathUnescape(rawValue)
		if err != nil {
			return match
		}
		if !isRewritablePageReference(unquoted) {
			return match
		}
		assetPath := unquoted
		if entryDir != "" {
			assetPath = path.Join(entryDir, unquoted)
		}
		if assetPath == "" || assetPath == "." || assetPath == ".." || strings.HasPrefix(assetPath, "../") {
			return match
		}
		assetPath = path.Clean(assetPath)
		rewritten := fmt.Sprintf("%s/%s/%s", pageAssetsURLPrefix, pageDir, assetPath)
		if quote == "" {
			rewritten = fmt.Sprintf("%s'%s'", quote, rewritten)
		}
		return strings.Replace(match, rawValue, rewritten, 1)
	})
}

func isRewritablePageReference(value string) bool {
	if value == "" || strings.HasPrefix(value, "#") {
		return false
	}
	if strings.HasPrefix(value, "http://") || strings.HasPrefix(value, "https://") ||
		strings.HasPrefix(value, "//") || strings.HasPrefix(value, "data:") ||
		strings.HasPrefix(value, "mailto:") || strings.HasPrefix(value, "tel:") ||
		strings.HasPrefix(value, "javascript:") || strings.HasPrefix(value, "{{") {
		return false
	}
	return true
}
