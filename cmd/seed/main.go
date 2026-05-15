// cmd/seed seeds the database with a small fixture set so a fresh checkout can
// be demoed end-to-end: one language (Python), a few problems (with on-disk
// input/expected files), tags, and one admin user.
//
// The seeder is idempotent: re-running it inserts only what's missing and
// re-writes the on-disk test-case files so paths stay valid after a `make clean`.
//
// Problem wrapper code: each problem's wrapper_code MUST write its final
// answer to fd 3 and close fd 3, because the worker invokes the program as
//
//	<run_cmd> <<< "<stdin line>" 3>"<actual_output_file>"
//
// i.e. bash redirects fd 3 to a file before exec'ing the program. Closing fd 3
// after writing flushes the buffer and finalizes the output file.
package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/NemCaBong/executify/internal/adapter/repository/entity"
	"github.com/NemCaBong/executify/internal/config"
)

const (
	adminEmail    = "admin@executify.local"
	adminUsername = "admin"
	adminPassword = "admin12345" // dev-only; rotate before any deploy
)

// pythonWrapperCode is the harness injected around the user's solve() function.
// {{.}} is replaced with the user-submitted source at submission time
// (see internal/application/submission/usecase.go).
//
// Each test case is fed as ONE line on stdin. The user's solve(line) should
// return a string-convertible answer. The wrapper writes it to fd 3 and closes
// fd 3 so the runner can read the actual output file cleanly.
const pythonWrapperCode = `import os
import sys

{{.}}

if __name__ == "__main__":
    line = sys.stdin.readline().rstrip("\n")
    result = solve(line)
    os.write(3, str(result).encode())
    os.close(3)
`

func main() {
	cfg := config.Load()
	db := config.NewMySQLConnection(cfg)

	dataRoot, err := filepath.Abs("seed/data/problems")
	if err != nil {
		log.Fatalf("resolve data root: %v", err)
	}
	if err := os.MkdirAll(dataRoot, 0o755); err != nil {
		log.Fatalf("create data root: %v", err)
	}

	pyLangID := seedLanguage(db)
	tagIDs := seedTags(db, []entity.Tag{
		{Name: "Math", Slug: "math"},
		{Name: "String", Slug: "string"},
		{Name: "Easy Warmup", Slug: "easy-warmup"},
	})
	seedProblems(db, dataRoot, tagIDs)
	seedAdmin(db)

	log.Printf("seed complete — python_lang_id=%d", pyLangID)
}

func seedLanguage(db *gorm.DB) int {
	lang := entity.Language{
		Name:       "Python 3",
		CompileCmd: nil,
		RunCmd:     "python3 main.py",
		SourceFile: "main.py",
	}
	if err := db.Where("name = ?", lang.Name).
		Attrs(lang).
		FirstOrCreate(&lang).Error; err != nil {
		log.Fatalf("seed language: %v", err)
	}
	return lang.ID
}

func seedTags(db *gorm.DB, tags []entity.Tag) map[string]int {
	out := make(map[string]int, len(tags))
	for _, t := range tags {
		row := t
		if err := db.Where("slug = ?", t.Slug).
			Attrs(t).
			FirstOrCreate(&row).Error; err != nil {
			log.Fatalf("seed tag %q: %v", t.Slug, err)
		}
		out[t.Slug] = row.ID
	}
	return out
}

func seedProblems(db *gorm.DB, dataRoot string, tagIDs map[string]int) {
	type seed struct {
		Slug        string
		Name        string
		Difficulty  string
		Description string
		SampleIn    string
		SampleOut   string
		Template    string
		InputLines  []string
		OutputLines []string
		TagSlugs    []string
		Hints       []string
	}

	seeds := []seed{
		{
			Slug:        "add-two-numbers",
			Name:        "Add Two Numbers",
			Difficulty:  "easy",
			Description: "Given a line `a b`, return the sum a+b as an integer.",
			SampleIn:    "1 2",
			SampleOut:   "3",
			Template: `def solve(line):
    a, b = map(int, line.split())
    return a + b
`,
			InputLines:  []string{"1 2", "10 20", "-5 7", "0 0", "1000 2345"},
			OutputLines: []string{"3", "30", "2", "0", "3345"},
			TagSlugs:    []string{"math", "easy-warmup"},
			Hints:       []string{"Use split() and map(int, ...).", "Return an int, not a string."},
		},
		{
			Slug:        "reverse-string",
			Name:        "Reverse String",
			Difficulty:  "easy",
			Description: "Given a single token on stdin, return it reversed.",
			SampleIn:    "hello",
			SampleOut:   "olleh",
			Template: `def solve(line):
    return line[::-1]
`,
			InputLines:  []string{"hello", "a", "ab", "executify", "racecar"},
			OutputLines: []string{"olleh", "a", "ba", "yfitucexe", "racecar"},
			TagSlugs:    []string{"string", "easy-warmup"},
			Hints:       []string{"Slicing with [::-1] reverses a string in Python."},
		},
		{
			Slug:        "sum-range",
			Name:        "Sum 1..N",
			Difficulty:  "easy",
			Description: "Given an integer N, return the sum 1+2+...+N.",
			SampleIn:    "5",
			SampleOut:   "15",
			Template: `def solve(line):
    n = int(line)
    return n * (n + 1) // 2
`,
			InputLines:  []string{"1", "5", "10", "100", "1000"},
			OutputLines: []string{"1", "15", "55", "5050", "500500"},
			TagSlugs:    []string{"math"},
			Hints:       []string{"Closed form: n(n+1)/2.", "Avoid building the whole list."},
		},
	}

	for _, s := range seeds {
		probDir := filepath.Join(dataRoot, s.Slug)
		if err := os.MkdirAll(probDir, 0o755); err != nil {
			log.Fatalf("mkdir %s: %v", probDir, err)
		}
		inputPath := filepath.Join(probDir, "input.txt")
		outputPath := filepath.Join(probDir, "expected_output.txt")

		if err := os.WriteFile(inputPath, []byte(joinLines(s.InputLines)), 0o644); err != nil {
			log.Fatalf("write input %s: %v", inputPath, err)
		}
		if err := os.WriteFile(outputPath, []byte(joinLines(s.OutputLines)), 0o644); err != nil {
			log.Fatalf("write expected %s: %v", outputPath, err)
		}

		diff := s.Difficulty
		slug := s.Slug
		prob := entity.Problem{
			Name:               s.Name,
			Slug:               &slug,
			Difficulty:         &diff,
			IsPublic:           true,
			Description:        s.Description,
			OutputFormat:       "Single token per line.",
			SampleInput:        s.SampleIn,
			SampleOutput:       s.SampleOut,
			TimeLimit:          1,
			MemoryLimit:        128, // MB; runner converts to KB
			InputFile:          inputPath,
			ExpectedOutputFile: outputPath,
			TemplateCode:       s.Template,
			WrapperCode:        pythonWrapperCode,
			Hints:              datatypes.NewJSONSlice(s.Hints),
		}

		// Upsert by slug — re-running seed should refresh wrapper/paths but
		// not duplicate rows.
		if err := db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "slug"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "difficulty", "is_public", "description",
				"output_format", "sample_input", "sample_output",
				"time_limit", "memory_limit",
				"input_file", "expected_output_file",
				"template_code", "wrapper_code", "hints",
			}),
		}).Create(&prob).Error; err != nil {
			log.Fatalf("upsert problem %q: %v", s.Slug, err)
		}

		// Refresh associations: clear then re-attach.
		if err := db.Exec("DELETE FROM problem_tags WHERE problem_id = ?", prob.ID).Error; err != nil {
			log.Fatalf("clear tags for %d: %v", prob.ID, err)
		}
		for _, tagSlug := range s.TagSlugs {
			tagID, ok := tagIDs[tagSlug]
			if !ok {
				log.Fatalf("problem %q references unknown tag %q", s.Slug, tagSlug)
			}
			if err := db.Exec(
				"INSERT INTO problem_tags (problem_id, tag_id) VALUES (?, ?)",
				prob.ID, tagID,
			).Error; err != nil {
				log.Fatalf("attach tag %q to %q: %v", tagSlug, s.Slug, err)
			}
		}
	}
}

func seedAdmin(db *gorm.DB) {
	hash, err := bcrypt.GenerateFromPassword([]byte(adminPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("hash admin password: %v", err)
	}
	user := entity.User{
		Email:        adminEmail,
		Username:     adminUsername,
		PasswordHash: string(hash),
		IsActive:     true,
	}
	if err := db.Where("email = ?", user.Email).
		Attrs(user).
		FirstOrCreate(&user).Error; err != nil {
		log.Fatalf("seed admin: %v", err)
	}
	log.Printf("admin user ready — email=%s username=%s password=%s", adminEmail, adminUsername, adminPassword)
}

func joinLines(lines []string) string {
	return strings.Join(lines, "\n") + "\n"
}
