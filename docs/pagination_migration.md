# Pagination Migration Guide

## Overview
This document provides migration examples for transitioning from legacy collection queries to the new paginated implementations.

## Filters Module Migration

### Before (Legacy Implementation)
```go
// Get all filters (memory intensive)
filters := db.GetAllFilters(chatID)

// Process all filters at once
for _, filter := range filters {
    // ... process filter
}
```

### After (Paginated Implementation)
```go 
// Option 1: Simple pagination (auto chunking)
result, err := db.GetAllFiltersPaginated(chatID, db.PaginationOptions{
    Limit: 100, // Process 100 at a time
})
if err != nil {
    // handle error
}
filters := result.Data.([]*db.ChatFilters)

// Option 2: Manual cursor pagination 
var cursor interface{}
for {
    result, err := db.GetAllFiltersPaginated(chatID, db.PaginationOptions{
        Cursor: cursor,
        Limit:  100,
    })
    if err != nil || len(result.Data) == 0 {
        break
    }
    
    filters := result.Data.([]*db.ChatFilters)
    cursor = result.NextCursor
    
    // Process batch
    for _, filter := range filters {
        // ... process filter
    }
}
```

## Notes Module Migration

### Before (Legacy Implementation)
```go
notes := db.getAllChatNotes(chatID)
// Process all notes at once
```

### After (Paginated Implementation)
```go
// Process notes in batches
result, err := db.GetAllNotesPaginated(chatID, db.PaginationOptions{
    Limit: 100, 
})
if err != nil {
    // handle error
}
notes := result.Data

// Or with cursor pagination:
var cursor interface{}
for {
    result, err := db.GetAllNotesPaginated(chatID, db.PaginationOptions{
        Cursor: cursor,
        Limit:  100,
    })
    // ... similar to filters example
}
```

## Performance Considerations

1. **Batch Size**: 
   - Recommended: 100-500 documents per page
   - Adjust based on document size and memory constraints

2. **Cursor vs Offset**:
   - Use cursor for large datasets (>10k docs)
   - Offset works well for smaller datasets

3. **Error Handling**:
   - Always check pagination result errors
   - Handle empty result sets gracefully

## Backward Compatibility
- Legacy methods remain available but are deprecated
- New code should use paginated versions