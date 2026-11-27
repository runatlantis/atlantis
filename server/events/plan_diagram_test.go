// Copyright 2026 The Atlantis Authors
// SPDX-License-Identifier: Apache-2.0

package events_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/runatlantis/atlantis/server/events"
	"github.com/runatlantis/atlantis/server/events/models"
	. "github.com/runatlantis/atlantis/testing"
)

const planDiagramTestJSON = `{
  "format_version": "1.2",
  "terraform_version": "1.9.8",
  "resource_changes": [
    {
      "address": "aws_vpc.main",
      "mode": "managed",
      "type": "aws_vpc",
      "name": "main",
      "change": {"actions": ["create"], "before": null, "after": {"cidr_block": "10.0.0.0/16"}, "after_unknown": {"id": true}}
    },
    {
      "address": "aws_subnet.public[0]",
      "mode": "managed",
      "type": "aws_subnet",
      "name": "public",
      "index": 0,
      "change": {"actions": ["create"], "before": null, "after": {"cidr_block": "10.0.1.0/24"}, "after_unknown": {"id": true}}
    },
    {
      "address": "aws_instance.web",
      "mode": "managed",
      "type": "aws_instance",
      "name": "web",
      "change": {
        "actions": ["update"],
        "before": {"ami": "ami-old", "instance_type": "t3.micro", "tags": {"Name": "web"}},
        "after": {"ami": "ami-new", "instance_type": "t3.micro", "tags": {"Name": "web"}},
        "after_unknown": {"id": false}
      }
    },
    {
      "address": "aws_eip.old",
      "mode": "managed",
      "type": "aws_eip",
      "name": "old",
      "change": {"actions": ["delete"], "before": {"id": "eipalloc-1"}, "after": null, "after_unknown": {}}
    },
    {
      "address": "data.aws_ami.ubuntu",
      "mode": "data",
      "type": "aws_ami",
      "name": "ubuntu",
      "change": {"actions": ["read"], "before": null, "after": {"id": "ami-1"}, "after_unknown": {}}
    },
    {
      "address": "aws_route.nat",
      "mode": "managed",
      "type": "aws_route",
      "name": "nat",
      "change": {"actions": ["create"], "before": null, "after": {}, "after_unknown": {"id": true}}
    },
    {
      "address": "module.vpc.aws_subnet.private",
      "module_address": "module.vpc",
      "mode": "managed",
      "type": "aws_subnet",
      "name": "private",
      "change": {"actions": ["create"], "before": null, "after": {}, "after_unknown": {"id": true}}
    },
    {
      "address": "module.vpc.aws_nat_gateway.this",
      "module_address": "module.vpc",
      "mode": "managed",
      "type": "aws_nat_gateway",
      "name": "this",
      "change": {"actions": ["create"], "before": null, "after": {}, "after_unknown": {"id": true}}
    },
    {
      "address": "module.vpc.module.db.aws_db_instance.main",
      "module_address": "module.vpc.module.db",
      "mode": "managed",
      "type": "aws_db_instance",
      "name": "main",
      "change": {"actions": ["no-op"], "before": {}, "after": {}, "after_unknown": {}}
    },
    {
      "address": "module.app.aws_ssm_parameter.config",
      "module_address": "module.app",
      "mode": "managed",
      "type": "aws_ssm_parameter",
      "name": "config",
      "change": {"actions": ["delete", "create"], "before": {"type": "String"}, "after": {"type": "SecureString"}, "after_unknown": {}}
    }
  ],
  "configuration": {
    "root_module": {
      "resources": [
        {"address": "aws_vpc.main", "mode": "managed", "type": "aws_vpc", "name": "main", "expressions": {"cidr_block": {"constant_value": "10.0.0.0/16"}}},
        {"address": "aws_subnet.public", "mode": "managed", "type": "aws_subnet", "name": "public", "expressions": {"vpc_id": {"references": ["aws_vpc.main.id", "aws_vpc.main"]}}},
        {"address": "aws_instance.web", "mode": "managed", "type": "aws_instance", "name": "web", "expressions": {
          "ami": {"references": ["data.aws_ami.ubuntu.id"]},
          "subnet_id": {"references": ["aws_subnet.public[0].id"]},
          "root_block_device": {"block_types": {"root_block_device": {"nested_blocks": [
            {"expressions": {"volume_size": {"references": ["aws_eip.old.id"]}}}
          ]}}}
        }},
        {"address": "aws_eip.old", "mode": "managed", "type": "aws_eip", "name": "old", "expressions": {}},
        {"address": "data.aws_ami.ubuntu", "mode": "data", "type": "aws_ami", "name": "ubuntu", "expressions": {"owners": {"constant_value": ["099720109477"]}}},
        {"address": "aws_route.nat", "mode": "managed", "type": "aws_route", "name": "nat", "expressions": {"nat_gateway_id": {"references": ["module.vpc.nat_gateway_id"]}}}
      ],
      "module_calls": {
        "vpc": {
          "source": "./modules/vpc",
          "expressions": {"cidr_block": {"references": ["aws_vpc.main.cidr_block"]}},
          "module": {
            "resources": [
              {"address": "aws_subnet.private", "mode": "managed", "type": "aws_subnet", "name": "private", "expressions": {"vpc_id": {"references": ["var.vpc_id"]}}},
              {"address": "aws_nat_gateway.this", "mode": "managed", "type": "aws_nat_gateway", "name": "this", "expressions": {"subnet_id": {"references": ["aws_subnet.private[0].id"]}}}
            ],
            "module_calls": {
              "db": {
                "source": "./modules/db",
                "module": {
                  "resources": [
                    {"address": "aws_db_instance.main", "mode": "managed", "type": "aws_db_instance", "name": "main", "expressions": {"name": {"constant_value": "app"}}}
                  ]
                }
              }
            }
          }
        },
        "app": {
          "source": "./modules/app",
          "expressions": {"database_endpoint": {"references": ["module.vpc.module.db.endpoint"]}},
          "module": {
            "resources": [
              {"address": "aws_ssm_parameter.config", "mode": "managed", "type": "aws_ssm_parameter", "name": "config", "expressions": {"value": {"references": ["var.value"]}}}
            ]
          }
        }
      }
    }
  }
}`

func parseTestDiagram(t *testing.T) *models.PlanDiagram {
	t.Helper()
	diagram, err := events.ParsePlanDiagram([]byte(planDiagramTestJSON))
	if err != nil {
		t.Fatalf("ParsePlanDiagram: %s", err)
	}
	return diagram
}

func diagramNode(t *testing.T, diagram *models.PlanDiagram, address string) models.DiagramNode {
	t.Helper()
	for _, node := range diagram.Nodes {
		if node.Address == address {
			return node
		}
	}
	t.Fatalf("resource %q is not in the diagram", address)
	return models.DiagramNode{}
}

func diagramEdges(diagram *models.PlanDiagram) map[string]bool {
	edges := make(map[string]bool, len(diagram.Edges))
	for _, edge := range diagram.Edges {
		edges[edge.From+" -> "+edge.To] = true
	}
	return edges
}

func TestParsePlanDiagram(t *testing.T) {
	diagram := parseTestDiagram(t)

	Equals(t, 10, len(diagram.Nodes))
	// Nodes are sorted by address so the diagram does not depend on map order.
	Equals(t, "aws_eip.old", diagram.Nodes[0].Address)
	Equals(t, "module.vpc.aws_subnet.private", diagram.Nodes[8].Address)
}

func TestParsePlanDiagram_Actions(t *testing.T) {
	diagram := parseTestDiagram(t)

	cases := map[string]models.DiagramAction{
		"aws_vpc.main":                              models.DiagramActionCreate,
		"aws_instance.web":                          models.DiagramActionUpdate,
		"aws_eip.old":                               models.DiagramActionDelete,
		"data.aws_ami.ubuntu":                       models.DiagramActionRead,
		"module.app.aws_ssm_parameter.config":       models.DiagramActionReplace,
		"module.vpc.module.db.aws_db_instance.main": models.DiagramActionNoOp,
	}
	for address, action := range cases {
		Equals(t, action, diagramNode(t, diagram, address).Action)
	}
}

func TestParsePlanDiagram_ChangedAttributes(t *testing.T) {
	diagram := parseTestDiagram(t)

	// Only "ami" differs; the computed id is not marked unknown.
	Equals(t, 1, diagramNode(t, diagram, "aws_instance.web").ChangedAttributes)
	Equals(t, 1, diagramNode(t, diagram, "module.app.aws_ssm_parameter.config").ChangedAttributes)
	// Create and delete do not report attribute counts.
	Equals(t, 0, diagramNode(t, diagram, "aws_vpc.main").ChangedAttributes)
	Equals(t, 0, diagramNode(t, diagram, "aws_eip.old").ChangedAttributes)
}

func TestParsePlanDiagram_Edges(t *testing.T) {
	diagram := parseTestDiagram(t)
	edges := diagramEdges(diagram)

	expected := []string{
		// Attribute, object and indexed references of the same resource
		// collapse into one edge.
		"aws_vpc.main -> aws_subnet.public[0]",
		"data.aws_ami.ubuntu -> aws_instance.web",
		"aws_subnet.public[0] -> aws_instance.web",
		// A reference inside a nested block still counts.
		"aws_eip.old -> aws_instance.web",
		// A module call's arguments depend on the caller's resources.
		"aws_vpc.main -> module.vpc.aws_subnet.private",
		"aws_vpc.main -> module.vpc.aws_nat_gateway.this",
		// References inside a child module resolve within that module.
		"module.vpc.aws_subnet.private -> module.vpc.aws_nat_gateway.this",
		// A module output reference stands for the module's resources, nested
		// modules included.
		"module.vpc.aws_nat_gateway.this -> aws_route.nat",
		"module.vpc.aws_subnet.private -> aws_route.nat",
		"module.vpc.module.db.aws_db_instance.main -> aws_route.nat",
		// Module calls fan out to every resource of the module they call,
		// nested modules included.
		"aws_vpc.main -> module.vpc.module.db.aws_db_instance.main",
		"module.vpc.module.db.aws_db_instance.main -> module.app.aws_ssm_parameter.config",
	}
	for _, edge := range expected {
		Assert(t, edges[edge], "expected edge %q, got %v", edge, diagram.Edges)
	}
	Equals(t, len(expected), len(diagram.Edges))
}

func TestParsePlanDiagram_ModuleReferenceFanOutIsBounded(t *testing.T) {
	var changes, configs []string
	for i := range 40 {
		changes = append(changes, fmt.Sprintf(
			`{"address": "module.storage.aws_s3_bucket.b%d", "module_address": "module.storage", "mode": "managed", "type": "aws_s3_bucket", "name": "b%d", "change": {"actions": ["create"]}}`, i, i))
		configs = append(configs, fmt.Sprintf(
			`{"address": "aws_s3_bucket.b%d", "mode": "managed", "type": "aws_s3_bucket", "name": "b%d", "expressions": {"bucket": {"constant_value": "b%d"}}}`, i, i, i))
	}
	plan := fmt.Sprintf(`{
	  "resource_changes": [
	    %s,
	    {"address": "aws_iam_policy.reader", "mode": "managed", "type": "aws_iam_policy", "name": "reader", "change": {"actions": ["create"]}}
	  ],
	  "configuration": {
	    "root_module": {
	      "resources": [
	        {"address": "aws_iam_policy.reader", "mode": "managed", "type": "aws_iam_policy", "name": "reader", "expressions": {"description": {"references": ["module.storage"]}}}
	      ],
	      "module_calls": {
	        "storage": {"source": "./storage", "module": {"resources": [
	          %s
	        ]}}
	      }
	    }
	  }
	}`, strings.Join(changes, ","), strings.Join(configs, ","))

	diagram, err := events.ParsePlanDiagram([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePlanDiagram: %s", err)
	}
	assertBoundedFanOut(t, diagram)
}

func assertBoundedFanOut(t *testing.T, diagram *models.PlanDiagram) {
	t.Helper()
	edges := diagramEdges(diagram)
	count := 0
	for edge := range edges {
		if strings.HasSuffix(edge, " -> aws_iam_policy.reader") {
			count++
		}
	}
	Equals(t, 25, count)
}

func TestParsePlanDiagram_NoResources(t *testing.T) {
	for _, plan := range []string{
		`{"format_version": "1.2"}`,
		`{"resource_changes": []}`,
		`{}`,
	} {
		diagram, err := events.ParsePlanDiagram([]byte(plan))
		if err != nil {
			t.Fatalf("ParsePlanDiagram(%s): %s", plan, err)
		}
		Assert(t, diagram == nil, "expected no diagram for %s, got %+v", plan, diagram)
	}
}

func TestParsePlanDiagram_InvalidJSON(t *testing.T) {
	if _, err := events.ParsePlanDiagram([]byte("not json")); err == nil {
		t.Fatal("expected an error for invalid plan JSON")
	}
}

func TestParsePlanDiagram_UnknownReferencesAreIgnored(t *testing.T) {
	plan := `{
	  "resource_changes": [
	    {"address": "aws_instance.web", "mode": "managed", "type": "aws_instance", "name": "web", "change": {"actions": ["create"]}}
	  ],
	  "configuration": {
	    "root_module": {
	      "resources": [
	        {"address": "aws_instance.web", "expressions": {
	          "ami": {"references": ["data.aws_ami.ubuntu.id", "var.ami", "local.name", "each.key", "module.absent.output", "aws_instance.absent.id"]}
	        }}
	      ]
	    }
	  }
	}`
	diagram, err := events.ParsePlanDiagram([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePlanDiagram: %s", err)
	}
	Equals(t, 1, len(diagram.Nodes))
	Equals(t, 0, len(diagram.Edges))
}

func TestParsePlanDiagram_ForgetAction(t *testing.T) {
	plan := `{"resource_changes": [
	  {"address": "aws_s3_bucket.old", "mode": "managed", "type": "aws_s3_bucket", "name": "old", "change": {"actions": ["forget"], "before": {"bucket": "old"}, "after": null}}
	]}`
	diagram, err := events.ParsePlanDiagram([]byte(plan))
	if err != nil {
		t.Fatalf("ParsePlanDiagram: %s", err)
	}
	Equals(t, models.DiagramActionForget, diagram.Nodes[0].Action)
	Assert(t, diagram.Nodes[0].Action.ChangesInfrastructure(), "forgetting a resource is a change")
}
