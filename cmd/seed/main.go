// cmd/seed loads the JSON fixtures under seed/data and writes them into the
// database so a fresh checkout can be demoed end-to-end.
//
// Layout:
//
//	seed/data/
//	  languages.json   — rows for the `languages` table
//	  users.json       — admin users (passwords are bcrypt-hashed at seed time)
//	  tags.json        — rows for `tags`
//	  problems.json    — problem metadata + per-language wrapper/template
//	                     code inlined as JSON strings (newlines as \n)
//	  problems/<slug>/ — per-problem test fixtures: input.txt and
//	                     expected_output.txt, one test case per line
//
// Idempotency: every entity is upserted by its natural key (language name,
// tag slug, problem slug, user email, (problem_id, language_id) pair) so
// re-running the seeder refreshes content without duplicating rows.
//
// Problem wrapper code: each wrapper MUST write its final answer to fd 3 and
// close fd 3, because the worker invokes the program as
//
//	<run_cmd> <<< "<stdin line>" 3>"<actual_output_file>"
//
// i.e. bash redirects fd 3 to a file before exec'ing the program.
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/NemCaBong/executify/internal/adapter/repository/entity"
	"github.com/NemCaBong/executify/internal/config"
)

// dataRoot is the on-disk root for seed fixtures. Resolved once at startup so
// every path the seeder writes (input/expected files) is absolute, since the
// worker later reads them from a different cwd.
var dataRoot string

type languageSeed struct {
	Name       string  `json:"name"`
	Version    string  `json:"version"`
	SourceFile string  `json:"source_file"`
	CompileCmd *string `json:"compile_cmd"`
	RunCmd     string  `json:"run_cmd"`
	Notes      string  `json:"notes"`
}

type userSeed struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
	IsActive bool   `json:"is_active"`
}

type tagSeed struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type problemLanguageSeed struct {
	Language string `json:"language"`
	Wrapper  string `json:"wrapper"`
	Template string `json:"template"`
}

type ioSchemaSeed struct {
	Kind      string `json:"kind"`
	LineIndex int    `json:"line_index"`
	KeyName   string `json:"key_name"`
	DataType  string `json:"data_type"`
}

type problemSeed struct {
	Slug                     string                `json:"slug"`
	Name                     string                `json:"name"`
	Difficulty               *string               `json:"difficulty"`
	IsPublic                 bool                  `json:"is_public"`
	Description              string                `json:"description"`
	OutputFormat             string                `json:"output_format"`
	SampleInput              string                `json:"sample_input"`
	SampleOutput             string                `json:"sample_output"`
	TimeLimit                int                   `json:"time_limit"`
	MemoryLimit              int                   `json:"memory_limit"`
	CPUTimeLimit             *float64              `json:"cpu_time_limit"`
	CPUExtraTime             *float64              `json:"cpu_extra_time"`
	WallTimeLimit            *float64              `json:"wall_time_limit"`
	StackLimit               *int                  `json:"stack_limit"`
	MaxProcessesAndOrThreads *int                  `json:"max_processes_and_or_threads"`
	FloatPrecision           *int                  `json:"float_precision"`
	InputFile                string                `json:"input_file"`
	ExpectedOutputFile       string                `json:"expected_output_file"`
	Tags                     []string              `json:"tags"`
	Hints                    []string              `json:"hints"`
	IOSchema                 []ioSchemaSeed        `json:"io_schema"`
	Languages                []problemLanguageSeed `json:"languages"`
}

func main() {
	cfg := config.Load()
	db := config.NewMySQLConnection(cfg)

	root, err := filepath.Abs("seed/data")
	if err != nil {
		log.Fatalf("resolve data root: %v", err)
	}
	dataRoot = root

	langIDs := seedLanguages(db, loadJSON[[]languageSeed](filepath.Join(dataRoot, "languages.json")))
	tagIDs := seedTags(db, loadJSON[[]tagSeed](filepath.Join(dataRoot, "tags.json")))
	seedProblems(db, loadJSON[[]problemSeed](filepath.Join(dataRoot, "problems.json")), langIDs, tagIDs)
	seedUsers(db, loadJSON[[]userSeed](filepath.Join(dataRoot, "users.json")))

	log.Printf("seed complete — %d language(s), %d tag(s)", len(langIDs), len(tagIDs))
}

// loadJSON reads and parses a JSON file. Fatal on any error so partial seeds
// don't silently land in the database.
func loadJSON[T any](path string) T {
	raw, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("read %s: %v", path, err)
	}
	var out T
	if err := json.Unmarshal(raw, &out); err != nil {
		log.Fatalf("parse %s: %v", path, err)
	}
	return out
}

func seedLanguages(db *gorm.DB, seeds []languageSeed) map[string]int {
	out := make(map[string]int, len(seeds))
	for _, s := range seeds {
		row := entity.Language{
			Name:       s.Name,
			CompileCmd: s.CompileCmd,
			RunCmd:     s.RunCmd,
			SourceFile: s.SourceFile,
		}
		// languages.name has no unique index, so clause.OnConflict can't dedup
		// (it INSERTs a fresh row every run). Look up by name and create only
		// when absent — keeping a stable ID across re-seeds — then refresh the
		// mutable command columns from JSON.
		if err := db.Where("name = ?", s.Name).
			Attrs(row).
			FirstOrCreate(&row).Error; err != nil {
			log.Fatalf("seed language %q: %v", s.Name, err)
		}
		if err := db.Model(&row).Updates(map[string]any{
			"compile_cmd": s.CompileCmd,
			"run_cmd":     s.RunCmd,
			"source_file": s.SourceFile,
		}).Error; err != nil {
			log.Fatalf("update language %q: %v", s.Name, err)
		}
		out[s.Name] = row.ID
		log.Printf("language ready — id=%d name=%q version=%s", row.ID, s.Name, s.Version)
	}
	return out
}

func seedTags(db *gorm.DB, seeds []tagSeed) map[string]int {
	out := make(map[string]int, len(seeds))
	for _, s := range seeds {
		row := entity.Tag{Name: s.Name, Slug: s.Slug}
		if err := db.Where("slug = ?", s.Slug).
			Attrs(row).
			FirstOrCreate(&row).Error; err != nil {
			log.Fatalf("seed tag %q: %v", s.Slug, err)
		}
		// Refresh name even if slug already existed, so JSON edits propagate.
		if row.Name != s.Name {
			if err := db.Model(&row).Update("name", s.Name).Error; err != nil {
				log.Fatalf("update tag name %q: %v", s.Slug, err)
			}
		}
		out[s.Slug] = row.ID
	}
	return out
}

func seedProblems(db *gorm.DB, seeds []problemSeed, langIDs, tagIDs map[string]int) {
	for _, s := range seeds {
		probDir := filepath.Join(dataRoot, "problems", s.Slug)

		// input/expected paths must be absolute because the worker reads
		// them from a different cwd.
		inputPath := filepath.Join(probDir, s.InputFile)
		expectedPath := filepath.Join(probDir, s.ExpectedOutputFile)
		if _, err := os.Stat(inputPath); err != nil {
			log.Fatalf("problem %q: input file missing: %v", s.Slug, err)
		}
		if _, err := os.Stat(expectedPath); err != nil {
			log.Fatalf("problem %q: expected output file missing: %v", s.Slug, err)
		}

		slug := s.Slug
		prob := entity.Problem{
			Name:                     s.Name,
			Slug:                     &slug,
			Difficulty:               s.Difficulty,
			IsPublic:                 s.IsPublic,
			Description:              s.Description,
			OutputFormat:             s.OutputFormat,
			SampleInput:              s.SampleInput,
			SampleOutput:             s.SampleOutput,
			TimeLimit:                s.TimeLimit,
			MemoryLimit:              s.MemoryLimit,
			CPUTimeLimit:             s.CPUTimeLimit,
			CPUExtraTime:             s.CPUExtraTime,
			WallTimeLimit:            s.WallTimeLimit,
			StackLimit:               s.StackLimit,
			MaxProcessesAndOrThreads: s.MaxProcessesAndOrThreads,
			FloatPrecision:           s.FloatPrecision,
			InputFile:                inputPath,
			ExpectedOutputFile:       expectedPath,
			Hints:                    datatypes.NewJSONSlice(s.Hints),
		}

		if err := db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "slug"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "difficulty", "is_public", "description",
				"output_format", "sample_input", "sample_output",
				"time_limit", "memory_limit",
				"cpu_time_limit", "cpu_extra_time", "wall_time_limit",
				"stack_limit", "max_processes_and_or_threads", "float_precision",
				"input_file", "expected_output_file", "hints",
			}),
		}).Create(&prob).Error; err != nil {
			log.Fatalf("upsert problem %q: %v", s.Slug, err)
		}

		// MySQL's ON DUPLICATE KEY UPDATE leaves prob.ID at 0 on the update
		// path, so re-fetch the real problem_id; the child rows (tags,
		// languages, io_schema) below must reference it, not 0.
		if prob.ID == 0 {
			if err := db.Where("slug = ?", slug).First(&prob).Error; err != nil {
				log.Fatalf("load problem id for %q: %v", s.Slug, err)
			}
		}

		// Refresh tag associations: clear then re-attach so renamed/removed
		// tags in JSON take effect.
		if err := db.Exec("DELETE FROM problem_tags WHERE problem_id = ?", prob.ID).Error; err != nil {
			log.Fatalf("clear tags for %q: %v", s.Slug, err)
		}
		for _, slug := range s.Tags {
			tagID, ok := tagIDs[slug]
			if !ok {
				log.Fatalf("problem %q references unknown tag %q", s.Slug, slug)
			}
			if err := db.Exec(
				"INSERT INTO problem_tags (problem_id, tag_id) VALUES (?, ?)",
				prob.ID, tagID,
			).Error; err != nil {
				log.Fatalf("attach tag %q to %q: %v", slug, s.Slug, err)
			}
		}

		// Refresh per-language wrappers/templates the same way: blow away
		// existing rows for this problem and rewrite from JSON.
		if err := db.Exec("DELETE FROM problem_languages WHERE problem_id = ?", prob.ID).Error; err != nil {
			log.Fatalf("clear problem_languages for %q: %v", s.Slug, err)
		}
		for _, pl := range s.Languages {
			langID, ok := langIDs[pl.Language]
			if !ok {
				log.Fatalf("problem %q references unknown language %q", s.Slug, pl.Language)
			}
			row := entity.ProblemLanguage{
				ProblemID:    prob.ID,
				LanguageID:   langID,
				WrapperCode:  pl.Wrapper,
				TemplateCode: pl.Template,
			}
			if err := db.Create(&row).Error; err != nil {
				log.Fatalf("insert problem_language %q/%q: %v", s.Slug, pl.Language, err)
			}
		}

		// Refresh the IO schema: drop the old rows then rewrite from JSON, so
		// the runtime knows how many lines each test case spans and how to type
		// each field (see domain.CompareOutput / code_runner test-case chunking).
		if err := db.Exec("DELETE FROM problem_io_schema WHERE problem_id = ?", prob.ID).Error; err != nil {
			log.Fatalf("clear problem_io_schema for %q: %v", s.Slug, err)
		}
		for _, f := range s.IOSchema {
			row := entity.ProblemIOSchema{
				ProblemID: prob.ID,
				Kind:      f.Kind,
				LineIndex: f.LineIndex,
				KeyName:   f.KeyName,
				DataType:  f.DataType,
			}
			if err := db.Create(&row).Error; err != nil {
				log.Fatalf("insert problem_io_schema %q/%s[%d]: %v", s.Slug, f.Kind, f.LineIndex, err)
			}
		}

		log.Printf("problem ready — id=%d slug=%q languages=%d io_fields=%d", prob.ID, s.Slug, len(s.Languages), len(s.IOSchema))
	}
}

func seedUsers(db *gorm.DB, seeds []userSeed) {
	for _, s := range seeds {
		hash, err := bcrypt.GenerateFromPassword([]byte(s.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("hash password for %q: %v", s.Email, err)
		}
		user := entity.User{
			Email:        s.Email,
			Username:     s.Username,
			PasswordHash: string(hash),
			IsActive:     s.IsActive,
		}
		// FirstOrCreate by email — re-seeding does NOT overwrite an existing
		// password hash, so a user who rotated their password locally is safe.
		if err := db.Where("email = ?", user.Email).
			Attrs(user).
			FirstOrCreate(&user).Error; err != nil {
			log.Fatalf("seed user %q: %v", s.Email, err)
		}
		log.Printf("user ready — email=%s username=%s%s",
			s.Email, s.Username, devPasswordHint(s))
	}
}

// devPasswordHint prints the plaintext password only for the seed-fixture
// admin account, never for an arbitrary user.
func devPasswordHint(s userSeed) string {
	if s.Email == "admin@executify.local" {
		return fmt.Sprintf(" password=%s (dev only)", s.Password)
	}
	return ""
}
