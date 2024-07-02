package listener

const (
	ExternalPublicListenerName        = "extpub"
	ExternalPublicListenerPortNumber  = 9092 //this port is intended to be used by clients output the private network and outside the container cluster
	ExternalPrivateListenerName       = "extpvt"
	ExternalPrivateListenerPortNumber = 9093 //this port is intended to be used by clients inside the private network but outside the container cluster
	InternalListenerName              = "int"
	InternalListenerPortNumber        = 9094 //this port is intended to be used by clients inside the container cluster
)
