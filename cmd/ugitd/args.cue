import "strings"

// Schema
#Port: int & >0 & <65536
#Link: string & strings.Contains(",")

#Config: {
	"repo-dir":     string
	"show-private": bool
	ssh: {
		enable:            bool
		"authorized-keys": string
		"clone-url":       string
		port:              #Port
		"host-key":        string
	}
	http: {
		enable:      bool
		"clone-url": string
		port:        #Port
	}
	meta: {
		title:       string
		description: string
	}
	profile?: {
		username?: string
		email?:    string
		links?: [...#Link]
	}
	log: {
		json:  bool
		level: "debug" | "info" | "warn" | "warning" | "error"
	}

	// Constraints
	if ssh.port == http.port {
		error("ssh.port and http.port cannot be the same")
	}
}

// Defaults
#Config: {
	"repo-dir":     ".ugit"
	"show-private": false
	ssh: {
		enable:            true
		"authorized-keys": ".ssh/authorized_keys"
		"clone-url":       "ssh://localhost:8448"
		port:              8448
		"host-key":        ".ssh/ugit_ed25519"
	}
	http: {
		enable:      true
		"clone-url": "http://localhost:8449"
		port:        8449
	}
	meta: {
		title:       "uGit"
		description: "Minimal git server"
	}
	log: {
		json:  false
		level: "info"
	}
}

// Apply schema
#Config
