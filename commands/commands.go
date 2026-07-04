package commands

// Registry holds all available commands dynamically
var CommandRegistry = map[string]Commander{
	"b64":     B64Command{},
	"help":    HelpCommand{},
	"version": VersionCommand{},
}

// Commander defines the behavior for all CLI commands
type Commander interface {
	Execute(args []string) error
	Description() string
}
