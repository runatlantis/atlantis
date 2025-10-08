# Atlantis Locking Integration Guide

A practical, hands-on guide for integrating Atlantis locking systems with your infrastructure workflows. This guide covers real-world deployment patterns, common scenarios, and operational best practices.

## Quick Start Integration

### Step 1: Assess Your Current Setup

Before implementing locking, understand your current Atlantis deployment:

```bash
# Check your current locking configuration
atlantis server --help | grep -i lock

# Review existing atlantis.yaml files
find . -name "atlantis.yaml" -exec echo "=== {} ===" \; -exec cat {} \;

# Check for parallel configuration
grep -r "parallel_plan\|parallel_apply" .
```

### Step 2: Choose Your Locking Strategy

| Scenario | Recommended Strategy | Lock Mode | Parallel Settings |
|----------|---------------------|-----------|------------------|
| Single developer/team | Legacy locking | `on_plan` | `parallel_plan: false` |
| Small team (2-5 people) | Legacy locking | `on_plan` | `parallel_plan: true` |
| Medium team (5-15 people) | Enhanced locking | `on_plan` | `parallel_plan: true, parallel_apply: false` |
| Large team (15+ people) | Enhanced locking | `on_apply` | `parallel_plan: true, parallel_apply: true` |
| Multi-tenant environment | Enhanced locking | `on_apply` | Full parallel + priority queues |

### Step 3: Basic Integration

Start with this minimal `atlantis.yaml`:

```yaml
version: 3
# Conservative settings for initial deployment
parallel_plan: false
parallel_apply: false

projects:
  - name: main-infrastructure
    dir: .
    workspace: default
    repo_locks:
      mode: on_plan  # Lock during planning phase

    # Basic autoplan configuration
    autoplan:
      when_modified: ["**/*.tf", "**/*.tfvars"]
      enabled: true

    # Safety requirements
    plan_requirements: [mergeable]
    apply_requirements: [mergeable]
```

## Directory-Based vs Project-Based Locking

### Directory-Based Locking Pattern

Best for: Simple repository structures with clear directory boundaries.

```yaml
version: 3
# Each directory gets independent locking
projects:
  # Infrastructure directory
  - name: infrastructure
    dir: infrastructure
    workspace: default
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["infrastructure/**/*.tf"]

  # Applications directory
  - name: applications
    dir: applications
    workspace: default
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["applications/**/*.tf"]

  # Monitoring directory - can run independently
  - name: monitoring
    dir: monitoring
    workspace: default
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["monitoring/**/*.tf"]
```

**Benefits:**
- Simple to understand and implement
- Clear isolation boundaries
- Good for microservices architectures
- Parallel execution across directories

**Drawbacks:**
- May not capture cross-directory dependencies
- Can lead to resource conflicts if directories share resources

### Project-Based Locking Pattern

Best for: Complex deployments with explicit dependencies and multiple environments.

```yaml
version: 3
# Project-based with explicit dependencies
projects:
  # Core networking - must be deployed first
  - name: core-networking
    dir: infrastructure/network
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply  # Lock during destructive operations
    autoplan:
      when_modified: ["infrastructure/network/**/*.tf"]

  # Security infrastructure - depends on networking
  - name: security-infrastructure
    dir: infrastructure/security
    workspace: production
    execution_order_group: 2
    depends_on: ["core-networking"]
    repo_locks:
      mode: on_apply
    autoplan:
      when_modified: ["infrastructure/security/**/*.tf"]

  # Application infrastructure - depends on both
  - name: app-infrastructure
    dir: infrastructure/compute
    workspace: production
    execution_order_group: 3
    depends_on: ["core-networking", "security-infrastructure"]
    repo_locks:
      mode: on_apply
    autoplan:
      when_modified: ["infrastructure/compute/**/*.tf"]

  # Applications - final deployment stage
  - name: web-applications
    dir: applications
    workspace: production
    execution_order_group: 4
    depends_on: ["app-infrastructure"]
    repo_locks:
      mode: on_plan  # Less critical, can lock during planning
    autoplan:
      when_modified: ["applications/**/*.tf"]
```

**Benefits:**
- Explicit dependency management
- Controlled deployment order
- Better for complex, interdependent infrastructure
- Supports execution order groups

**Drawbacks:**
- More complex to set up initially
- Requires understanding of infrastructure dependencies
- May reduce parallelism due to dependencies

## Workspace Isolation Patterns

### Multi-Environment Workspace Isolation

Perfect for maintaining separate development, staging, and production environments:

```yaml
version: 3
parallel_plan: true
parallel_apply: false  # Conservative for cross-environment safety

projects:
  # Development environment - faster, less restrictive
  - name: api-development
    dir: api
    workspace: development
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["api/**/*.tf", "api/**/*.tfvars"]
      enabled: true
    # No approval requirements for development

  # Staging environment - some protection
  - name: api-staging
    dir: api
    workspace: staging
    execution_order_group: 1  # Deploy staging first
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["api/**/*.tf", "api/**/*.tfvars"]
      enabled: true
    plan_requirements: [mergeable]
    apply_requirements: [mergeable]

  # Production environment - full protection
  - name: api-production
    dir: api
    workspace: production
    execution_order_group: 2  # Deploy production after staging
    repo_locks:
      mode: on_apply  # Strict locking for production
    autoplan:
      when_modified: ["api/**/*.tf", "api/**/*.tfvars"]
      enabled: true
    plan_requirements: [approved, mergeable, undiverged]
    apply_requirements: [approved, mergeable, undiverged]

  # Shared resources - affect all environments
  - name: shared-dns
    dir: shared/dns
    workspace: global
    execution_order_group: 0  # Deploy shared resources first
    repo_locks:
      mode: on_apply  # Critical shared resource
    autoplan:
      when_modified: ["shared/dns/**/*.tf"]
      enabled: true
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]
```

### Multi-Tenant Workspace Isolation

For SaaS platforms or managed service providers:

```yaml
version: 3
parallel_plan: true
parallel_apply: true  # Safe due to workspace isolation

projects:
  # Customer A infrastructure
  - name: customer-a-infrastructure
    dir: customers/infrastructure
    workspace: customer-a
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["customers/infrastructure/**/*.tf", "customers/customer-a.tfvars"]
      enabled: true

  # Customer A applications
  - name: customer-a-applications
    dir: customers/applications
    workspace: customer-a
    depends_on: ["customer-a-infrastructure"]
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["customers/applications/**/*.tf", "customers/customer-a.tfvars"]
      enabled: true

  # Customer B infrastructure (completely isolated)
  - name: customer-b-infrastructure
    dir: customers/infrastructure
    workspace: customer-b
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["customers/infrastructure/**/*.tf", "customers/customer-b.tfvars"]
      enabled: true

  # Customer B applications
  - name: customer-b-applications
    dir: customers/applications
    workspace: customer-b
    depends_on: ["customer-b-infrastructure"]
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["customers/applications/**/*.tf", "customers/customer-b.tfvars"]
      enabled: true

workflows:
  customer-deployment:
    plan:
      steps:
        - init
        - run: terraform workspace select ${WORKSPACE}
        - plan:
            extra_args: ["-var-file", "customers/${WORKSPACE}.tfvars"]
    apply:
      steps:
        - run: terraform workspace select ${WORKSPACE}
        - apply:
            extra_args: ["-var-file", "customers/${WORKSPACE}.tfvars"]
```

## Parallel Execution and Locking Coordination

### Understanding Parallel Behavior

Parallel execution interacts with locking in specific ways:

```yaml
version: 3
# Global parallel settings
parallel_plan: true    # Plans run in parallel across projects
parallel_apply: true   # Applies run in parallel across projects
abort_on_execution_order_fail: true  # Stop all if any fails

projects:
  # Group 1: Independent infrastructure components
  - name: vpc
    dir: infrastructure/vpc
    execution_order_group: 1
    repo_locks:
      mode: on_plan
    # Will run in parallel with other group 1 projects

  - name: iam-roles
    dir: infrastructure/iam
    execution_order_group: 1  # Same group as VPC
    repo_locks:
      mode: on_plan
    # Runs in parallel with VPC

  # Group 2: Components that depend on Group 1
  - name: eks-cluster
    dir: infrastructure/eks
    execution_order_group: 2
    depends_on: ["vpc", "iam-roles"]
    repo_locks:
      mode: on_apply  # More critical component
    # Won't start until Group 1 completes

  - name: rds-database
    dir: infrastructure/rds
    execution_order_group: 2  # Same group as EKS
    depends_on: ["vpc"]
    repo_locks:
      mode: on_apply
    # Runs in parallel with EKS cluster

  # Group 3: Applications
  - name: web-app
    dir: applications/web
    execution_order_group: 3
    depends_on: ["eks-cluster", "rds-database"]
    repo_locks:
      mode: on_plan  # Less critical
```

### High-Throughput Parallel Configuration

For environments with many microservices or frequent deployments:

```yaml
version: 3
# Maximize parallelism
parallel_plan: true
parallel_apply: true

# Server configuration should include:
# parallel-pool-size: 50  # Increase from default 15

projects:
  # Microservice pattern with independent locking
  - name: user-service-dev
    dir: services/user-service
    workspace: development
    repo_locks:
      mode: on_plan  # Fast locking for development
    autoplan:
      when_modified: ["services/user-service/**/*.tf"]
      enabled: true

  - name: user-service-prod
    dir: services/user-service
    workspace: production
    repo_locks:
      mode: on_apply  # Careful locking for production
    plan_requirements: [approved]
    apply_requirements: [approved]

  - name: order-service-dev
    dir: services/order-service
    workspace: development
    repo_locks:
      mode: on_plan

  - name: order-service-prod
    dir: services/order-service
    workspace: production
    repo_locks:
      mode: on_apply
    plan_requirements: [approved]
    apply_requirements: [approved]

  # Shared services with careful coordination
  - name: shared-redis-prod
    dir: shared/redis
    workspace: production
    execution_order_group: 1  # Deploy shared services first
    repo_locks:
      mode: on_apply
    plan_requirements: [approved]
    apply_requirements: [approved]
```

## Real-World Deployment Scenarios

### Scenario 1: Single-Tenant Web Application

**Requirements:**
- Development, staging, production environments
- Database, application, and CDN components
- Small team (3-5 developers)

```yaml
version: 3
parallel_plan: true
parallel_apply: false  # Sequential applies for safety

projects:
  # Database tier
  - name: database-dev
    dir: infrastructure/database
    workspace: development
    execution_order_group: 1
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["infrastructure/database/**/*.tf", "env/dev.tfvars"]

  - name: database-staging
    dir: infrastructure/database
    workspace: staging
    execution_order_group: 1
    repo_locks:
      mode: on_plan
    plan_requirements: [mergeable]
    apply_requirements: [mergeable]

  - name: database-prod
    dir: infrastructure/database
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply  # Strict locking for production DB
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  # Application tier
  - name: app-dev
    dir: infrastructure/application
    workspace: development
    execution_order_group: 2
    depends_on: ["database-dev"]
    repo_locks:
      mode: on_plan

  - name: app-staging
    dir: infrastructure/application
    workspace: staging
    execution_order_group: 2
    depends_on: ["database-staging"]
    repo_locks:
      mode: on_plan
    plan_requirements: [mergeable]
    apply_requirements: [mergeable]

  - name: app-prod
    dir: infrastructure/application
    workspace: production
    execution_order_group: 2
    depends_on: ["database-prod"]
    repo_locks:
      mode: on_apply
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  # CDN tier
  - name: cdn-prod
    dir: infrastructure/cdn
    workspace: production
    execution_order_group: 3
    depends_on: ["app-prod"]
    repo_locks:
      mode: on_plan  # CDN changes are less risky
    plan_requirements: [approved]
    apply_requirements: [approved]

workflows:
  # Custom workflow for database migrations
  database-migration:
    plan:
      steps:
        - init
        - run: echo "Planning database migration for workspace $WORKSPACE"
        - plan:
            extra_args: ["-var-file", "env/$WORKSPACE.tfvars"]
    apply:
      steps:
        - run: echo "Starting database migration - this may take several minutes"
        - run: ./scripts/pre-migration-backup.sh $WORKSPACE
        - apply:
            extra_args: ["-var-file", "env/$WORKSPACE.tfvars"]
        - run: ./scripts/post-migration-verify.sh $WORKSPACE
```

Server configuration:
```yaml
# Conservative settings for single-tenant
parallel-pool-size: 5
enhanced-locking:
  enabled: false  # Use legacy locking for simplicity
```

### Scenario 2: Multi-Tenant SaaS Platform

**Requirements:**
- Multiple customer tenants
- Shared infrastructure and tenant-specific resources
- High availability and isolation
- Large team (20+ developers)

```yaml
version: 3
parallel_plan: true
parallel_apply: true  # Safe due to tenant isolation

projects:
  # Shared platform infrastructure
  - name: platform-networking
    dir: platform/networking
    workspace: shared
    execution_order_group: 1
    repo_locks:
      mode: on_apply  # Critical shared infrastructure
    autoplan:
      when_modified: ["platform/networking/**/*.tf"]
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]
    silence_pr_comments: ["plan"]  # Reduce noise for shared components

  - name: platform-security
    dir: platform/security
    workspace: shared
    execution_order_group: 1
    repo_locks:
      mode: on_apply
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  - name: platform-monitoring
    dir: platform/monitoring
    workspace: shared
    execution_order_group: 2
    depends_on: ["platform-networking"]
    repo_locks:
      mode: on_apply
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  # Customer tenant resources
  - name: tenant-enterprise-corp
    dir: tenants/infrastructure
    workspace: enterprise-corp
    execution_order_group: 3
    depends_on: ["platform-networking", "platform-security"]
    repo_locks:
      mode: on_plan  # Tenant isolation allows lighter locking
    autoplan:
      when_modified: ["tenants/**/*.tf", "tenants/configs/enterprise-corp.tfvars"]

  - name: tenant-startup-inc
    dir: tenants/infrastructure
    workspace: startup-inc
    execution_order_group: 3
    depends_on: ["platform-networking", "platform-security"]
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["tenants/**/*.tf", "tenants/configs/startup-inc.tfvars"]

  # Tenant-specific applications
  - name: apps-enterprise-corp
    dir: tenants/applications
    workspace: enterprise-corp
    execution_order_group: 4
    depends_on: ["tenant-enterprise-corp"]
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: ["tenants/applications/**/*.tf", "tenants/configs/enterprise-corp.tfvars"]

  - name: apps-startup-inc
    dir: tenants/applications
    workspace: startup-inc
    execution_order_group: 4
    depends_on: ["tenant-startup-inc"]
    repo_locks:
      mode: on_plan

workflows:
  tenant-provisioning:
    plan:
      steps:
        - init
        - run: echo "Provisioning resources for tenant $WORKSPACE"
        - run: terraform workspace select $WORKSPACE
        - plan:
            extra_args: ["-var-file", "tenants/configs/$WORKSPACE.tfvars"]
    apply:
      steps:
        - run: terraform workspace select $WORKSPACE
        - run: ./scripts/pre-tenant-provisioning.sh $WORKSPACE
        - apply:
            extra_args: ["-var-file", "tenants/configs/$WORKSPACE.tfvars"]
        - run: ./scripts/post-tenant-provisioning.sh $WORKSPACE

  shared-platform:
    plan:
      steps:
        - init
        - run: echo "Planning shared platform changes"
        - plan
    apply:
      steps:
        - run: echo "Applying shared platform changes - this affects all tenants"
        - run: ./scripts/notify-maintenance-start.sh
        - apply
        - run: ./scripts/verify-platform-health.sh
        - run: ./scripts/notify-maintenance-end.sh
```

Server configuration:
```yaml
# High-performance settings for multi-tenant
parallel-pool-size: 25

enhanced-locking:
  enabled: true
  default-timeout: "20m"

  priority-queue:
    enabled: true
    max-queue-size: 1000

  retry:
    enabled: true
    max-attempts: 3

  deadlock-detection:
    enabled: true
    resolution-policy: "lowest_priority"

  redis:
    cluster-mode: true  # High availability
    key-prefix: "saas:atlantis:lock:"
```

### Scenario 3: High-Throughput DevOps Pipeline

**Requirements:**
- 50+ microservices
- Continuous deployment
- Multiple environments per service
- DevOps team of 15+ engineers

```yaml
version: 3
# Maximum parallelism for high throughput
parallel_plan: true
parallel_apply: true
abort_on_execution_order_fail: false  # Don't let one service block others

# Auto-generated projects pattern (use script to generate full config)
projects:
  # Core infrastructure deployed first
  - name: core-kubernetes-prod
    dir: core/kubernetes
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply
    autoplan:
      when_modified: ["core/kubernetes/**/*.tf"]
      enabled: true
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  # Service mesh infrastructure
  - name: core-service-mesh-prod
    dir: core/service-mesh
    workspace: production
    execution_order_group: 1
    repo_locks:
      mode: on_apply
    depends_on: ["core-kubernetes-prod"]
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  # Microservices - high parallelism
  - name: user-service-dev
    dir: services/user-service
    workspace: development
    execution_order_group: 2
    repo_locks:
      mode: disabled  # No locking for dev deployments
    autoplan:
      when_modified: ["services/user-service/**/*.tf"]
      enabled: true

  - name: user-service-staging
    dir: services/user-service
    workspace: staging
    execution_order_group: 2
    repo_locks:
      mode: on_plan
    plan_requirements: [mergeable]
    apply_requirements: [mergeable]

  - name: user-service-prod
    dir: services/user-service
    workspace: production
    execution_order_group: 3
    depends_on: ["core-service-mesh-prod"]
    repo_locks:
      mode: on_apply
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  # Pattern repeats for all microservices...
  # (In practice, you'd generate this config programmatically)

workflows:
  # Fast deployment for development
  dev-deploy:
    plan:
      steps:
        - init
        - plan
    apply:
      steps:
        - apply

  # Controlled deployment for staging
  staging-deploy:
    plan:
      steps:
        - init
        - run: ./scripts/staging-pre-plan.sh
        - plan
    apply:
      steps:
        - run: ./scripts/staging-pre-apply.sh
        - apply
        - run: ./scripts/staging-post-apply.sh

  # Safe deployment for production
  prod-deploy:
    plan:
      steps:
        - init
        - run: ./scripts/security-scan.sh
        - run: ./scripts/compliance-check.sh
        - plan
    apply:
      steps:
        - run: ./scripts/pre-prod-deploy.sh
        - apply
        - run: ./scripts/health-check.sh
        - run: ./scripts/post-prod-deploy.sh
```

Server configuration:
```yaml
# Maximum performance settings
parallel-pool-size: 100

enhanced-locking:
  enabled: true
  default-timeout: "10m"  # Shorter for faster turnover

  priority-queue:
    enabled: true
    max-queue-size: 5000
    queue-timeout: "5m"

  retry:
    enabled: true
    max-attempts: 2  # Fast failure for high throughput
    base-delay: "500ms"
    max-delay: "5s"

  deadlock-detection:
    enabled: false  # Disabled for maximum performance

  redis:
    cluster-mode: true
    key-prefix: "devops:lock:"
    lock-ttl: "15m"  # Shorter TTL for faster cleanup
```

## Troubleshooting Integration Issues

### Common Locking Problems and Solutions

#### Problem 1: Frequent Lock Conflicts

**Symptoms:**
```bash
# Error in Atlantis logs
ERROR: Lock already exists for project=api workspace=production
INFO: Lock is held by PR #123 from user @developer1
```

**Solution - Adjust Locking Strategy:**
```yaml
# Change from on_plan to on_apply for less contention
version: 3
projects:
  - name: high-contention-project
    dir: api
    workspace: production
    repo_locks:
      mode: on_apply  # Only lock during destructive operations

    # Enable shorter planning cycles
    autoplan:
      when_modified: ["api/**/*.tf"]
      enabled: true
```

#### Problem 2: Deadlocks in Complex Dependencies

**Symptoms:**
```bash
# Atlantis detects circular dependencies
ERROR: Deadlock detected between projects: api-prod, database-prod
INFO: Resolution policy: abort - terminating conflicting operations
```

**Solution - Restructure Dependencies:**
```yaml
version: 3
# Use consistent ordering to prevent deadlocks
projects:
  - name: database-prod
    dir: database
    workspace: production
    execution_order_group: 1  # Always first

  - name: cache-prod
    dir: cache
    workspace: production
    execution_order_group: 1  # Same group - can run in parallel

  - name: api-prod
    dir: api
    workspace: production
    execution_order_group: 2  # Always after data layer
    depends_on: ["database-prod", "cache-prod"]
```

#### Problem 3: Timeout Issues with Long-Running Operations

**Symptoms:**
```bash
# Lock timeout in enhanced locking
ERROR: Lock operation timed out after 30m0s
INFO: Consider increasing timeout or optimizing Terraform operations
```

**Solution - Configure Appropriate Timeouts:**
```yaml
version: 3
projects:
  - name: large-migration
    dir: database/migration
    workspace: production
    repo_locks:
      mode: on_apply
    workflow: long-running-migration

workflows:
  long-running-migration:
    apply:
      steps:
        - init
        - run: echo "This migration may take up to 2 hours"
        - run: ./scripts/pre-migration-checks.sh
        - apply  # Will use server-configured timeout
        - run: ./scripts/post-migration-validation.sh
```

Server configuration:
```yaml
enhanced-locking:
  default-timeout: "2h"  # Increase for long operations
  max-timeout: "4h"
```

### Performance Optimization Patterns

#### Pattern 1: Selective Locking

```yaml
version: 3
projects:
  # Critical infrastructure - strict locking
  - name: prod-database
    dir: infrastructure/database
    workspace: production
    repo_locks:
      mode: on_apply
    plan_requirements: [approved, mergeable]
    apply_requirements: [approved, mergeable]

  # Development resources - minimal locking
  - name: dev-database
    dir: infrastructure/database
    workspace: development
    repo_locks:
      mode: disabled  # No locking for development
    autoplan:
      enabled: true

  # Monitoring - non-critical locking
  - name: monitoring
    dir: monitoring
    workspace: production
    repo_locks:
      mode: on_plan  # Lighter locking for non-critical systems
```

#### Pattern 2: Batched Operations

```yaml
version: 3
# Group related changes to reduce lock overhead
projects:
  - name: microservices-batch-1
    dir: services/batch1
    workspace: production
    repo_locks:
      mode: on_plan
    autoplan:
      when_modified: [
        "services/user-service/**/*.tf",
        "services/auth-service/**/*.tf",
        "services/notification-service/**/*.tf"
      ]
    # Custom workflow to deploy multiple services atomically
    workflow: batch-deployment

workflows:
  batch-deployment:
    plan:
      steps:
        - init
        - run: ./scripts/batch-pre-plan.sh
        - plan
    apply:
      steps:
        - run: ./scripts/batch-pre-apply.sh
        - apply
        - run: ./scripts/batch-post-apply.sh
```

### Monitoring and Alerting

#### Lock Status Monitoring

```bash
# Script to monitor lock status
#!/bin/bash
# monitor-locks.sh

# Check for stuck locks (older than 1 hour)
atlantis server locks | jq '.[] | select(.time_locked | . < (now - 3600)) | {project: .project, workspace: .workspace, user: .user, time_locked: .time_locked}'

# Alert if too many locks are active
ACTIVE_LOCKS=$(atlantis server locks | jq '. | length')
if [ "$ACTIVE_LOCKS" -gt 10 ]; then
    echo "WARNING: High number of active locks: $ACTIVE_LOCKS"
    # Send alert to monitoring system
fi
```

#### Integration with Monitoring Systems

```yaml
# Prometheus monitoring integration
version: 3
projects:
  - name: monitoring-setup
    dir: monitoring
    workspace: production
    repo_locks:
      mode: on_apply
    workflow: monitoring-deployment

workflows:
  monitoring-deployment:
    apply:
      steps:
        - apply
        - run: |
            # Update Prometheus with lock metrics
            curl -X POST http://prometheus:9091/metrics/job/atlantis-locks \
              --data-binary "atlantis_locks_active{instance=\"$ATLANTIS_INSTANCE\"} $(atlantis server locks | jq '. | length')"
```

This integration guide provides practical, tested patterns for implementing Atlantis locking in real-world scenarios. The examples are designed to be copy-paste ready while being adaptable to specific organizational needs.