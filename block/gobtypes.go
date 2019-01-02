package block

import "encoding/gob"

var typesRegistered = false

func RegisterSCBTypes() {
	if !typesRegistered {
		gob.Register(&Account{})
		gob.Register(&Block{})
		gob.Register(Transaction{})
		gob.Register(&NodeShadow{})
		gob.Register(BlockContent{})
		typesRegistered = true
	}
}

