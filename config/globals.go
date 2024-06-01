package config

import (
	"strings"
)

var GlobalPayload string
var Delay int
var Round int
var Node int
var Shard3Leader int
var Shard2Leader int
var Shard1Leader int
var Shard4Leader int
var Shard5Leader int
var Shard6Leader int
var Shard7Leader int
var MaliciousNode int
var ShardNumber int

func init() {
	GlobalPayload = strings.Repeat("a", 1000) // Initialize the global payload with 400 'a' characters
	Delay = 100
	Round = 5
	Node = 4
	ShardNumber = 5
	MaliciousNode = 1
	Shard3Leader = 8000 + Node*2 + 1
	Shard1Leader = 8000 + Node*0 + 1
	Shard2Leader = 8000 + Node*1 + 1
	Shard4Leader = 8000 + Node*3 + 1
	Shard5Leader = 8000 + Node*4 + 1
	Shard6Leader = 8000 + Node*5 + 1
	Shard7Leader = 8000 + Node*6 + 1
}
