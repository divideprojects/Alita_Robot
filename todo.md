- [ ] **Deduplicate database CRUD logic across *_db.go files**  
  *Duplicate*: Every domain‐specific database file (e.g. `filters_db.go`, `notes_db.go`, `blacklists_db.go`, `disable_db.go`, …) re-implements the same CRUD helpers: `GetX`, `AddX`, `RemoveX`, `DoesXExist`, `LoadXStats`, etc. The only thing that changes is the collection name and the concrete struct.  
  *Fix*: Introduce a generic repository pattern or use Go 1.18+ generics so that one set of typed helper functions can serve any struct that embeds common fields (ChatID, Keyword/Name, etc.). Consolidate stats helpers into a single function that accepts the collection as a parameter.  
  *Reason*: Centralising this logic cuts 2000+ LOC, eliminates copy-paste bugs (e.g. forgotten error handling), and makes future schema changes (like switching DB driver) one-touch.

- [ ] **Merge Filters and Notes modules into a single “Keyed-Response” engine**  
  *Duplicate*: `alita/modules/filters.go` (~500 LOC) and `alita/modules/notes.go` (~730 LOC) share ~70 % identical code: argument parsing, overwrite-confirmation, CRUD calls, button handlers, list & purge commands, watchers, etc.  
  *Fix*: Extract a generic handler that maps a keyword to a stored response. Keep a thin wrapper that sets module-specific defaults (rate-limits, max items, etc.). Reuse the same overwrite map & callback logic.  
  *Reason*: Removes ~900 duplicate lines, slashes maintenance effort, and ensures feature parity between filters and notes automatically.

- [ ] **Abstract permission / connection checks into reusable decorator**  
  *Duplicate*: Almost every command starts with identical blocks: `connectedChat := helpers.IsUserConnected(...)`, `chat_status.CanUserChangeInfo(...)`, argument length sanity checks, etc.  
  *Fix*: Provide a middleware/decorator that injects the resolved chat, validates admin rights, and short-circuits on failure. Commands then become small business-logic handlers.  
  *Reason*: Reduces noise in command files, enforces consistent access control, and lowers chance of missing a permission check in new code.

- [ ] **Unify message extraction helpers (`GetNoteAndFilterType`, etc.)**  
  *Duplicate*: Filters and Notes each call slightly different versions of the same parsing routine tucked inside `utils/helpers`.  
  *Fix*: Create one canonical extractor that returns a struct with all flags (`{private}`, `{buttons}`, `{no_preview}`, …). Both modules can post-process the same output.  
  *Reason*: Prevents divergent parsing behaviour (bug reports where a flag works for notes but not filters) and simplifies future flag additions.

- [ ] **Consolidate button callback handlers**  
  *Duplicate*: `filtersButtonHandler` and `notesButtonHandler` differ only in DB lookup function.  
  *Fix*: Build a parameterised callback handler that accepts a lookup closure. Register it twice with different prefixes (`filters_`, `notes_`).  
  *Reason*: DRY principle, fewer callback registration bugs.

- [x] **Generic stats loaders**  
  *Duplicate*: Every *_db.go file has a `LoadXYZStats` function with identical cursor scanning code.  
  *Fix*: Add a helper `CountByChat(collection string, filter bson.M, chatField string) (items, chats int64)` and call it from thin wrappers.  
  *Reason*: Cuts ≥20 nearly identical functions, speeds up writing new features.

- [ ] **Share chat-settings retrieval logic**  
  *Duplicate*: `getNotesSettings`, `getChatGreetingSettings`, `getChatRulesSettings`, … all follow the same “findOne or insert default” pattern.  
  *Fix*: Create `GetOrInitSettings[T any](collection *mongo.Collection, id int64, default T) (*T, error)` generic.  
  *Reason*: Prevents silent divergence (e.g. default values drifting) and reduces code repetition.

- [ ] **Factor common struct fields into embedded base types**  
  *Duplicate*: Many DB structs repeat `ChatId int64`, `MsgType int`, `FileID string`, `Buttons []Button`, etc.  
  *Fix*: Define an embedded `MessagePayload` struct and reuse it. Combine with generics for collection-specific fields.  
  *Reason*: Cleaner models and easier marshalling/unmarshalling changes.

- [ ] **Eliminate duplicate HTML utility wrappers (`helpers.Shtml()` & friends)**  
  *Duplicate*: Several modules wrap the same `ParseMode` + error/sanitisation boilerplate.  
  *Fix*: Provide constants or small helpers in one place and reference them.  
  *Reason*: Single source of truth for Telegram formatting, easier migration if API changes.

- [ ] **Consolidate repetitive EnumFunc maps in helpers**  
  *Duplicate*: `FiltersEnumFuncMap` and `NotesEnumFuncMap` map the same integer message types to near-identical sender functions.  
  *Fix*: Create one `MessageTypeDispatch` map keyed by enum + module flags.  
  *Reason*: Avoids inconsistent behaviour where a new message type is added for filters but forgotten for notes. 