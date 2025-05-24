package command

import (
	"github.com/PaulSonOfLars/gotgbot/v2"
	"github.com/PaulSonOfLars/gotgbot/v2/ext"
	"github.com/divideprojects/Alita_Robot/alita/db"
	"github.com/divideprojects/Alita_Robot/alita/utils/helpers"
	"github.com/divideprojects/Alita_Robot/alita/utils/messaging"
	"github.com/divideprojects/Alita_Robot/alita/utils/permissions"
	"github.com/divideprojects/Alita_Robot/alita/utils/validation"
	log "github.com/sirupsen/logrus"
)

// CommandHandler provides a fluent API for setting up command handlers with integrated frameworks
type CommandHandler struct {
	connectCheck    func(*gotgbot.Bot, *ext.Context) *gotgbot.Chat
	permissionCheck *permissions.PermissionChecker
	userValidation  func(*gotgbot.Bot, *ext.Context) (*validation.UserValidationResult, error)
	argValidator    func([]string, *gotgbot.Message) (bool, string)
	handler         func(*gotgbot.Bot, *ext.Context, *gotgbot.Chat, *gotgbot.User, []string) error
	errorHandler    func(error) error
}

// NewCommandHandler creates a new command handler with default settings
func NewCommandHandler() *CommandHandler {
	return &CommandHandler{
		connectCheck: defaultConnectCheck,
		errorHandler: defaultErrorHandler,
	}
}

// WithConnectionCheck sets the connection check function
func (ch *CommandHandler) WithConnectionCheck(check func(*gotgbot.Bot, *ext.Context) *gotgbot.Chat) *CommandHandler {
	ch.connectCheck = check
	return ch
}

// WithPermissions sets the permission checker
func (ch *CommandHandler) WithPermissions(checker *permissions.PermissionChecker) *CommandHandler {
	ch.permissionCheck = checker
	return ch
}

// WithUserValidation sets the user validator
func (ch *CommandHandler) WithUserValidation(validator func(*gotgbot.Bot, *ext.Context) (*validation.UserValidationResult, error)) *CommandHandler {
	ch.userValidation = validator
	return ch
}

// WithArgValidation sets the argument validator
func (ch *CommandHandler) WithArgValidation(validator func([]string, *gotgbot.Message) (bool, string)) *CommandHandler {
	ch.argValidator = validator
	return ch
}

// WithErrorHandler sets a custom error handler
func (ch *CommandHandler) WithErrorHandler(handler func(error) error) *CommandHandler {
	ch.errorHandler = handler
	return ch
}

// WithHandler sets the main command handler function
func (ch *CommandHandler) WithHandler(handler func(*gotgbot.Bot, *ext.Context, *gotgbot.Chat, *gotgbot.User, []string) error) *CommandHandler {
	ch.handler = handler
	return ch
}

// Execute runs the command handler with all configured checks and validations
func (ch *CommandHandler) Execute(b *gotgbot.Bot, ctx *ext.Context) error {
	msg := ctx.EffectiveMessage

	// Connection check
	connectedChat := ch.connectCheck(b, ctx)
	if connectedChat == nil {
		return ext.EndGroups
	}
	ctx.EffectiveChat = connectedChat
	chat := ctx.EffectiveChat
	user := ctx.EffectiveSender.User
	args := ctx.Args()

	// Permission check using new framework
	if ch.permissionCheck != nil {
		if !ch.permissionCheck.Check() {
			return ext.EndGroups
		}
	}

	// User validation using new framework
	if ch.userValidation != nil {
		result, err := ch.userValidation(b, ctx)
		if err != nil {
			return ch.errorHandler(err)
		}
		if !result.Valid {
			return ext.EndGroups
		}
	}

	// Argument validation
	if ch.argValidator != nil {
		if ok, errMsg := ch.argValidator(args, msg); !ok {
			if errMsg != "" {
				content := &messaging.GreetingContent{
					Text:     errMsg,
					DataType: db.TEXT,
				}
				opts := &messaging.SendOptions{
					ReplyToMessageID: msg.MessageId,
					ParseMode:        helpers.HTML,
				}
				_, err := messaging.SendMessage(b, ctx, content, opts)
				if err != nil {
					return ch.errorHandler(err)
				}
			}
			return ext.EndGroups
		}
	}

	// Execute main handler
	if ch.handler != nil {
		err := ch.handler(b, ctx, chat, user, args)
		if err != nil {
			return ch.errorHandler(err)
		}
	}

	return ext.EndGroups
}

// Helper function to create a gotgbot handler function from CommandHandler
func (ch *CommandHandler) ToHandler() func(*gotgbot.Bot, *ext.Context) error {
	return ch.Execute
}

// Default implementations
func defaultConnectCheck(b *gotgbot.Bot, ctx *ext.Context) *gotgbot.Chat {
	// Default implementation - just return the effective chat
	return ctx.EffectiveChat
}

func defaultErrorHandler(err error) error {
	log.Error(err)
	return err
}

// Convenience functions for common patterns

// SimpleAdminCommand creates a command handler for admin-only commands
func SimpleAdminCommand(b *gotgbot.Bot, ctx *ext.Context, handler func(*gotgbot.Bot, *ext.Context, *gotgbot.Chat, *gotgbot.User, []string) error) *CommandHandler {
	return NewCommandHandler().
		WithPermissions(permissions.NewPermissionChecker(b, ctx).RequireUserAdmin(ctx.EffectiveSender.User.Id)).
		WithHandler(handler)
}

// SimpleUserCommand creates a command handler for regular user commands
func SimpleUserCommand(handler func(*gotgbot.Bot, *ext.Context, *gotgbot.Chat, *gotgbot.User, []string) error) *CommandHandler {
	return NewCommandHandler().
		WithUserValidation(validation.ValidateUser).
		WithHandler(handler)
}

// AdminCommandWithArgs creates an admin command that requires arguments
func AdminCommandWithArgs(
	b *gotgbot.Bot,
	ctx *ext.Context,
	argValidator func([]string, *gotgbot.Message) (bool, string),
	handler func(*gotgbot.Bot, *ext.Context, *gotgbot.Chat, *gotgbot.User, []string) error,
) *CommandHandler {
	return NewCommandHandler().
		WithPermissions(permissions.NewPermissionChecker(b, ctx).RequireUserAdmin(ctx.EffectiveSender.User.Id)).
		WithUserValidation(validation.ValidateUser).
		WithArgValidation(argValidator).
		WithHandler(handler)
}

// ConnectedAdminCommand creates an admin command that works with connections
func ConnectedAdminCommand(
	b *gotgbot.Bot,
	ctx *ext.Context,
	connectCheck func(*gotgbot.Bot, *ext.Context) *gotgbot.Chat,
	handler func(*gotgbot.Bot, *ext.Context, *gotgbot.Chat, *gotgbot.User, []string) error,
) *CommandHandler {
	return NewCommandHandler().
		WithConnectionCheck(connectCheck).
		WithPermissions(permissions.NewPermissionChecker(b, ctx).RequireUserAdmin(ctx.EffectiveSender.User.Id)).
		WithUserValidation(validation.ValidateUser).
		WithHandler(handler)
}

// Enhanced version of the existing HandleCommandWithChecks that integrates new frameworks
func HandleCommandWithChecksEnhanced(
	b *gotgbot.Bot,
	ctx *ext.Context,
	connectCheck func(*gotgbot.Bot, *ext.Context) *gotgbot.Chat,
	permissionChecker *permissions.PermissionChecker,
	userValidator func(*gotgbot.Bot, *ext.Context) (*validation.UserValidationResult, error),
	argCheck func([]string, *gotgbot.Message) (bool, string),
	handler func(*gotgbot.Bot, *ext.Context, *gotgbot.Chat, *gotgbot.User, []string) error,
) error {
	return NewCommandHandler().
		WithConnectionCheck(connectCheck).
		WithPermissions(permissionChecker).
		WithUserValidation(userValidator).
		WithArgValidation(argCheck).
		WithHandler(handler).
		Execute(b, ctx)
}
