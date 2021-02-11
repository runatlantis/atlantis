package main

import input as tfplan

deny[reason] {
	num_deletes.null_resource > 0
	reason := "WARNING: Null Resource creation is prohibited."
}

resource_types = {"null_resource"}

resources[resource_type] = all {
	some resource_type
	resource_types[resource_type]
	all := [name |
		name := tfplan.resource_changes[_]
		name.type == resource_type
	]
}

# number of deletions of resources of a given type
num_deletes[resource_type] = num {
	some resource_type
	resource_types[resource_type]
	all := resources[resource_type]
	deletions := [res | res := all[_]; res.change.actions[_] == "create"]
	num := count(deletions)
}
