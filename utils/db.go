package utils
import "github.com/boltdb/bolt"
import osuser "os/user"
import "log"


var(
	CURRENT_USER, _ = osuser.Current()
	dbFile = CURRENT_USER.HomeDir + "/.blockchain.db"
)


func Update(upfunc func(tx *bolt.Tx) error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	err = db.Update(upfunc)
	if err != nil {
		log.Panic(err)
	}
}

func View(upfunc func(tx *bolt.Tx) error) {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}
	defer db.Close()
	err = db.View(upfunc)
	if err != nil {
		log.Panic(err)
	}
}