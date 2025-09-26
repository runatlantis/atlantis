#!/bin/bash
# Enhanced Locking System Rollback Script
# This script provides comprehensive rollback capabilities for the enhanced locking system

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(dirname "$(dirname "$SCRIPT_DIR")")")
CONFIG_DIR="${ROOT_DIR}/config"
BACKUP_DIR="${ROOT_DIR}/backups"
LOG_DIR="${ROOT_DIR}/logs"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1" | tee -a "${LOG_DIR}/rollback.log"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1" | tee -a "${LOG_DIR}/rollback.log"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" | tee -a "${LOG_DIR}/rollback.log"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1" | tee -a "${LOG_DIR}/rollback.log"
}

# Rollback types
ROLLBACK_TYPES=(
    "immediate"    # < 5 minutes - configuration rollback
    "extended"     # < 30 minutes - full application rollback
    "emergency"    # < 2 minutes - kill switch activation
    "selective"    # Rollback specific features only
)

# Usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS] [TYPE]

Rollback the enhanced locking system deployment.

ROLLBACK TYPES:
    immediate   - Quick configuration rollback (< 5 minutes)
    extended    - Full application rollback with cleanup (< 30 minutes)
    emergency   - Emergency kill switch activation (< 2 minutes)
    selective   - Rollback specific features only

OPTIONS:
    -h, --help          Show this help message
    -d, --dry-run       Show what would be rolled back without making changes
    -f, --force         Skip confirmation prompts
    -v, --verbose       Enable verbose output
    -b, --backup        Specific backup to rollback to (timestamp or path)
    --preserve-data     Preserve Redis data during rollback
    --preserve-config   Preserve enhanced config structure
    --disable-only      Only disable features without removing configuration
    --feature           Specific feature to rollback (for selective type)

FEATURES (for selective rollback):
    redis-backend
    priority-queue
    deadlock-detection
    queue-monitoring
    event-streaming

EXAMPLES:
    $0 immediate                           # Quick configuration rollback
    $0 emergency                          # Emergency kill switch
    $0 selective --feature priority-queue # Disable only priority queue
    $0 extended --backup 20231201_143022  # Rollback to specific backup
    $0 --dry-run extended                 # Show what would be done

EOF
}

# Parse command line arguments
parse_args() {
    ROLLBACK_TYPE=""
    DRY_RUN=false
    FORCE=false
    VERBOSE=false
    SPECIFIC_BACKUP=""
    PRESERVE_DATA=false
    PRESERVE_CONFIG=false
    DISABLE_ONLY=false
    SPECIFIC_FEATURE=""

    while [[ $# -gt 0 ]]; do
        case $1 in
            -h|--help)
                show_usage
                exit 0
                ;;
            -d|--dry-run)
                DRY_RUN=true
                shift
                ;;
            -f|--force)
                FORCE=true
                shift
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -b|--backup)
                SPECIFIC_BACKUP="$2"
                shift 2
                ;;
            --preserve-data)
                PRESERVE_DATA=true
                shift
                ;;
            --preserve-config)
                PRESERVE_CONFIG=true
                shift
                ;;
            --disable-only)
                DISABLE_ONLY=true
                shift
                ;;
            --feature)
                SPECIFIC_FEATURE="$2"
                shift 2
                ;;
            -*)
                log_error "Unknown option: $1"
                show_usage
                exit 1
                ;;
            *)
                if [[ -z "$ROLLBACK_TYPE" ]]; then
                    ROLLBACK_TYPE="$1"
                else
                    log_error "Too many arguments"
                    show_usage
                    exit 1
                fi
                shift
                ;;
        esac
    done

    # Validate rollback type
    if [[ -n "$ROLLBACK_TYPE" ]]; then
        if [[ ! " ${ROLLBACK_TYPES[*]} " =~ " ${ROLLBACK_TYPE} " ]]; then
            log_error "Invalid rollback type: $ROLLBACK_TYPE"
            log_info "Valid types: ${ROLLBACK_TYPES[*]}"
            exit 1
        fi
    fi

    # Validate selective rollback
    if [[ "$ROLLBACK_TYPE" == "selective" && -z "$SPECIFIC_FEATURE" ]]; then
        log_error "Selective rollback requires --feature option"
        show_usage
        exit 1
    fi
}

# Initialize logging
init_logging() {
    mkdir -p "$LOG_DIR"
    
    local log_file="${LOG_DIR}/rollback.log"
    echo "================================" >> "$log_file"
    echo "Rollback started: $(date)" >> "$log_file"
    echo "Type: $ROLLBACK_TYPE" >> "$log_file"
    echo "User: $(whoami)" >> "$log_file"
    echo "================================" >> "$log_file"
}

# Check current system state
check_system_state() {
    log_info "Checking current system state..."
    
    # Check if Atlantis is running
    if ! curl -sf http://localhost:4141/healthz &>/dev/null; then
        log_warn "Atlantis appears to be down or not responding"
        if [[ "$FORCE" == "false" ]]; then
            read -p "Continue with rollback anyway? (y/N): " -r
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                exit 1
            fi
        fi
    else
        log_success "Atlantis is responding"
    fi
    
    # Check enhanced locking status
    if curl -sf http://localhost:4141/healthz | jq -e '.enhanced_locking.enabled == true' &>/dev/null; then
        log_info "Enhanced locking is currently enabled"
    else
        log_info "Enhanced locking is currently disabled"
    fi
    
    # Check for active locks
    local active_locks
    active_locks=$(curl -sf http://localhost:4141/api/locks | jq length 2>/dev/null || echo "0")
    if [[ $active_locks -gt 0 ]]; then
        log_warn "There are $active_locks active locks in the system"
        if [[ "$FORCE" == "false" ]]; then
            read -p "Proceed with rollback while locks are active? (y/N): " -r
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                exit 1
            fi
        fi
    fi
}

# Find available backups
find_backups() {
    log_info "Scanning for available backups..."
    
    if [[ ! -d "$BACKUP_DIR" ]]; then
        log_error "Backup directory not found: $BACKUP_DIR"
        return 1
    fi
    
    local backups=()
    while IFS= read -r -d '' backup; do
        backups+=("$(basename "$backup")")
    done < <(find "$BACKUP_DIR" -type d -name "backup_*" -print0 | sort -rz)
    
    if [[ ${#backups[@]} -eq 0 ]]; then
        log_error "No backups found in $BACKUP_DIR"
        return 1
    fi
    
    log_info "Available backups:"
    for i in "${!backups[@]}"; do
        local backup_path="${BACKUP_DIR}/${backups[i]}"
        local backup_date
        if [[ -f "${backup_path}/metadata.txt" ]]; then
            backup_date=$(grep "Backup created:" "${backup_path}/metadata.txt" | cut -d: -f2- | xargs)
        else
            backup_date="Unknown"
        fi
        printf "  %2d. %s (%s)\n" $((i+1)) "${backups[i]}" "$backup_date"
    done
    
    # Set latest backup as default if not specified
    if [[ -z "$SPECIFIC_BACKUP" ]]; then
        SPECIFIC_BACKUP="${backups[0]}"
        log_info "Using latest backup: $SPECIFIC_BACKUP"
    fi
}

# Validate backup
validate_backup() {
    local backup_path
    
    # Handle backup specification (timestamp or full path)
    if [[ "$SPECIFIC_BACKUP" =~ ^/ ]]; then
        backup_path="$SPECIFIC_BACKUP"
    else
        backup_path="${BACKUP_DIR}/backup_${SPECIFIC_BACKUP}"
        if [[ ! -d "$backup_path" ]]; then
            backup_path="${BACKUP_DIR}/${SPECIFIC_BACKUP}"
        fi
    fi
    
    if [[ ! -d "$backup_path" ]]; then
        log_error "Backup not found: $backup_path"
        exit 1
    fi
    
    log_info "Validating backup: $backup_path"
    
    # Check backup contents
    local has_config=false
    local has_db=false
    
    if [[ -f "${backup_path}/config.yaml.bak" ]]; then
        has_config=true
        log_info "  ✓ Configuration backup found"
    fi
    
    if [[ -f "${backup_path}/atlantis.db.bak" ]]; then
        has_db=true
        log_info "  ✓ Database backup found"
    fi
    
    if [[ "$has_config" == "false" && "$has_db" == "false" ]]; then
        log_error "Backup contains no restorable files"
        exit 1
    fi
    
    # Display backup metadata
    if [[ -f "${backup_path}/metadata.txt" ]]; then
        log_info "Backup metadata:"
        while IFS= read -r line; do
            log_info "  $line"
        done < "${backup_path}/metadata.txt"
    fi
    
    echo "$backup_path"  # Return backup path
}

# Immediate rollback (< 5 minutes)
perform_immediate_rollback() {
    log_info "Performing immediate rollback..."
    
    local start_time
    start_time=$(date +%s)
    
    # Step 1: Disable enhanced locking via configuration
    disable_enhanced_locking
    
    # Step 2: Restart services to pick up configuration change
    restart_services_quick
    
    # Step 3: Verify fallback to legacy system
    verify_legacy_system
    
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_success "Immediate rollback completed in ${duration}s"
    
    if [[ $duration -gt 300 ]]; then  # 5 minutes
        log_warn "Rollback took longer than expected (${duration}s > 300s)"
    fi
}

# Extended rollback (< 30 minutes)
perform_extended_rollback() {
    log_info "Performing extended rollback..."
    
    local start_time
    start_time=$(date +%s)
    
    # Step 1: Find and validate backup
    find_backups
    local backup_path
    backup_path=$(validate_backup)
    
    # Step 2: Stop services
    stop_services
    
    # Step 3: Restore configuration
    restore_configuration "$backup_path"
    
    # Step 4: Restore database if needed
    restore_database "$backup_path"
    
    # Step 5: Clean up Redis state if not preserving data
    if [[ "$PRESERVE_DATA" == "false" ]]; then
        cleanup_redis_state
    fi
    
    # Step 6: Restart services
    restart_services_full
    
    # Step 7: Verify system health
    verify_system_health
    
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_success "Extended rollback completed in ${duration}s"
    
    if [[ $duration -gt 1800 ]]; then  # 30 minutes
        log_warn "Rollback took longer than expected (${duration}s > 1800s)"
    fi
}

# Emergency rollback (< 2 minutes)
perform_emergency_rollback() {
    log_info "Performing emergency rollback..."
    
    local start_time
    start_time=$(date +%s)
    
    # Step 1: Activate kill switch
    activate_kill_switch
    
    # Step 2: Force traffic to legacy backend
    force_legacy_backend
    
    # Step 3: Alert operations team
    send_emergency_alert
    
    local end_time
    end_time=$(date +%s)
    local duration=$((end_time - start_time))
    
    log_success "Emergency rollback completed in ${duration}s"
    
    if [[ $duration -gt 120 ]]; then  # 2 minutes
        log_error "Emergency rollback took too long (${duration}s > 120s)"
    fi
}

# Selective rollback
perform_selective_rollback() {
    log_info "Performing selective rollback for feature: $SPECIFIC_FEATURE"
    
    case "$SPECIFIC_FEATURE" in
        "redis-backend")
            disable_redis_backend
            ;;
        "priority-queue")
            disable_priority_queue
            ;;
        "deadlock-detection")
            disable_deadlock_detection
            ;;
        "queue-monitoring")
            disable_queue_monitoring
            ;;
        "event-streaming")
            disable_event_streaming
            ;;
        *)
            log_error "Unknown feature: $SPECIFIC_FEATURE"
            exit 1
            ;;
    esac
    
    # Restart services to pick up changes
    restart_services_quick
    
    # Verify feature is disabled
    verify_feature_disabled "$SPECIFIC_FEATURE"
    
    log_success "Selective rollback completed for feature: $SPECIFIC_FEATURE"
}

# Core rollback functions

disable_enhanced_locking() {
    log_info "Disabling enhanced locking in configuration..."
    
    local config_file="${CONFIG_DIR}/enhanced-locking.yaml"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would disable enhanced locking in $config_file"
        return
    fi
    
    # Create backup of current config before modifying
    cp "$config_file" "${config_file}.pre-rollback"
    
    # Disable enhanced locking
    sed -i 's/enabled: true/enabled: false/' "$config_file"
    
    # Set backend to BoltDB
    sed -i 's/backend: "redis"/backend: "boltdb"/' "$config_file"
    
    log_success "Enhanced locking disabled in configuration"
}

activate_kill_switch() {
    log_info "Activating emergency kill switch..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would activate kill switch"
        return
    fi
    
    # Set environment variable to disable enhanced locking immediately
    export DISABLE_ENHANCED_LOCKING=true
    
    # Write to environment file if it exists
    local env_file="${ROOT_DIR}/.env"
    if [[ -f "$env_file" ]]; then
        echo "DISABLE_ENHANCED_LOCKING=true" >> "$env_file"
    fi
    
    # Signal running processes to disable enhanced locking
    if pgrep atlantis &>/dev/null; then
        log_info "Sending signal to running Atlantis processes..."
        pkill -USR1 atlantis || true
    fi
    
    log_success "Emergency kill switch activated"
}

force_legacy_backend() {
    log_info "Forcing all operations to legacy backend..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would force legacy backend"
        return
    fi
    
    # Create circuit breaker file
    touch "${ROOT_DIR}/.circuit-breaker-active"
    
    # Set environment variables
    export FORCE_LEGACY_BACKEND=true
    export ENHANCED_LOCKING_CIRCUIT_BREAKER=true
    
    log_success "Legacy backend forced for all operations"
}

# Service management functions

stop_services() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would stop Atlantis services"
        return
    fi
    
    log_info "Stopping Atlantis services..."
    
    # Check if running with Docker Compose
    if [[ -f "${ROOT_DIR}/docker-compose.yml" ]]; then
        docker-compose -f "${ROOT_DIR}/docker-compose.yml" stop atlantis
    elif command -v systemctl &> /dev/null; then
        sudo systemctl stop atlantis
    else
        # Try to stop via PID file
        local pid_file="${ROOT_DIR}/atlantis.pid"
        if [[ -f "$pid_file" ]]; then
            kill "$(cat "$pid_file")" || true
            rm -f "$pid_file"
        fi
    fi
    
    # Wait for services to stop
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if ! curl -sf http://localhost:4141/healthz &>/dev/null; then
            log_success "Services stopped successfully"
            return 0
        fi
        
        log_info "Waiting for services to stop... (attempt $attempt/$max_attempts)"
        sleep 2
        ((attempt++))
    done
    
    log_warn "Services may still be running after $max_attempts attempts"
}

restart_services_quick() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would restart Atlantis services (quick)"
        return
    fi
    
    log_info "Restarting Atlantis services (quick mode)..."
    
    # Quick restart without full stop
    if [[ -f "${ROOT_DIR}/docker-compose.yml" ]]; then
        docker-compose -f "${ROOT_DIR}/docker-compose.yml" restart atlantis
    elif command -v systemctl &> /dev/null; then
        sudo systemctl restart atlantis
    else
        log_warn "Unable to automatically restart services. Please restart Atlantis manually."
        if [[ "$FORCE" == "false" ]]; then
            read -p "Press Enter after restarting Atlantis..."
        fi
    fi
    
    # Quick health check
    sleep 10
    if curl -sf http://localhost:4141/healthz &>/dev/null; then
        log_success "Services restarted successfully"
    else
        log_warn "Services may not be fully ready yet"
    fi
}

restart_services_full() {
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would restart Atlantis services (full)"
        return
    fi
    
    log_info "Starting Atlantis services..."
    
    # Start services
    if [[ -f "${ROOT_DIR}/docker-compose.yml" ]]; then
        docker-compose -f "${ROOT_DIR}/docker-compose.yml" up -d atlantis
    elif command -v systemctl &> /dev/null; then
        sudo systemctl start atlantis
    else
        log_warn "Unable to automatically start services. Please start Atlantis manually."
        if [[ "$FORCE" == "false" ]]; then
            read -p "Press Enter after starting Atlantis..."
        fi
    fi
    
    # Wait for service to be ready
    wait_for_service_ready
    
    log_success "Services started successfully"
}

wait_for_service_ready() {
    log_info "Waiting for Atlantis to be ready..."
    
    local max_attempts=60  # 5 minutes
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -sf http://localhost:4141/healthz &>/dev/null; then
            log_success "Atlantis is ready"
            return 0
        fi
        
        log_info "Waiting for Atlantis... (attempt $attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    log_error "Atlantis failed to become ready after $max_attempts attempts"
    return 1
}

# Restoration functions

restore_configuration() {
    local backup_path="$1"
    local config_backup="${backup_path}/config.yaml.bak"
    
    if [[ ! -f "$config_backup" ]]; then
        log_warn "No configuration backup found in $backup_path"
        return
    fi
    
    log_info "Restoring configuration from backup..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would restore config from $config_backup"
        return
    fi
    
    local config_file="${CONFIG_DIR}/enhanced-locking.yaml"
    
    # Backup current config before restoring
    cp "$config_file" "${config_file}.pre-restore"
    
    # Restore configuration
    cp "$config_backup" "$config_file"
    
    log_success "Configuration restored from backup"
}

restore_database() {
    local backup_path="$1"
    local db_backup="${backup_path}/atlantis.db.bak"
    
    if [[ ! -f "$db_backup" ]]; then
        log_info "No database backup found in $backup_path"
        return
    fi
    
    log_info "Restoring BoltDB database from backup..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would restore database from $db_backup"
        return
    fi
    
    local db_file="${ROOT_DIR}/atlantis.db"
    
    # Backup current database before restoring
    if [[ -f "$db_file" ]]; then
        cp "$db_file" "${db_file}.pre-restore"
    fi
    
    # Restore database
    cp "$db_backup" "$db_file"
    
    log_success "Database restored from backup"
}

cleanup_redis_state() {
    log_info "Cleaning up Redis state..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would cleanup Redis state"
        return
    fi
    
    # Find Redis connection info from config
    local config_file="${CONFIG_DIR}/enhanced-locking.yaml"
    local redis_host="localhost"
    local redis_port="6379"
    
    if [[ -f "$config_file" ]]; then
        redis_host=$(grep -A10 "redis:" "$config_file" | grep "addresses:" -A1 | tail -1 | sed 's/.*- "\([^:]*\):.*/\1/' || echo "localhost")
        redis_port=$(grep -A10 "redis:" "$config_file" | grep "addresses:" -A1 | tail -1 | sed 's/.*:\([0-9]*\)".*/\1/' || echo "6379")
    fi
    
    # Clean up Redis keys related to enhanced locking
    if command -v redis-cli &> /dev/null; then
        log_info "Removing enhanced locking keys from Redis..."
        redis-cli -h "$redis_host" -p "$redis_port" --scan --pattern "atlantis:enhanced:*" | xargs -r redis-cli -h "$redis_host" -p "$redis_port" del
        log_success "Redis state cleaned up"
    else
        log_warn "redis-cli not available, skipping Redis cleanup"
    fi
}

# Feature-specific disable functions

disable_redis_backend() {
    log_info "Disabling Redis backend..."
    
    local config_file="${CONFIG_DIR}/enhanced-locking.yaml"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would disable Redis backend in $config_file"
        return
    fi
    
    sed -i 's/backend: "redis"/backend: "boltdb"/' "$config_file"
    
    log_success "Redis backend disabled"
}

disable_priority_queue() {
    log_info "Disabling priority queue..."
    
    local config_file="${CONFIG_DIR}/enhanced-locking.yaml"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would disable priority queue in $config_file"
        return
    fi
    
    sed -i 's/priority-queue: true/priority-queue: false/' "$config_file"
    sed -i 's/retry-mechanism: true/retry-mechanism: false/' "$config_file"
    
    log_success "Priority queue disabled"
}

disable_deadlock_detection() {
    log_info "Disabling deadlock detection..."
    
    local config_file="${CONFIG_DIR}/enhanced-locking.yaml"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would disable deadlock detection in $config_file"
        return
    fi
    
    sed -i 's/deadlock-detection: true/deadlock-detection: false/' "$config_file"
    
    log_success "Deadlock detection disabled"
}

disable_queue_monitoring() {
    log_info "Disabling queue monitoring..."
    
    local config_file="${CONFIG_DIR}/enhanced-locking.yaml"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would disable queue monitoring in $config_file"
        return
    fi
    
    sed -i 's/queue-monitoring: true/queue-monitoring: false/' "$config_file"
    
    log_success "Queue monitoring disabled"
}

disable_event_streaming() {
    log_info "Disabling event streaming..."
    
    local config_file="${CONFIG_DIR}/enhanced-locking.yaml"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        log_info "[DRY RUN] Would disable event streaming in $config_file"
        return
    fi
    
    sed -i 's/event-streaming: true/event-streaming: false/' "$config_file"
    
    log_success "Event streaming disabled"
}

# Verification functions

verify_legacy_system() {
    log_info "Verifying fallback to legacy system..."
    
    # Check health endpoint
    local max_attempts=30
    local attempt=1
    
    while [[ $attempt -le $max_attempts ]]; do
        if curl -sf http://localhost:4141/healthz | jq -e '.enhanced_locking.enabled == false' &>/dev/null; then
            log_success "Legacy system is active"
            return 0
        fi
        
        log_info "Waiting for legacy system... (attempt $attempt/$max_attempts)"
        sleep 5
        ((attempt++))
    done
    
    log_error "Failed to verify legacy system activation"
    return 1
}

verify_system_health() {
    log_info "Verifying overall system health..."
    
    # Check basic health
    if ! curl -sf http://localhost:4141/healthz | jq -e '.status == "healthy"' &>/dev/null; then
        log_error "System health check failed"
        return 1
    fi
    
    # Test basic lock operation
    # This would require implementing a test lock operation
    
    log_success "System health verified"
}

verify_feature_disabled() {
    local feature="$1"
    
    log_info "Verifying $feature is disabled..."
    
    # Check specific feature status via health endpoint
    case "$feature" in
        "redis-backend")
            if curl -sf http://localhost:4141/healthz | jq -e '.enhanced_locking.backend == "boltdb"' &>/dev/null; then
                log_success "Redis backend disabled successfully"
            else
                log_error "Redis backend still active"
                return 1
            fi
            ;;
        "priority-queue")
            if curl -sf http://localhost:4141/healthz | jq -e '.enhanced_locking.features.priority_queue == false' &>/dev/null; then
                log_success "Priority queue disabled successfully"
            else
                log_error "Priority queue still active"
                return 1
            fi
            ;;
        # Add other feature verifications as needed
    esac
}

# Alert functions

send_emergency_alert() {
    log_info "Sending emergency alert..."
    
    local alert_message="EMERGENCY: Enhanced locking system has been disabled due to critical issues. System has been rolled back to legacy BoltDB backend."
    
    # Send to various alert channels
    # Slack, email, PagerDuty, etc. - customize based on your infrastructure
    
    # Example: Send to Slack webhook
    local slack_webhook="${SLACK_WEBHOOK_URL:-}"
    if [[ -n "$slack_webhook" ]]; then
        curl -X POST -H 'Content-type: application/json' \
            --data "{\"text\":\"$alert_message\"}" \
            "$slack_webhook" || log_warn "Failed to send Slack alert"
    fi
    
    # Example: Write to syslog
    logger -p user.crit "$alert_message"
    
    log_info "Emergency alert sent"
}

# Main execution
main() {
    parse_args "$@"
    
    # Initialize logging
    init_logging
    
    # Show header
    echo "========================================="
    echo "Enhanced Locking System Rollback Script"
    echo "========================================="
    echo
    
    if [[ -z "$ROLLBACK_TYPE" ]]; then
        log_error "No rollback type specified"
        show_usage
        exit 1
    fi
    
    # Check system state
    check_system_state
    
    # Confirm rollback unless forced
    if [[ "$FORCE" == "false" && "$DRY_RUN" == "false" ]]; then
        echo
        log_warn "This will perform a $ROLLBACK_TYPE rollback of the enhanced locking system."
        read -p "Are you sure you want to proceed? (yes/NO): " -r
        if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
            log_info "Rollback cancelled by user"
            exit 0
        fi
    fi
    
    # Perform rollback based on type
    case "$ROLLBACK_TYPE" in
        "immediate")
            perform_immediate_rollback
            ;;
        "extended")
            perform_extended_rollback
            ;;
        "emergency")
            perform_emergency_rollback
            ;;
        "selective")
            perform_selective_rollback
            ;;
        *)
            log_error "Unknown rollback type: $ROLLBACK_TYPE"
            exit 1
            ;;
    esac
    
    # Final verification
    if [[ "$ROLLBACK_TYPE" != "emergency" ]]; then
        verify_system_health
    fi
    
    echo
    log_success "Rollback completed successfully!"
    log_info "Please monitor the system closely and check logs for any issues."
    
    if [[ "$ROLLBACK_TYPE" == "emergency" ]]; then
        log_warn "Emergency rollback completed. Manual intervention may be required."
        log_info "Check system logs and contact support if issues persist."
    fi
}

# Run main function
main "$@"
