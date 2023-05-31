package raw

type GoogleWorkloadIdentityFederation struct {
	ServiceAccountEmail      string `yaml:"service_account_email,omitempty"`
	WorkloadIdentityProvider string `yaml:"workload_identity_provider,omitempty"`
}
