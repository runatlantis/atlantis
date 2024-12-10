package raw

import "github.com/runatlantis/atlantis/server/core/config/valid"

type TeamAuthz struct {
	Command string   `yaml:"command" json:"command"`
	Args    []string `yaml:"args" json:"args"`
}

func (t *TeamAuthz) ToValid() valid.TeamAuthz {
	var v valid.TeamAuthz
	v.Command = t.Command
	v.Args = make([]string, 0)
	if t.Args != nil {
		v.Args = append(v.Args, t.Args...)
	}

	return v
}
