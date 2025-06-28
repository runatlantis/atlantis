# Plan Queue API Documentation

This document describes the API endpoints for managing Atlantis plan queues. The queue system allows users to queue up plan requests when a project/workspace is already locked, providing better resource management and user experience.

## Base URL

All API endpoints are relative to your Atlantis server URL. For example, if Atlantis is running at `https://atlantis.example.com`, the API base URL would be `https://atlantis.example.com/api/queues`.

## Authentication

Currently, the queue API endpoints do not require authentication. However, this may change in future versions. It's recommended to secure your Atlantis instance appropriately.

## Endpoints

### 1. Get All Queues

Retrieves all active plan queues across all projects and workspaces.

**Endpoint:** `GET /api/queues`

**Response:**

```json
{
   "queues": [
      {
         "project": "owner/repo:path",
         "workspace": "default",
         "repo_full_name": "owner/repo",
         "entries": [
            {
               "id": "queue-entry-id",
               "pull_num": 123,
               "username": "user1",
               "time": "2024-01-15 10:30:00",
               "position": 1
            }
         ],
         "updated_at": "2024-01-15 10:30:00"
      }
   ],
   "count": 1
}
```

**Status Codes:**

-  `200 OK` - Successfully retrieved queues
-  `500 Internal Server Error` - Server error

### 2. Get Queue Status

Retrieves the current status of a specific queue for a project/workspace combination.

**Endpoint:** `GET /api/queues/{repo}/{project}/{workspace}`

**Parameters:**

-  `repo` (string, required) - Repository full name (e.g., "owner/repo")
-  `project` (string, required) - Project path (e.g., "terraform/prod")
-  `workspace` (string, required) - Workspace name (e.g., "default")

**Example:** `GET /api/queues/owner/repo/terraform/prod/default`

**Response:**

```json
{
   "project": "owner/repo:terraform/prod",
   "workspace": "default",
   "entries": [
      {
         "id": "queue-entry-id",
         "pull_num": 123,
         "username": "user1",
         "time": "2024-01-15 10:30:00",
         "position": 1
      }
   ],
   "updated_at": "2024-01-15 10:30:00"
}
```

**Status Codes:**

-  `200 OK` - Successfully retrieved queue status
-  `404 Not Found` - Queue not found
-  `500 Internal Server Error` - Server error

### 3. Remove from Queue

Removes a specific pull request from a queue.

**Endpoint:** `DELETE /api/queues/{repo}/{project}/{workspace}/{pull_num}`

**Parameters:**

-  `repo` (string, required) - Repository full name (e.g., "owner/repo")
-  `project` (string, required) - Project path (e.g., "terraform/prod")
-  `workspace` (string, required) - Workspace name (e.g., "default")
-  `pull_num` (integer, required) - Pull request number

**Example:** `DELETE /api/queues/owner/repo/terraform/prod/default/123`

**Response:**

```json
{
   "message": "Successfully removed from queue"
}
```

**Status Codes:**

-  `200 OK` - Successfully removed from queue
-  `400 Bad Request` - Invalid pull number
-  `500 Internal Server Error` - Server error

## Data Models

### Queue Entry

Represents a single entry in a plan queue.

```json
{
   "id": "string", // Unique identifier for the queue entry
   "pull_num": 123, // Pull request number
   "username": "string", // Username of the person who queued the request
   "time": "string", // ISO 8601 formatted timestamp when entry was added
   "position": 1 // Position in the queue (1-based)
}
```

### Queue

Represents a complete queue for a project/workspace combination.

```json
{
   "project": "string", // Project identifier (repo:path format)
   "workspace": "string", // Workspace name
   "repo_full_name": "string", // Repository full name
   "entries": [], // Array of queue entries
   "updated_at": "string" // ISO 8601 formatted timestamp of last update
}
```

## Error Responses

All endpoints may return error responses in the following format:

```json
{
   "error": "Error message describing what went wrong"
}
```

## Usage Examples

### Using curl

```bash
# Get all queues
curl -X GET "https://atlantis.example.com/api/queues"

# Get specific queue status
curl -X GET "https://atlantis.example.com/api/queues/owner/repo/terraform/prod/default"

# Remove from queue
curl -X DELETE "https://atlantis.example.com/api/queues/owner/repo/terraform/prod/default/123"
```

### Using JavaScript

```javascript
// Get all queues
fetch("/api/queues")
   .then((response) => response.json())
   .then((data) => {
      console.log("Queues:", data.queues);
   });

// Remove from queue
fetch("/api/queues/owner/repo/terraform/prod/default/123", {
   method: "DELETE",
})
   .then((response) => response.json())
   .then((data) => {
      console.log("Result:", data.message);
   });
```

## Web Interface

In addition to the API, Atlantis provides a web interface for managing queues:

-  **Queue Overview:** `https://atlantis.example.com/queues` - View all active queues
-  **API Documentation:** `https://atlantis.example.com/api/queues` - Access the API directly

## Integration Notes

1. **Queue Position Updates:** When entries are added or removed from a queue, the positions of remaining entries may change.

2. **Concurrent Access:** The API is designed to handle concurrent requests safely.

3. **Queue Cleanup:** Queues are automatically cleaned up when pull requests are closed or merged.

4. **Rate Limiting:** Consider implementing appropriate rate limiting for production use.

## Future Enhancements

The following features may be added in future versions:

-  Authentication and authorization
-  Queue priority management
-  Queue entry metadata
-  Webhook notifications for queue changes
-  Queue analytics and metrics
