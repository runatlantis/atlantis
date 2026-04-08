// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

// seed-db populates the Atlantis BoltDB with sample data for UI development and testing.
//
// Usage:
//
//	go run scripts/seed-db/main.go [data-dir]
//
// If data-dir is not specified, it defaults to ./data in the current directory.
// The script creates sample project outputs and locks to exercise the UI.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/runatlantis/atlantis/server/events/models"
	bolt "go.etcd.io/bbolt"
)

const (
	projectOutputsBucketName = "projectOutputs"
	jobIDIndexBucketName     = "job-id-index"
	pullsBucketName          = "pulls"
	locksBucketName          = "runLocks"
)

// Sample terraform outputs with ANSI color codes for realistic rendering
var (
	successfulPlanOutput = "\033[1mTerraform used the selected providers to generate the following execution plan.\033[0m\n" +
		"Resource actions are indicated with the following symbols:\n" +
		"  \033[32m+\033[0m create\n" +
		"  \033[33m~\033[0m update in-place\n" +
		"  \033[31m-\033[0m destroy\n" +
		"\n" +
		"Terraform will perform the following actions:\n" +
		"\n" +
		"  \033[36m# aws_instance.web\033[0m will be created\n" +
		"  \033[32m+\033[0m resource \"aws_instance\" \"web\" {\n" +
		"      \033[32m+\033[0m ami                          = \"ami-0c55b159cbfafe1f0\"\n" +
		"      \033[32m+\033[0m instance_type                = \"t3.micro\"\n" +
		"      \033[32m+\033[0m tags                         = {\n" +
		"          \033[32m+\033[0m \"Environment\" = \"staging\"\n" +
		"          \033[32m+\033[0m \"Name\"        = \"web-server\"\n" +
		"        }\n" +
		"    }\n" +
		"\n" +
		"  \033[36m# aws_security_group.web\033[0m will be created\n" +
		"  \033[32m+\033[0m resource \"aws_security_group\" \"web\" {\n" +
		"      \033[32m+\033[0m description = \"Security group for web servers\"\n" +
		"      \033[32m+\033[0m name        = \"web-sg\"\n" +
		"    }\n" +
		"\n" +
		"\033[1mPlan:\033[0m 2 to add, 0 to change, 0 to destroy."

	changeOutput = "Terraform will perform the following actions:\n" +
		"\n" +
		"  \033[36m# aws_instance.api\033[0m will be updated in-place\n" +
		"  \033[33m~\033[0m resource \"aws_instance\" \"api\" {\n" +
		"        id            = \"i-0abc123def456\"\n" +
		"      \033[33m~\033[0m instance_type = \"t3.small\" -> \"t3.medium\"\n" +
		"    }\n" +
		"\n" +
		"  \033[36m# aws_autoscaling_group.api\033[0m will be updated in-place\n" +
		"  \033[33m~\033[0m resource \"aws_autoscaling_group\" \"api\" {\n" +
		"      \033[33m~\033[0m desired_capacity = 2 -> 4\n" +
		"      \033[33m~\033[0m max_size         = 4 -> 8\n" +
		"    }\n" +
		"\n" +
		"\033[1mPlan:\033[0m 0 to add, 2 to change, 0 to destroy."

	destroyOutput = "Terraform will perform the following actions:\n" +
		"\n" +
		"  \033[36m# aws_instance.legacy\033[0m will be \033[31mdestroyed\033[0m\n" +
		"  \033[31m-\033[0m resource \"aws_instance\" \"legacy\" {\n" +
		"      \033[31m-\033[0m ami           = \"ami-oldversion123\"\n" +
		"      \033[31m-\033[0m instance_type = \"t2.micro\"\n" +
		"    }\n" +
		"\n" +
		"  \033[36m# aws_ebs_volume.legacy_data\033[0m will be \033[31mdestroyed\033[0m\n" +
		"  \033[31m-\033[0m resource \"aws_ebs_volume\" \"legacy_data\" {\n" +
		"      \033[31m-\033[0m size = 100\n" +
		"    }\n" +
		"\n" +
		"\033[1mPlan:\033[0m 0 to add, 0 to change, 2 to destroy."

	failedOutput = "Initializing the backend...\n" +
		"\n" +
		"\033[31m\033[1mError:\033[0m \033[1mFailed to get existing workspaces: S3 bucket does not exist.\033[0m\n" +
		"\n" +
		"The referenced S3 bucket must have been previously created.\n" +
		"\n" +
		"\033[31m\033[1mError:\033[0m \033[1mInitialization failed\033[0m"

	noChangesOutput = "No changes. Your infrastructure matches the configuration.\n\n" +
		"Terraform has compared your real infrastructure against your configuration\n" +
		"and found no differences, so no changes are needed."

	policyPassOutput = `Running policy check...

Checking policies for aws_instance.web...
  [PASS] instance_type must be t3 series
  [PASS] must have Environment tag

All policies passed! 2/2 checks succeeded.`

	policyFailOutput = `Running policy check...

Checking policies for aws_instance.public...
  [FAIL] must not have public IP in production
         Found: associate_public_ip_address = true

Policy check failed! 1/2 checks failed.`

	applySuccessOutput = "\033[36maws_instance.api\033[0m: Modifying... [id=i-0abc123]\n" +
		"\033[36maws_instance.api\033[0m: Modifications complete after 15s\n" +
		"\n" +
		"\033[32m\033[1mApply complete!\033[0m Resources: 0 added, 1 changed, 0 destroyed."

	applyFailedOutput = "\033[36maws_db_instance.primary\033[0m: Modifying...\n" +
		"\n" +
		"\033[31m\033[1mError:\033[0m \033[1mCannot upgrade mysql from 5.7 to 8.0 directly.\033[0m\n" +
		"\n" +
		"\033[31m\033[1mError:\033[0m \033[1mApply failed.\033[0m"

	// Realistic error output with timestamps and ANSI codes like terragrunt output
	longErrorOutput = "running 'sh -c 'TERRAGRUNT_STRICT_CONTROL=skip-dependencies-inputs terragrunt plan -out=$PLANFILE' in '/home/atlantis/data/repos/acme/infrastructure/102/terraform_environments_production_us-east-1_web': exit status 1: running \"TERRAGRUNT_STRICT_CONTROL=skip-dependencies-inputs terragrunt plan -out=$PLANFILE\" in \"/home/atlantis/data/repos/acme/infrastructure/102/terraform_environments_production_us-east-1_web\":\n" +
		"07:59:39.429 WARN   The `TERRAGRUNT_STRICT_CONTROL` environment variable is deprecated and will be removed in a future version of Terragrunt. Use `TG_STRICT_CONTROL=skip-dependencies-inputs,skip-dependencies-inputs` instead.\n" +
		"07:59:39.429 WARN   The following strict control(s) are already completed: skip-dependencies-inputs. Please remove any completed strict controls, as setting them no longer does anything. For a list of all ongoing strict controls, and the outcomes of previous strict controls, see https://terragrunt.gruntwork.io/docs/reference/strict-mode or get the actual list by running the `terragrunt info strict` command.\n" +
		"07:59:39.430 WARN   The `TERRAGRUNT_TFPATH` environment variable is deprecated and will be removed in a future version of Terragrunt. Use `TG_TF_PATH=terraform1.13.3` instead.\n" +
		"07:59:39.547 INFO   terraform1.13.3: Initializing the backend...\n" +
		"07:59:39.594 INFO   terraform1.13.3: Successfully configured the backend \"gcs\"! Terraform will automatically\n" +
		"07:59:39.594 INFO   terraform1.13.3: use this backend unless the backend configuration changes.\n" +
		"07:59:39.669 INFO   terraform1.13.3: Initializing provider plugins...\n" +
		"07:59:39.669 INFO   terraform1.13.3: - Finding latest version of hashicorp/aws...\n" +
		"07:59:39.751 INFO   terraform1.13.3: - Using hashicorp/aws v5.31.0 from the shared cache directory\n" +
		"07:59:39.772 INFO   terraform1.13.3: Terraform has created a lock file .terraform.lock.hcl to record the provider\n" +
		"07:59:39.772 INFO   terraform1.13.3: selections it made above. Include this file in your version control repository\n" +
		"07:59:39.772 INFO   terraform1.13.3: so that Terraform can guarantee to make the same selections by default when\n" +
		"07:59:39.772 INFO   terraform1.13.3: you run \"terraform init\" in the future.\n" +
		"07:59:39.772 INFO   terraform1.13.3: Terraform has been successfully initialized!\n" +
		"07:59:40.418 STDOUT terraform1.13.3: aws_instance.web: Refreshing state... [id=i-0abc123def456]\n" +
		"07:59:40.419 STDOUT terraform1.13.3: aws_security_group.web: Refreshing state... [id=sg-0123456789]\n" +
		"07:59:40.628 STDOUT terraform1.13.3: Changes to Outputs:\n" +
		"07:59:40.628 STDOUT terraform1.13.3:   ~ instance_id = \"i-0abc123def456\" -> null\n" +
		"07:59:40.628 STDOUT terraform1.13.3:   ~ public_ip   = \"10.0.1.100\" -> null\n" +
		"07:59:40.628 STDOUT terraform1.13.3: You can apply this plan to save these new output values to the Terraform\n" +
		"07:59:40.628 STDOUT terraform1.13.3: state, without changing any real infrastructure.\n" +
		"\033[31m07:59:40.628 STDERR terraform1.13.3: Error: error configuring S3 Backend: no valid credential sources for S3 Backend found.\033[0m\n" +
		"\033[31m07:59:40.628 STDERR terraform1.13.3: Error: Failed to get existing workspaces: S3 bucket \"acme-terraform-state\" does not exist or is not accessible\033[0m"
)

// GenerateLockKey creates the key for the locks bucket (matches boltdb.go)
func GenerateLockKey(project models.Project, workspace string) string {
	return fmt.Sprintf("%s/%s/%s/%s", project.RepoFullName, project.Path, workspace, project.ProjectName)
}

func main() {
	dataDir := "data"
	if len(os.Args) > 1 {
		dataDir = os.Args[1]
	}

	dbPath := filepath.Join(dataDir, "atlantis.db")
	fmt.Printf("Seeding database at: %s\n", dbPath)

	if err := os.MkdirAll(dataDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating data dir: %v\n", err)
		os.Exit(1)
	}

	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening database: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	// Delete and recreate buckets for clean state
	err = db.Update(func(tx *bolt.Tx) error {
		_ = tx.DeleteBucket([]byte(projectOutputsBucketName))
		_ = tx.DeleteBucket([]byte(jobIDIndexBucketName))
		_ = tx.DeleteBucket([]byte(locksBucketName))

		if _, err := tx.CreateBucketIfNotExists([]byte(projectOutputsBucketName)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(jobIDIndexBucketName)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(pullsBucketName)); err != nil {
			return err
		}
		if _, err := tx.CreateBucketIfNotExists([]byte(locksBucketName)); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating buckets: %v\n", err)
		os.Exit(1)
	}

	now := time.Now()

	// Define sample project outputs
	outputs := []models.ProjectOutput{
		// ============================================================
		// PR #101 - acme/infrastructure - "Add VPC endpoints for S3 and DynamoDB"
		// Multiple projects, vpc has run history: plan -> apply -> plan
		// ============================================================
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       101,
			PullURL:       "https://github.com/acme/infrastructure/pull/101",
			PullTitle:     "Add VPC endpoints for S3 and DynamoDB",
			Path:          "terraform/environments/staging/us-east-1/vpc",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-1 * time.Hour).UnixMilli(),
			Output:        successfulPlanOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 2, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			PolicyOutput:  policyPassOutput,
			TriggeredBy:   "jwalton",
			StartedAt:     now.Add(-1 * time.Hour),
			CompletedAt:   now.Add(-1*time.Hour + 2*time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       101,
			PullURL:       "https://github.com/acme/infrastructure/pull/101",
			PullTitle:     "Add VPC endpoints for S3 and DynamoDB",
			Path:          "terraform/environments/staging/us-east-1/vpc",
			Workspace:     "default",
			CommandName:   "apply",
			RunTimestamp:  now.Add(-30 * time.Minute).UnixMilli(),
			Output:        applySuccessOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 2, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "jwalton",
			StartedAt:     now.Add(-30 * time.Minute),
			CompletedAt:   now.Add(-29 * time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       101,
			PullURL:       "https://github.com/acme/infrastructure/pull/101",
			PullTitle:     "Add VPC endpoints for S3 and DynamoDB",
			Path:          "terraform/environments/staging/us-east-1/vpc",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-10 * time.Minute).UnixMilli(),
			Output:        noChangesOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			PolicyOutput:  policyPassOutput,
			TriggeredBy:   "jwalton",
			StartedAt:     now.Add(-10 * time.Minute),
			CompletedAt:   now.Add(-8 * time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       101,
			PullURL:       "https://github.com/acme/infrastructure/pull/101",
			PullTitle:     "Add VPC endpoints for S3 and DynamoDB",
			Path:          "terraform/environments/staging/us-east-1/ecs",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-10 * time.Minute).UnixMilli(),
			Output:        changeOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 2, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "jwalton",
			StartedAt:     now.Add(-10 * time.Minute),
			CompletedAt:   now.Add(-7 * time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       101,
			PullURL:       "https://github.com/acme/infrastructure/pull/101",
			PullTitle:     "Add VPC endpoints for S3 and DynamoDB",
			Path:          "terraform/environments/staging/us-east-1/rds",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-10 * time.Minute).UnixMilli(),
			Output:        noChangesOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "jwalton",
			StartedAt:     now.Add(-10 * time.Minute),
			CompletedAt:   now.Add(-9 * time.Minute),
		},

		// ============================================================
		// PR #102 - acme/infrastructure - "Upgrade API servers to t3.large"
		// Has failures and policy violations
		// ============================================================
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       102,
			PullURL:       "https://github.com/acme/infrastructure/pull/102",
			PullTitle:     "Upgrade API servers to t3.large",
			Path:          "terraform/environments/production/us-east-1/api",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-2 * time.Hour).UnixMilli(),
			Output:        failedOutput,
			Status:        models.FailedOutputStatus,
			Error:         "exit status 1: initialization failed",
			ResourceStats: models.ResourceStats{},
			PolicyPassed:  false,
			TriggeredBy:   "ssmith",
			StartedAt:     now.Add(-2 * time.Hour),
			CompletedAt:   now.Add(-2*time.Hour + 1*time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       102,
			PullURL:       "https://github.com/acme/infrastructure/pull/102",
			PullTitle:     "Upgrade API servers to t3.large",
			Path:          "terraform/environments/production/us-east-1/api",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-30 * time.Minute).UnixMilli(),
			Output:        changeOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 2, Destroy: 0},
			PolicyPassed:  false,
			PolicyOutput:  policyFailOutput,
			TriggeredBy:   "ssmith",
			StartedAt:     now.Add(-30 * time.Minute),
			CompletedAt:   now.Add(-28 * time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       102,
			PullURL:       "https://github.com/acme/infrastructure/pull/102",
			PullTitle:     "Upgrade API servers to t3.large",
			Path:          "terraform/environments/production/us-east-1/web",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-30 * time.Minute).UnixMilli(),
			Output:        failedOutput,
			Status:        models.FailedOutputStatus,
			Error:         longErrorOutput,
			ResourceStats: models.ResourceStats{},
			PolicyPassed:  false,
			TriggeredBy:   "ssmith",
			StartedAt:     now.Add(-30 * time.Minute),
			CompletedAt:   now.Add(-29 * time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       102,
			PullURL:       "https://github.com/acme/infrastructure/pull/102",
			PullTitle:     "Upgrade API servers to t3.large",
			Path:          "terraform/environments/production/us-east-1/cache",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-30 * time.Minute).UnixMilli(),
			Output:        successfulPlanOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 2, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "ssmith",
			StartedAt:     now.Add(-30 * time.Minute),
			CompletedAt:   now.Add(-27 * time.Minute),
		},

		// ============================================================
		// PR #103 - acme/infrastructure - "Database version upgrade to MySQL 8.0"
		// Shows: plan -> apply FAIL -> plan -> apply SUCCESS -> plan
		// ============================================================
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       103,
			PullURL:       "https://github.com/acme/infrastructure/pull/103",
			PullTitle:     "Database version upgrade to MySQL 8.0",
			Path:          "terraform/environments/production/us-east-1/database",
			Workspace:     "default",
			ProjectName:   "mysql-primary",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-3 * time.Hour).UnixMilli(),
			Output:        changeOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 1, Destroy: 0},
			PolicyPassed:  true,
			PolicyOutput:  policyPassOutput,
			TriggeredBy:   "dba-team",
			StartedAt:     now.Add(-3 * time.Hour),
			CompletedAt:   now.Add(-3*time.Hour + 2*time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       103,
			PullURL:       "https://github.com/acme/infrastructure/pull/103",
			PullTitle:     "Database version upgrade to MySQL 8.0",
			Path:          "terraform/environments/production/us-east-1/database",
			Workspace:     "default",
			ProjectName:   "mysql-primary",
			CommandName:   "apply",
			RunTimestamp:  now.Add(-150 * time.Minute).UnixMilli(),
			Output:        applyFailedOutput,
			Status:        models.FailedOutputStatus,
			Error:         "exit status 1: incompatible MySQL version upgrade",
			ResourceStats: models.ResourceStats{Add: 0, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "dba-team",
			StartedAt:     now.Add(-150 * time.Minute),
			CompletedAt:   now.Add(-149 * time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       103,
			PullURL:       "https://github.com/acme/infrastructure/pull/103",
			PullTitle:     "Database version upgrade to MySQL 8.0",
			Path:          "terraform/environments/production/us-east-1/database",
			Workspace:     "default",
			ProjectName:   "mysql-primary",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-1 * time.Hour).UnixMilli(),
			Output:        changeOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 1, Destroy: 0},
			PolicyPassed:  true,
			PolicyOutput:  policyPassOutput,
			TriggeredBy:   "dba-team",
			StartedAt:     now.Add(-1 * time.Hour),
			CompletedAt:   now.Add(-58 * time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       103,
			PullURL:       "https://github.com/acme/infrastructure/pull/103",
			PullTitle:     "Database version upgrade to MySQL 8.0",
			Path:          "terraform/environments/production/us-east-1/database",
			Workspace:     "default",
			ProjectName:   "mysql-primary",
			CommandName:   "apply",
			RunTimestamp:  now.Add(-45 * time.Minute).UnixMilli(),
			Output:        applySuccessOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 1, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "dba-team",
			StartedAt:     now.Add(-45 * time.Minute),
			CompletedAt:   now.Add(-44 * time.Minute),
		},
		{
			RepoFullName:  "acme/infrastructure",
			PullNum:       103,
			PullURL:       "https://github.com/acme/infrastructure/pull/103",
			PullTitle:     "Database version upgrade to MySQL 8.0",
			Path:          "terraform/environments/production/us-east-1/database",
			Workspace:     "default",
			ProjectName:   "mysql-primary",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-20 * time.Minute).UnixMilli(),
			Output:        noChangesOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			PolicyOutput:  policyPassOutput,
			TriggeredBy:   "dba-team",
			StartedAt:     now.Add(-20 * time.Minute),
			CompletedAt:   now.Add(-18 * time.Minute),
		},

		// ============================================================
		// PR #203 - acme/microservices - "Add authentication service"
		// Has pending job
		// ============================================================
		{
			RepoFullName:  "acme/microservices",
			PullNum:       203,
			PullURL:       "https://github.com/acme/microservices/pull/203",
			PullTitle:     "Add authentication service",
			Path:          "services/auth/terraform",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-5 * time.Minute).UnixMilli(),
			Output:        successfulPlanOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 2, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "mchen",
			StartedAt:     now.Add(-5 * time.Minute),
			CompletedAt:   now.Add(-3 * time.Minute),
		},
		{
			RepoFullName:  "acme/microservices",
			PullNum:       203,
			PullURL:       "https://github.com/acme/microservices/pull/203",
			PullTitle:     "Add authentication service",
			Path:          "services/payments/terraform",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-2 * time.Minute).UnixMilli(),
			Output:        "",
			Status:        models.RunningOutputStatus,
			ResourceStats: models.ResourceStats{},
			PolicyPassed:  true,
			TriggeredBy:   "mchen",
			StartedAt:     now.Add(-2 * time.Minute),
		},

		// ============================================================
		// PR #55 - acme/data-platform - "Remove deprecated ETL pipeline"
		// Destroy operations
		// ============================================================
		{
			RepoFullName:  "acme/data-platform",
			PullNum:       55,
			PullURL:       "https://github.com/acme/data-platform/pull/55",
			PullTitle:     "Remove deprecated ETL pipeline",
			Path:          "terraform/legacy-cleanup",
			Workspace:     "default",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-1 * time.Hour).UnixMilli(),
			Output:        destroyOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 0, Destroy: 2},
			PolicyPassed:  true,
			TriggeredBy:   "devops-bot",
			StartedAt:     now.Add(-1 * time.Hour),
			CompletedAt:   now.Add(-58 * time.Minute),
		},

		// ============================================================
		// PR #789 - bigcorp/platform - "Upgrade EKS to 1.29"
		// Large PR with many projects
		// ============================================================
		{
			RepoFullName:  "bigcorp/platform",
			PullNum:       789,
			PullURL:       "https://github.com/bigcorp/platform/pull/789",
			PullTitle:     "Upgrade EKS to 1.29",
			Path:          "infra/networking/vpc",
			Workspace:     "production",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-15 * time.Minute).UnixMilli(),
			Output:        successfulPlanOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 2, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "platform-team",
			StartedAt:     now.Add(-15 * time.Minute),
			CompletedAt:   now.Add(-12 * time.Minute),
		},
		{
			RepoFullName:  "bigcorp/platform",
			PullNum:       789,
			PullURL:       "https://github.com/bigcorp/platform/pull/789",
			PullTitle:     "Upgrade EKS to 1.29",
			Path:          "infra/networking/dns",
			Workspace:     "production",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-15 * time.Minute).UnixMilli(),
			Output:        changeOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 2, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "platform-team",
			StartedAt:     now.Add(-15 * time.Minute),
			CompletedAt:   now.Add(-11 * time.Minute),
		},
		{
			RepoFullName:  "bigcorp/platform",
			PullNum:       789,
			PullURL:       "https://github.com/bigcorp/platform/pull/789",
			PullTitle:     "Upgrade EKS to 1.29",
			Path:          "infra/compute/eks",
			Workspace:     "production",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-15 * time.Minute).UnixMilli(),
			Output:        successfulPlanOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 2, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "platform-team",
			StartedAt:     now.Add(-15 * time.Minute),
			CompletedAt:   now.Add(-10 * time.Minute),
		},
		{
			RepoFullName:  "bigcorp/platform",
			PullNum:       789,
			PullURL:       "https://github.com/bigcorp/platform/pull/789",
			PullTitle:     "Upgrade EKS to 1.29",
			Path:          "infra/compute/ec2",
			Workspace:     "production",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-15 * time.Minute).UnixMilli(),
			Output:        changeOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 2, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "platform-team",
			StartedAt:     now.Add(-15 * time.Minute),
			CompletedAt:   now.Add(-9 * time.Minute),
		},
		{
			RepoFullName:  "bigcorp/platform",
			PullNum:       789,
			PullURL:       "https://github.com/bigcorp/platform/pull/789",
			PullTitle:     "Upgrade EKS to 1.29",
			Path:          "infra/storage/s3",
			Workspace:     "production",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-15 * time.Minute).UnixMilli(),
			Output:        successfulPlanOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 2, Change: 0, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "platform-team",
			StartedAt:     now.Add(-15 * time.Minute),
			CompletedAt:   now.Add(-8 * time.Minute),
		},
		{
			RepoFullName:  "bigcorp/platform",
			PullNum:       789,
			PullURL:       "https://github.com/bigcorp/platform/pull/789",
			PullTitle:     "Upgrade EKS to 1.29",
			Path:          "infra/storage/rds",
			Workspace:     "production",
			CommandName:   "plan",
			RunTimestamp:  now.Add(-15 * time.Minute).UnixMilli(),
			Output:        changeOutput,
			Status:        models.SuccessOutputStatus,
			ResourceStats: models.ResourceStats{Add: 0, Change: 2, Destroy: 0},
			PolicyPassed:  true,
			TriggeredBy:   "platform-team",
			StartedAt:     now.Add(-15 * time.Minute),
			CompletedAt:   now.Add(-7 * time.Minute),
		},
	}

	// Save all outputs
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(projectOutputsBucketName))
		indexBucket := tx.Bucket([]byte(jobIDIndexBucketName))
		for _, output := range outputs {
			key := output.Key()
			bytes, err := json.Marshal(output)
			if err != nil {
				return fmt.Errorf("marshaling output: %w", err)
			}
			if err := bucket.Put([]byte(key), bytes); err != nil {
				return fmt.Errorf("saving output: %w", err)
			}
			if output.JobID != "" {
				if err := indexBucket.Put([]byte(output.JobID), []byte(key)); err != nil {
					return fmt.Errorf("saving job ID index: %w", err)
				}
			}
			fmt.Printf("  Saved: %s\n", key)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving outputs: %v\n", err)
		os.Exit(1)
	}

	// Define sample locks
	locks := []models.ProjectLock{
		{
			Project: models.Project{
				RepoFullName: "acme/infrastructure",
				Path:         "terraform/environments/staging/us-east-1/vpc",
			},
			Pull: models.PullRequest{
				Num:        101,
				HeadCommit: "abc123def456789",
				URL:        "https://github.com/acme/infrastructure/pull/101",
				HeadBranch: "feature/add-vpc-endpoints",
				BaseBranch: "main",
				Author:     "jwalton",
				State:      models.OpenPullState,
				BaseRepo: models.Repo{
					FullName: "acme/infrastructure",
					Owner:    "acme",
					Name:     "infrastructure",
					VCSHost: models.VCSHost{
						Hostname: "github.com",
						Type:     models.Github,
					},
				},
			},
			User: models.User{
				Username: "jwalton",
			},
			Workspace: "default",
			Time:      now.Add(-10 * time.Minute),
		},
		{
			Project: models.Project{
				RepoFullName: "acme/infrastructure",
				Path:         "terraform/environments/staging/us-east-1/ecs",
			},
			Pull: models.PullRequest{
				Num:        101,
				HeadCommit: "abc123def456789",
				URL:        "https://github.com/acme/infrastructure/pull/101",
				HeadBranch: "feature/add-vpc-endpoints",
				BaseBranch: "main",
				Author:     "jwalton",
				State:      models.OpenPullState,
				BaseRepo: models.Repo{
					FullName: "acme/infrastructure",
					Owner:    "acme",
					Name:     "infrastructure",
					VCSHost: models.VCSHost{
						Hostname: "github.com",
						Type:     models.Github,
					},
				},
			},
			User: models.User{
				Username: "jwalton",
			},
			Workspace: "default",
			Time:      now.Add(-10 * time.Minute),
		},
		{
			Project: models.Project{
				ProjectName:  "eks-cluster",
				RepoFullName: "bigcorp/platform",
				Path:         "infra/compute/eks",
			},
			Pull: models.PullRequest{
				Num:        789,
				HeadCommit: "def789abc123456",
				URL:        "https://github.com/bigcorp/platform/pull/789",
				HeadBranch: "feature/upgrade-eks-1.29",
				BaseBranch: "main",
				Author:     "platform-team",
				State:      models.OpenPullState,
				BaseRepo: models.Repo{
					FullName: "bigcorp/platform",
					Owner:    "bigcorp",
					Name:     "platform",
					VCSHost: models.VCSHost{
						Hostname: "github.com",
						Type:     models.Github,
					},
				},
			},
			User: models.User{
				Username: "platform-bot",
			},
			Workspace: "production",
			Time:      now.Add(-15 * time.Minute),
		},
		{
			Project: models.Project{
				ProjectName:  "rds-primary",
				RepoFullName: "bigcorp/platform",
				Path:         "infra/storage/rds",
			},
			Pull: models.PullRequest{
				Num:        789,
				HeadCommit: "def789abc123456",
				URL:        "https://github.com/bigcorp/platform/pull/789",
				HeadBranch: "feature/upgrade-eks-1.29",
				BaseBranch: "main",
				Author:     "platform-team",
				State:      models.OpenPullState,
				BaseRepo: models.Repo{
					FullName: "bigcorp/platform",
					Owner:    "bigcorp",
					Name:     "platform",
					VCSHost: models.VCSHost{
						Hostname: "github.com",
						Type:     models.Github,
					},
				},
			},
			User: models.User{
				Username: "platform-bot",
			},
			Workspace: "production",
			Time:      now.Add(-7 * time.Minute),
		},
	}

	// Save all locks
	err = db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(locksBucketName))
		for _, lock := range locks {
			key := GenerateLockKey(lock.Project, lock.Workspace)
			bytes, err := json.Marshal(lock)
			if err != nil {
				return fmt.Errorf("marshaling lock: %w", err)
			}
			if err := bucket.Put([]byte(key), bytes); err != nil {
				return fmt.Errorf("saving lock: %w", err)
			}
			fmt.Printf("  Saved lock: %s\n", key)
		}
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error saving locks: %v\n", err)
		os.Exit(1)
	}

	// Summary
	fmt.Println("\n--- Seed Complete ---")
	fmt.Printf("\nCreated %d project outputs across 6 PRs:\n", len(outputs))
	fmt.Println("  - acme/infrastructure #101: Add VPC endpoints for S3 and DynamoDB")
	fmt.Println("  - acme/infrastructure #102: Upgrade API servers to t3.large")
	fmt.Println("  - acme/infrastructure #103: Database version upgrade to MySQL 8.0")
	fmt.Println("  - acme/microservices #203: Add authentication service")
	fmt.Println("  - acme/data-platform #55: Remove deprecated ETL pipeline")
	fmt.Println("  - bigcorp/platform #789: Upgrade EKS to 1.29")
	fmt.Printf("\nCreated %d locks\n", len(locks))
	fmt.Println("\nRestart Atlantis server and visit: http://localhost:4141/prs")
}
