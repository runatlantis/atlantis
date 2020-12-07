package terraform

import (
	"github.com/runatlantis/atlantis/server/logging"
	. "github.com/runatlantis/atlantis/testing"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestNewOutput(t *testing.T) {
	// Create tmp folder to hold the mocked tf output files
	tmp, cleanup := TempDir(t)
	defer cleanup()

	_, err := NewOutputHelper(tmp)
	Ok(t, err)
}

func TestFileOutputHelper_List(t *testing.T) {
	// Create tmp folder to hold the mocked tf output files
	tmp, cleanup := TempDir(t)
	defer cleanup()

	helper, err := NewOutputHelper(tmp)
	Ok(t, err)

	tfOutputFileNames := []string{
		"20201121175848-runatalntis_atlantis-1-1a2b3c4-test-default-init",
		"20201121175849-runatalntis_atlantis-1-1a2b3c4-test-default-plan",
		"20201121175850-runatalntis_atlantis-1-1a2b3c4-test-default-apply",
	}

	var tfOutputs []TfOutputFile

	// Create the mocked files and parse the file name.
	for _, tfOutputFileName := range tfOutputFileNames {
		_, err := os.Create(filepath.Join(tmp, tfOutputFileName))
		Ok(t, err)

		tfOutput, err := helper.ParseFileName(tfOutputFileName)
		Ok(t, err)

		tfOutputs = append(tfOutputs, tfOutput)
	}

	outputs, err := helper.List()
	Ok(t, err)

	// Sort both slices with the same criteria
	sort.Slice(tfOutputs, func(i, j int) bool {
		return tfOutputs[i].CreatedAt.Before(tfOutputs[j].CreatedAt)
	})

	sort.Slice(outputs, func(i, j int) bool {
		return outputs[i].CreatedAt.Before(outputs[j].CreatedAt)
	})

	Equals(t, tfOutputs, outputs)
}

func TestFileOutputHelper_ParseFileName(t *testing.T) {
	type fields struct {
		outputCmdDir string
	}
	type args struct {
		fileName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		exp     TfOutputFile
		wantErr bool
	}{
		{
			name: "File name should be parsed",
			fields: fields{
				outputCmdDir: "test",
			},
			args: args{
				fileName: "20201125130042-runatlantis_atlantis-1-1a2b3c4-test-default-init",
			},
			exp: TfOutputFile{
				CreatedAt: func() time.Time {
					createdAt, err := time.Parse(outputTimeFmt, "20201125130042")
					Ok(t, err)
					return createdAt
				}(),
				FullRepoName:  "runatlantis/atlantis",
				PullRequestNr: 1,
				HeadCommit:    "1a2b3c4",
				Project:       "test",
				TfCommand:     "init",
				Workspace:     "default",
			},
			wantErr: false,
		},
		{
			name: "File name should be parsed when repo name has '-'",
			fields: fields{
				outputCmdDir: "test",
			},
			args: args{
				fileName: "20201125130042-runatlantis_atlantis__test-1-1a2b3c4-test-default-init",
			},
			exp: TfOutputFile{
				CreatedAt: func() time.Time {
					createdAt, err := time.Parse(outputTimeFmt, "20201125130042")
					Ok(t, err)
					return createdAt
				}(),
				FullRepoName:  "runatlantis/atlantis-test",
				PullRequestNr: 1,
				HeadCommit:    "1a2b3c4",
				Project:       "test",
				TfCommand:     "init",
				Workspace:     "default",
			},
			wantErr: false,
		},
		{
			name: "File name should be parsed when project name has '-'",
			fields: fields{
				outputCmdDir: "test",
			},
			args: args{
				fileName: "20201125130042-runatlantis_atlantis__test-1-1a2b3c4-test_123-default-init",
			},
			exp: TfOutputFile{
				CreatedAt: func() time.Time {
					createdAt, err := time.Parse(outputTimeFmt, "20201125130042")
					Ok(t, err)
					return createdAt
				}(),
				FullRepoName:  "runatlantis/atlantis-test",
				PullRequestNr: 1,
				HeadCommit:    "1a2b3c4",
				Project:       "test-123",
				TfCommand:     "init",
				Workspace:     "default",
			},
			wantErr: false,
		},
		{
			name: "It should fail with invalid file name",
			args: args{
				fileName: "test123",
			},
			exp:     TfOutputFile{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &FileOutputHelper{
				outputCmdDir: tt.fields.outputCmdDir,
			}
			got, err := f.ParseFileName(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			Equals(t, tt.exp, got)
		})
	}
}

func TestFileOutputHelper_ContinueReadFile(t *testing.T) {
	tmp, cleanup := TempDir(t)
	defer cleanup()

	testFileName := strconv.FormatInt(time.Now().UnixNano(), 10)
	testFile, err := os.OpenFile(filepath.Join(tmp, testFileName), os.O_CREATE|os.O_WRONLY, os.ModePerm)
	Ok(t, err)
	defer testFile.Close()

	helper, err := NewOutputHelper(tmp)
	Ok(t, err)

	log := logging.NewSimpleLogger("test", false, logging.Debug)
	content := make(chan string)
	done := make(chan bool)
	go func() {
		err = helper.ContinueReadFile(log, testFileName, content, done)
		Ok(t, err)
	}()

	testNewLines := []string{"ab\n", "cd\n", "ef\n"}
	fullMsg := ""
	for i, data := range testNewLines {
		_, err = testFile.WriteString(data)
		Ok(t, err)

		// Receives the output file from the channel
		msg := <-content
		fullMsg += msg

		// Verify if the buff has all the data written in the file being read
		Equals(t, strings.Join(testNewLines[:i+1], ""), fullMsg)
	}
	// Stop the continue read file method
	done <- true
}

func TestFileOutputHelper_CreateFileName(t *testing.T) {
	type args struct {
		fullRepoName  string
		pullRequestNr int
		headCommit    string
		project       string
		workspace     string
		tfCommand     string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "It should create a file name when the repo name has '-'",
			args: args{
				fullRepoName:  "test/test-repo",
				pullRequestNr: 1,
				headCommit:    "1234567890",
				project:       "test",
				workspace:     "default",
				tfCommand:     "init",
			},
			want: "test_test__repo-1-1234567-test-default-init",
		},
		{
			name: "It should create a file name",
			args: args{
				fullRepoName:  "test/test",
				pullRequestNr: 1,
				headCommit:    "1234567890",
				project:       "test",
				workspace:     "default",
				tfCommand:     "init",
			},
			want: "test_test-1-1234567-test-default-init",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helper := &FileOutputHelper{
				outputCmdDir: "",
			}

			got := helper.CreateFileName(
				tt.args.fullRepoName,
				tt.args.pullRequestNr,
				tt.args.headCommit,
				tt.args.project,
				tt.args.workspace,
				tt.args.tfCommand,
			)

			parts := strings.Split(got, "-")
			reg, err := regexp.Compile("[0-9]{14}$")
			Ok(t, err)

			// Check if the first part of the file name is a timestamp
			Equals(t, true, reg.Match([]byte(parts[0])))

			// Checks the other parts of the file name
			Equals(t, tt.want, strings.Join(parts[1:], "-"))
		})
	}
}

func TestFileOutputHelper_FindOutputFile(t *testing.T) {
	type args struct {
		createdAt    string
		fullRepoName string
		pullNr       string
		headCommit   string
		project      string
		workspace    string
		tfCommand    string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "It should find the file",
			args:    args{
				createdAt:    "20200101000000",
				fullRepoName: "test/test-repo",
				pullNr:       "1",
				headCommit:   "1234567",
				project:      "test",
				workspace:    "default",
				tfCommand:    "init",
			},
			want:    "20200101000000-test_test__repo-1-1234567-test-default-init",
			wantErr: false,
		},
		{
			name:    "It should not find the file",
			args:    args{
				createdAt:    "20200101000000",
				fullRepoName: "test/test-repo",
				pullNr:       "1",
				headCommit:   "1234567",
				project:      "test",
				workspace:    "default",
				tfCommand:    "init",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmp, cleanup := TempDir(t)
			defer cleanup()

			// Do not create the test file if want is empty
			if len(tt.want) > 0 {
				f, err := os.Create(filepath.Join(tmp, tt.want))
				Ok(t, err)
				err = f.Close()
				Ok(t, err)
			}

			helper := &FileOutputHelper{
				outputCmdDir: tmp,
			}

			got, err := helper.FindOutputFile(tt.args.createdAt, tt.args.fullRepoName, tt.args.pullNr, tt.args.headCommit, tt.args.project, tt.args.workspace, tt.args.tfCommand)
			if (err != nil) != tt.wantErr {
				t.Errorf("FindOutputFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindOutputFile() got = %v, want %v", got, tt.want)
			}
		})
	}
}
