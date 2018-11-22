package block
import "bytes"
import "encoding/gob"
import "log"
import "github.com/boltdb/bolt"
import "fmt"
import "github.com/symphonyprotocol/scb/utils"

// const accountBucket = "account"

type Account struct{
	Address string
	Balance int64
	Nonce  int64
}
  
// Serializes the block
func (a *Account) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(a)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func ChangeBalance(address string, balance int64, isGas bool){

	
	utils.Update(func(tx *bolt.Tx) error {
			var bucket *bolt.Bucket

			if isGas{
				bucket = tx.Bucket([]byte(gasBucket))
			}else{
				bucket = tx.Bucket([]byte(accountBucket))
			}
			
			accountbytes := bucket.Get([]byte(address))
			
			var newbalance int64
			var newnonce int64
			var newaccount *Account
	
			if accountbytes == nil{
				newbalance = balance
				newnonce = 0
				
			}else{
				account := DeserializeAccount(accountbytes)
				newbalance = account.Balance + balance
				newnonce =  account.Nonce + 1
			}
	
			if newbalance < 0 {
				return fmt.Errorf("no enough amount")
			}
	
			newaccount = NewAccount(address, newbalance, newnonce)
	
			if accountbytes == nil{
				bucket.Put([]byte(address), newaccount.Serialize())
			}else{
				bucket.Delete([]byte(address))
				bucket.Put([]byte(address), newaccount.Serialize())
			}
			return nil
		})
}




func GetBalance(address string, isGas bool) int64{
	var balance int64 = 0
	account := GetAccount(address, isGas)
	if account != nil{
		balance = account.Balance
	}
	// fmt.Printf("balance is: %v\n", balance)
	return balance
}

func GetAccount(address string, isGas bool) *Account{
	var account *Account
	utils.View(func(tx *bolt.Tx) error {
		var bucket *bolt.Bucket
		if isGas{
			bucket = tx.Bucket([]byte(gasBucket))
		}else{
			bucket = tx.Bucket([]byte(accountBucket))
		}
		accountbytes := bucket.Get([]byte(address))
		if accountbytes != nil{
			account = DeserializeAccount(accountbytes)
		}
		return nil
	})
	return account
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

func InitAccount(address string) *Account{
	account := NewAccount(address, 0 , 0)
	return account
}

func NewAccount(address string, balance, nonce int64) *Account{
	account := Account{
		Address : address,
		Balance : balance,
		Nonce   : nonce,
	}
	return &account
}