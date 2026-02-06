/**
 * Atlantis WebUI - Shared Alpine.js Components
 *
 * Factory functions for reusable component patterns.
 * Uses object spread to compose behavior into page-specific state functions.
 */

// ============================================================================
// UTILITY FUNCTIONS
// ============================================================================

/**
 * Format Unix timestamp as relative time (e.g., "5 minutes ago")
 * @param {number} unixSeconds - Unix timestamp in seconds
 * @returns {string} Human-readable relative time
 */
function formatRelativeTime(unixSeconds) {
  if (!unixSeconds) return 'unknown';

  const now = Math.floor(Date.now() / 1000);
  const diff = now - unixSeconds;

  if (diff < 60) return 'just now';
  if (diff < 3600) {
    const minutes = Math.floor(diff / 60);
    return minutes === 1 ? '1 min ago' : `${minutes} min ago`;
  }
  if (diff < 86400) {
    const hours = Math.floor(diff / 3600);
    return hours === 1 ? '1 hour ago' : `${hours} hours ago`;
  }
  const days = Math.floor(diff / 86400);
  return days === 1 ? '1 day ago' : `${days} days ago`;
}

/**
 * Re-initialize Lucide icons after DOM updates
 */
function refreshIcons() {
  if (window.lucide) {
    window.lucide.createIcons();
  }
}

// ============================================================================
// GROUPED TABLE FACTORY
// ============================================================================

/**
 * Factory function for grouped table state.
 * Use with object spread to compose into page-specific state functions.
 *
 * @example
 * function prListState() {
 *   return {
 *     ...createGroupedTableState({
 *       dataElementId: 'pr-data',
 *       htmxContainerId: 'pr-data-container',
 *       itemsKey: 'prs',
 *       searchFields: ['repo', 'pullNum', 'title']
 *     }),
 *     // Page-specific methods
 *     navigateToPR(pr) { window.location.href = pr.url; }
 *   };
 * }
 *
 * @param {Object} config - Configuration options
 * @param {string} config.dataElementId - ID of the script tag containing JSON data
 * @param {string} [config.htmxContainerId] - ID of the HTMX swap container (optional)
 * @param {string} [config.itemsKey='items'] - Key name for items in grouped result
 * @param {string[]} [config.searchFields=['repo']] - Fields to search when filtering
 * @param {boolean} [config.expandAllOnInit=true] - Whether to expand all groups on init
 * @param {boolean} [config.hasStatusFilters=false] - Whether to enable status filter support
 * @param {string} [config.groupKeyField='repo'] - Field to use for grouping items
 * @param {string} [config.groupDisplayField=null] - Field for display name (defaults to groupKeyField)
 * @returns {Object} Alpine.js component state object
 */
function createGroupedTableState(config) {
  const {
    dataElementId,
    htmxContainerId = null,
    itemsKey = 'items',
    searchFields = ['repo'],
    expandAllOnInit = true,
    hasStatusFilters = false,
    groupKeyField = 'repo',
    groupDisplayField = null
  } = config;

  const actualGroupDisplayField = groupDisplayField || groupKeyField;

  return {
    // ========== State ==========
    allItems: [],
    searchQuery: '',
    selectedRepo: '',
    expandedRepos: {},
    filteredItems: [],
    groupedItems: [],

    // Status filters (enabled via config.hasStatusFilters)
    statusFilters: {
      passed: false,
      failed: false,
      pending: false
    },

    // ========== Lifecycle ==========
    init() {
      this._loadData();
      this._updateGroups();

      if (expandAllOnInit) {
        this.expandAll();
      }

      this.$nextTick(() => refreshIcons());

      // Set up HTMX listener if container ID provided
      if (htmxContainerId) {
        this._setupHtmxListener();
      }
    },

    _setupHtmxListener() {
      document.body.addEventListener('htmx:afterSwap', (e) => {
        if (e.detail.target.id !== htmxContainerId) return;

        const savedState = { ...this.expandedRepos };

        this._loadData();
        this._updateGroups();
        this._restoreExpandState(savedState);

        this.$nextTick(() => refreshIcons());
      });
    },

    // ========== Data Loading ==========
    _loadData() {
      const el = document.getElementById(dataElementId);
      if (el) {
        try {
          this.allItems = JSON.parse(el.textContent);
        } catch (e) {
          console.error(`Failed to parse data from #${dataElementId}:`, e);
          this.allItems = [];
        }
      }
    },

    // ========== Filtering & Grouping ==========
    _updateGroups() {
      const query = this.searchQuery.toLowerCase().trim();

      // Filter items
      this.filteredItems = this.allItems.filter(item => {
        // Search filter - check all configured fields
        const matchesSearch = !query || searchFields.some(field => {
          const value = item[field];
          if (value === null || value === undefined) return false;
          return String(value).toLowerCase().includes(query);
        });

        // Repo filter
        const matchesRepo = !this.selectedRepo || item.repo === this.selectedRepo;

        // Status filter (only if enabled and filters are active)
        let matchesStatus = true;
        if (hasStatusFilters && this._hasActiveStatusFilters() && item.status) {
          matchesStatus =
            (this.statusFilters.passed && item.status === 'passed') ||
            (this.statusFilters.failed && item.status === 'failed') ||
            (this.statusFilters.pending && item.status === 'pending');
        }

        return matchesSearch && matchesRepo && matchesStatus;
      });

      // Group items by configured key field
      const groups = {};
      this.filteredItems.forEach(item => {
        const key = item[groupKeyField];
        const displayName = item[actualGroupDisplayField] || key;
        if (!groups[key]) {
          groups[key] = {
            key: key,
            displayName: displayName,
            [itemsKey]: []
          };
        }
        groups[key][itemsKey].push(item);
      });

      // Sort groups alphabetically by key
      this.groupedItems = Object.values(groups).sort((a, b) =>
        a.key.localeCompare(b.key)
      );
    },

    // ========== Status Filters ==========
    toggleStatusFilter(status) {
      this.statusFilters[status] = !this.statusFilters[status];
      this.applyFilters();
    },

    _hasActiveStatusFilters() {
      return this.statusFilters.passed ||
             this.statusFilters.failed ||
             this.statusFilters.pending;
    },

    hasActiveStatusFilters() {
      return this._hasActiveStatusFilters();
    },

    // ========== Expand/Collapse ==========
    isExpanded(key) {
      return this.expandedRepos[key] === true;
    },

    toggleGroup(key) {
      this.expandedRepos[key] = !this.expandedRepos[key];
      this.$nextTick(() => refreshIcons());
    },

    expandAll() {
      this.groupedItems.forEach(g => {
        this.expandedRepos[g.key] = true;
      });
      this.$nextTick(() => refreshIcons());
    },

    collapseAll() {
      this.expandedRepos = {};
      this.$nextTick(() => refreshIcons());
    },

    _restoreExpandState(savedState) {
      this.expandedRepos = {};
      this.groupedItems.forEach(g => {
        if (savedState[g.key]) {
          this.expandedRepos[g.key] = true;
        }
      });
      // If no saved state existed, expand all
      if (Object.keys(this.expandedRepos).length === 0 &&
          Object.keys(savedState).length === 0) {
        this.expandAll();
      }
    },

    // ========== Filter Actions ==========
    applyFilters() {
      this._updateGroups();
      this.$nextTick(() => refreshIcons());
    }
  };
}

// ============================================================================
// EXPORTS
// ============================================================================

window.createGroupedTableState = createGroupedTableState;
window.formatRelativeTime = formatRelativeTime;
window.refreshIcons = refreshIcons;
