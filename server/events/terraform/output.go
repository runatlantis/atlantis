package terraform

// Output define the operations related to the terraform output files
type Output interface {
	// List list all files from terraform output dir
	List() ([]string, error)
}

type FileOutput struct {
	outputCmdDir string
}

// NewOutput creates a terraform output
func NewOutput(outputCmdDir string) Output {
	return &FileOutput{outputCmdDir: outputCmdDir}
}

func (f *FileOutput) List() ([]string, error) {
	return []string{}, nil
}
