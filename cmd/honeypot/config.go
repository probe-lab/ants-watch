package main

var RootConfig = struct {
	AntsClickhouseAddress  string
	AntsClickhouseDatabase string
	AntsClickhouseUsername string
	AntsClickhousePassword string
	AntsClickhouseSSL      bool

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
	AntsClickhouseSSL:      true,

	NebulaDBConnString: "",
	KeyDBPath:          "keys.db",

	NumPorts:  128,
	FirstPort: 6000,
	UPnp:      false,
}
