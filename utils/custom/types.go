package custom

// custom types used throughout the application

// RequestContextKey type is used with request context value key
type RequestContextKey string

const OrganizationIDKey RequestContextKey = "org-id"
const UserRoleKey RequestContextKey = "user-role"