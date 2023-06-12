package raw

type OIDC struct {
	Google *Google `yaml:"google" json:"google"`
	AWS    *AWS    `yaml:"aws" json:"aws"`
}

type Google struct {
	ServiceAccountEmail      string `yaml:"service_account_email,omitempty"`
	WorkloadIdentityProvider string `yaml:"workload_identity_provider,omitempty"`
}

type AWS struct {
	Foobar string `yaml:"foobar,omitempty"`
}
