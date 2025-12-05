package schemas

// UserConfig defines the schema for the user configuration file at:
// ~/.config/sow/config.yaml
//
// This allows users to configure which AI CLI executors handle which agent roles.
#UserConfig: {
	// Agent configuration
	agents?: {
		// Executor definitions
		// Keys are executor names (e.g., "claude-code", "cursor")
		executors?: [string]: {
			// Type of executor
			type: "claude" | "cursor" | "windsurf"

			// Executor settings
			settings?: {
				// Skip permission prompts
				yolo_mode?: bool
				// AI model to use (only meaningful for claude type)
				model?: string
			} @go(,optional=nillable)

			// Additional CLI arguments
			custom_args?: [...string]
		}

		// Bindings from agent roles to executor names
		bindings?: {
			orchestrator?: string
			implementer?: string
			architect?: string
			reviewer?: string
			planner?: string
			researcher?: string
			decomposer?: string
		} @go(,optional=nillable)
	} @go(,optional=nillable)
}
