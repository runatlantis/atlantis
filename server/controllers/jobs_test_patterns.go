// Copyright 2025 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package controllers

import (
	"fmt"
	"time"
)

// TestPattern defines a pattern for generating test output
type TestPattern string

const (
	TestPatternDefault TestPattern = "default"
	TestPatternSlow    TestPattern = "slow"
	TestPatternBurst   TestPattern = "burst"
	TestPatternColors  TestPattern = "colors"
	TestPatternError   TestPattern = "error"
	TestPatternLong    TestPattern = "long"
)

// GenerateTestOutput generates test terraform-like output for the given pattern
func GenerateTestOutput(pattern TestPattern, output chan<- string) {
	defer close(output)

	switch pattern {
	case TestPatternSlow:
		generateSlowOutput(output)
	case TestPatternBurst:
		generateBurstOutput(output)
	case TestPatternColors:
		generateColorOutput(output)
	case TestPatternError:
		generateErrorOutput(output)
	case TestPatternLong:
		generateLongOutput(output)
	default:
		generateDefaultOutput(output)
	}
}

func generateDefaultOutput(output chan<- string) {
	lines := []string{
		"\033[1mTerraform will perform the following actions:\033[0m",
		"",
		"  # aws_instance.example will be created",
		"\033[32m  + resource \"aws_instance\" \"example\" {\033[0m",
		"\033[32m      + ami                          = \"ami-0c55b159cbfafe1f0\"\033[0m",
		"\033[32m      + instance_type                = \"t2.micro\"\033[0m",
		"\033[32m      + tags                         = {\033[0m",
		"\033[32m          + \"Name\" = \"example-instance\"\033[0m",
		"\033[32m        }\033[0m",
		"\033[32m    }\033[0m",
		"",
		"\033[1mPlan: 1 to add, 0 to change, 0 to destroy.\033[0m",
	}

	for _, line := range lines {
		output <- line
		time.Sleep(100 * time.Millisecond)
	}
}

func generateSlowOutput(output chan<- string) {
	for i := 1; i <= 10; i++ {
		output <- fmt.Sprintf("Processing step %d of 10...", i)
		time.Sleep(1 * time.Second)
	}
	output <- "\033[32mComplete!\033[0m"
}

func generateBurstOutput(output chan<- string) {
	for i := range 500 {
		output <- fmt.Sprintf("Burst line %d: This is a longer line to increase data volume and test buffer handling properly", i)
	}
	output <- "\033[32mBurst complete!\033[0m"
}

func generateColorOutput(output chan<- string) {
	output <- "\033[1;37mANSI Color Showcase\033[0m"
	output <- ""
	output <- "\033[30mBlack\033[0m \033[31mRed\033[0m \033[32mGreen\033[0m \033[33mYellow\033[0m"
	output <- "\033[34mBlue\033[0m \033[35mMagenta\033[0m \033[36mCyan\033[0m \033[37mWhite\033[0m"
	output <- ""
	output <- "\033[1;31mBold Red\033[0m \033[4;32mUnderline Green\033[0m"
	output <- "\033[7;34mReverse Blue\033[0m \033[2;33mDim Yellow\033[0m"
	output <- ""
	output <- "\033[32m+ resource \"added\"\033[0m"
	output <- "\033[31m- resource \"removed\"\033[0m"
	output <- "\033[33m~ resource \"changed\"\033[0m"
}

func generateErrorOutput(output chan<- string) {
	lines := []string{
		"Initializing provider plugins...",
		"- Finding latest version of hashicorp/aws...",
		"- Installing hashicorp/aws v4.67.0...",
		"",
	}
	for _, line := range lines {
		output <- line
		time.Sleep(200 * time.Millisecond)
	}

	output <- "\033[31mError: error configuring Terraform AWS Provider: no valid credential sources found\033[0m"
	output <- ""
	output <- "\033[31m  on main.tf line 1, in provider \"aws\":\033[0m"
	output <- "\033[31m   1: provider \"aws\" {\033[0m"
}

func generateLongOutput(output chan<- string) {
	output <- "\033[1mStarting long-running operation (5 minutes)...\033[0m"

	for i := range 300 {
		output <- fmt.Sprintf("[%02d:%02d] Processing batch %d...", i/60, i%60, i+1)
		time.Sleep(1 * time.Second)
	}

	output <- "\033[32mLong operation complete!\033[0m"
}
