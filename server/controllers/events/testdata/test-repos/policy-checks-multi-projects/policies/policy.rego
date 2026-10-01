package main

import input as tfplan

deny contains reason if {
	num_creates[_] > 0
	reason := "WARNING: Forbidden Resource creation is prohibited."
}

resource_names = {"forbidden"}

resources[resource_name] = all if {
	some resource_name
	resource_names[resource_name]
	all := [res |
		res := tfplan.resource_changes[_]
		res.name == resource_name
	]
}

# number of creations of resources of a given name
num_creates[resource_name] = num if {
	some resource_name
	resource_names[resource_name]
	all := resources[resource_name]
	creations := [res | res := all[_]; res.change.actions[_] == "create"]
	num := count(creations)
}
