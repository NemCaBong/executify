package domain

type Language struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	CompileCmd *string `json:"compile_cmd"`
	RunCmd     string  `json:"run_cmd"`
	SourceFile string  `json:"source_file"`
}
