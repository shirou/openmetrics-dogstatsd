package main

import (
	"os"
	"os/user"
	"path/filepath"
	"text/template"
)

var systemdServiceTmpl = `[Unit]
Description=send openmetrics to dogstats
After=network.target

[Service]
Type=simple
ExecStart={{ .binary_path }} -d 60
WorkingDirectory={{ .work_dir }}
Restart=always
User={{ .user }}
RestartSec=5
StartLimitInterval=0
StartLimitBurst=5

# Logging
StandardOutput=append:/var/log/{{ .binary_name }}.log
StandardError=append:/var/log/{{ .binary_name }}.log

[Install]
WantedBy=multi-user.target
`

func printSystemdService() error {
	tmpl, err := template.New("systemd").Parse(systemdServiceTmpl)
	if err != nil {
		return err
	}

	// get current path of executed binary
	ex, err := os.Executable()
	if err != nil {
		return err
	}

	currentUser, err := user.Current()
	if err != nil {
		return err
	}
	values := map[string]interface{}{
		"binary_path": ex,
		"work_dir":    filepath.Dir(ex),
		"binary_name": filepath.Base(ex),
		"user":        currentUser.Username,
	}

	return tmpl.Execute(os.Stdout, values)
}
