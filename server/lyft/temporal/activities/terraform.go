package activities

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type InitRequest struct {
	RootDir string
}

type InitResponse struct {
	Output string
}

type PlanRequest struct {
	RootDir string
}

type PlanResponse struct {
	Output   string
	Planfile string
}

type ApplyRequest struct {
	RootDir  string
	Planfile string
}

type ApplyResponse struct {
	Output string
}

func Init(ctx context.Context, request InitRequest) (InitResponse, error) {
	output, err := terraform(request.RootDir, "init", "-input=false")

	// output using fmt instead of logger to get pretty formatting.
	// this is probably susceptible to duplication though in the event of replays
	// this is only for prototype purposes anyways.
	fmt.Println(string(output))

	if err != nil {
		return InitResponse{}, errors.Wrap(err, "running terraform init")
	}

	return InitResponse{Output: string(output)}, nil
}

func Plan(ctx context.Context, request PlanRequest) (PlanResponse, error) {
	planFile := "plan.tfplan"

	output, err := terraform(request.RootDir, "plan", "-input=false", "-refresh", "-out", fmt.Sprintf("%q", planFile))

	// output using fmt instead of logger to get pretty formatting.
	// this is probably susceptible to duplication though in the event of replays
	// this is only for prototype purposes anyways.
	fmt.Println(string(output))

	if err != nil {
		return PlanResponse{}, errors.Wrap(err, "running terraform plan")
	}

	return PlanResponse{
		Output:   string(output),
		Planfile: planFile,
	}, nil
}

func Apply(ctx context.Context, request ApplyRequest) (ApplyResponse, error) {
	output, err := terraform(request.RootDir, "apply", "-input=false", filepath.Join(request.RootDir, request.Planfile))

	// output using fmt instead of logger to get pretty formatting.
	// this is probably susceptible to duplication though in the event of replays
	// this is only for prototype purposes anyways.
	fmt.Println(string(output))

	if err != nil {
		return ApplyResponse{}, errors.Wrap(err, "running terraform apply")
	}

	return ApplyResponse{Output: string(output)}, nil
}

func terraform(dir string, args ...string) ([]byte, error) {
	tfCmd := fmt.Sprintf("terraform %s", strings.Join(args, " "))
	cmd := exec.Command("sh", "-c", tfCmd)
	cmd.Dir = dir
	cmd.Env = os.Environ()

	return cmd.CombinedOutput()
}
