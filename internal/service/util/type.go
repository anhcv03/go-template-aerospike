package util

type CommandOption struct {
	Input   string   `json:"input,omitempty"`   // Command input. (Default: "")
	Command string   `json:"command,omitempty"` // Command. (Default: "")
	Options []string `json:"options,omitempty"` // Command options. (Default: nil)

	Logging  bool         `json:"logging,omitempty"`  // Is logging. (Default: false)
	Interval uint32       `json:"interval,omitempty"` // Interval time in milisecond for call callback function. (Default: 0)
	Callback func(string) `json:"callback,omitempty"` // Callback function. (Default: nil)
}
