package fsdata

import (
	"fmt"
	"path/filepath"
	"worker/internal/db"

	"github.com/google/uuid"
)

// how problems and submissions are laid out on disk under the data/ directory.

const (
	ProblemsSubdir    = "problems"
	SubmissionsSubdir = "submissions"

	ProblemMetaFile    = "meta.json"
	SubmissionMetaFile = "meta.json"
)

// data/problems
func ProblemsDir(dataDir string) string {
	return filepath.Join(dataDir, ProblemsSubdir)
}

// data/submissions
func SubmissionsDir(dataDir string) string {
	return filepath.Join(dataDir, SubmissionsSubdir)
}

// data/problems/<id>
func ProblemDir(dataDir, problemID string) string {
	return filepath.Join(ProblemsDir(dataDir), problemID)
}

// data/problems/<id>/meta.json
func ProblemMetaPath(dataDir, problemID string) string {
	return filepath.Join(ProblemDir(dataDir, problemID), ProblemMetaFile)
}

// data/problems/<id>/testcases
func TestcasesDir(dataDir, problemID string) string {
	return filepath.Join(ProblemDir(dataDir, problemID), "testcases")
}

// data/submissions/<id>
func SubmissionDir(dataDir string, id uuid.UUID) string {
	return filepath.Join(SubmissionsDir(dataDir), id.String())
}

// data/submissions/<id>/meta.json
func SubmissionMetaPath(dataDir string, id uuid.UUID) string {
	return filepath.Join(SubmissionDir(dataDir, id), SubmissionMetaFile)
}

// data/submissions/<id>/<source filename for lang>
func SubmissionSourcePath(dataDir string, id uuid.UUID, lang db.Lang) (string, error) {
	filename, err := SourceFilename(lang)
	if err != nil {
		return "", err
	}
	return filepath.Join(SubmissionDir(dataDir, id), filename), nil
}

// matches the chk_language constraint in the submissions table
func SourceFilename(lang db.Lang) (string, error) {
	switch lang {
	case "C++":
		return "main.cpp", nil
	case "C":
		return "main.c", nil
	case "Java":
		return "Main.java", nil
	case "Python":
		return "main.py", nil
	default:
		return "", fmt.Errorf("unrecognized language %q", lang)
	}
}