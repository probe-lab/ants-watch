package main

var rootConfig = struct {
	AntsClickhouseAddress  string
	AntsClickhouseDatabase string
	AntsClickhouseUsername string
	AntsClickhousePassword string

	NebulaDBConnString string
	KeyDBPath          string

	NumPorts  int
	FirstPort int
	UPnp      bool
}{
	AntsClickhouseAddress:  "",
	AntsClickhouseDatabase: "",
	AntsClickhouseUsername: "",
	AntsClickhousePassword: "",

	NebulaDBConnString: "",
	KeyDBPath:          "keys.db",

	NumPorts:  128,
	FirstPort: 6000,
	UPnp:      false,
}
