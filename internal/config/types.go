package config

type Config struct {
	Programs map[string]Program `yaml:"programs"`
}

type Program struct {
	Cmd          string            `yaml:"cmd"`
	NumProcs     int               `yaml:"numprocs"`
	AutoStart    bool              `yaml:"autostart"`
	AutoRestart  string            `yaml:"autorestart"` // always, never, unexpected
	ExitCodes    []int             `yaml:"exitcodes"`
	StartTime    int               `yaml:"starttime"`    // seconds to consider "successfully started"
	StartRetries int               `yaml:"startretries"` // max restart attempts
	StopSignal   string            `yaml:"stopsignal"`   // TERM, KILL, USR1, etc.
	StopTime     int               `yaml:"stoptime"`     // seconds to wait before KILL
	Stdout       string            `yaml:"stdout"`       // stdout redirection
	Stderr       string            `yaml:"stderr"`       // stderr redirection
	Env          map[string]string `yaml:"env"`          // environment variables
	WorkingDir   string            `yaml:"workingdir"`   // working directory
	Umask        string            `yaml:"umask"`        // umask for process
}
