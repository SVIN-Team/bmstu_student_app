# HTTP Handlers

This directory contains HTTP handlers for the StudentHub API.

## Structure

- **auth.go** - Authentication handlers (SignUp, SignIn, Refresh, SignOut, SignOutAll)
- **schedule.go** - Schedule/Lessons handlers (GetLessons, GetLessonByID)
- **queue.go** - Queue management handlers (CRUD operations for queues)
- **queue_slots.go** - Queue slots handlers (Sign-up, cancellation, slot updates)
- **user.go** - User profile handlers (GetCurrentUser, UpdateCurrentUser, GetCurrentUserSlots, TransferHeadmanRole)
- **admin.go** - Admin handlers (User management, Groups CRUD)

## DTOs

Data Transfer Objects are located in the `dto/` subdirectory:

- **common.go** - Common response types (GroupResponse, SubjectResponse, etc.)
- **login.go** - Login request DTO
- **signup.go** - Sign-up request DTO
- **queue.go** - Queue-related DTOs
- **slot.go** - Slot-related DTOs
- **lesson.go** - Lesson-related DTOs
- **user.go** - User-related DTOs

## Handler Responsibilities

Each handler is responsible for:
1. Request validation (parsing parameters, body)
2. Calling the appropriate use case
3. Error handling and mapping to HTTP status codes
4. Response formatting

## Error Codes

All error responses follow the format:
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message"
  }
}
```

Common error codes:
- `VALIDATION_ERROR` (400) - Invalid request parameters
- `UNAUTHORIZED` (401) - Missing or invalid authentication
- `FORBIDDEN` (403) - Insufficient permissions
- `NOT_FOUND` (404) - Resource not found
- `ALREADY_EXISTS` (409) - Resource already exists
- `QUEUE_FULL` (409) - Queue is at capacity
- `QUEUE_CLOSED` (409) - Queue is not open
- `INTERNAL_ERROR` (500) - Server error

## Success Responses

All success responses follow the format:
```json
{
  "success": true,
  "data": { ... }
}
```

## Authentication

Most handlers require authentication via JWT token in the Authorization header:
```
Authorization: Bearer <access_token>
```

The token is validated by middleware, which sets the following context values:
- `user_id` (uuid.UUID) - Authenticated user's ID
- `user_role` (string) - User's role
- `user_group_id` (uuid.UUID) - User's group ID (if applicable)

## Dependencies

**IMPORTANT:** Handlers depend ONLY on use case interfaces, NOT repositories. This follows Clean Architecture principles:
- Easy testing with mocks
- Flexibility in implementation
- Loose coupling between layers
- Proper separation of concerns

## Next Steps

To use these handlers:
1. Implement the corresponding use case interfaces
2. Wire up handlers in the router configuration
3. Add authentication middleware
4. Add role-based authorization middleware
