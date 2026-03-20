package domain

type Language struct {
	ID         int     `json:"id" db:"id"`
	Name       string  `json:"name" db:"name"`
	CompileCmd *string `json:"compile_cmd" db:"compile_cmd"`
	RunCmd     string  `json:"run_cmd" db:"run_cmd"`
	SourceFile string  `json:"source_file" db:"source_file"`
}
