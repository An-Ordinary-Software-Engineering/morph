package types

type defaults struct{
	CurrentTimestamp string
}

var Defaults = defaults{
	CurrentTimestamp: "CURRENT_TIMESTAMP()",
}
