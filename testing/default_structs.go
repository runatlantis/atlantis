package testing

import "github.com/runatlantis/atlantis/server/events/yaml/valid"

// Helper function to return a frequently repeated valid Autoplan struct.
func DefaultValidAutoplan() valid.Autoplan {
    return valid.Autoplan{
        WhenModified: []string{"**/*.tf*", "**/terragrunt.hcl"},
        Enabled:      true,
    }
}
