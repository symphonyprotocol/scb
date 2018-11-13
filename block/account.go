package block
import "bytes"
import "encoding/gob"
import "log"

// const accountBucket = "account"

type Account struct{
	Address string
	Balance int64
	Nonce  int64
}
  
// Serializes the block
func (b *Account) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// Deserializes a Account
func DeserializeAccount(d []byte) *Account {
	var account Account

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&account)
	if err != nil {
		log.Panic(err)
	}

	return &account
}

func NewAccount(address string) * Account{
	account := Account{
		Address : address,
		Balance : 0,
		Nonce   : 0,
	}
	return &account
}