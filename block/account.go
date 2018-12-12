package block
import "bytes"
import "encoding/gob"
import "log"
import "github.com/boltdb/bolt"
import "fmt"
import "github.com/symphonyprotocol/scb/utils"
// import "sort"  

// const accountBucket = "account"

type Account struct{
	Address string
	Balance int64
	GasBalance int64
	Nonce  int64
}

type AccountHistory struct{
	Timestamp int64
	Address string
	ExchangeAmount int64
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

func ChangeBalance(address string, balance int64, gas int64){

	utils.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(accountBucket))
			accountbytes := bucket.Get([]byte(address))
			
			var newbalance int64
			var newnonce int64
			var newgas int64

			var newaccount *Account
	
			if accountbytes == nil{
				newbalance = balance
				newgas = gas
				// newnonce = 0
			}else{
				account := DeserializeAccount(accountbytes)
				newbalance = account.Balance + balance
				newgas = account.GasBalance + gas
				newnonce =  account.Nonce
			}
	
			if newbalance < 0 {
				return fmt.Errorf("no enough amount")
			}
			if newgas < 0{
				return fmt.Errorf("no enount gas")
			}
	
			newaccount = NewAccount(address, newbalance, newnonce, newgas)
	
			if accountbytes == nil{
				bucket.Put([]byte(address), newaccount.Serialize())
			}else{
				bucket.Delete([]byte(address))
				bucket.Put([]byte(address), newaccount.Serialize())
			}
			return nil
		})
}

func NoncePlus(address string){
	utils.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(accountBucket))
		accountbytes := bucket.Get([]byte(address))
		if accountbytes != nil{
			account := DeserializeAccount(accountbytes)
			account.Nonce += 1
			bucket.Delete([]byte(address))
			bucket.Put([]byte(address), account.Serialize())
		}
		return nil
	})
}

func GetBalance(address string) (int64,int64){
	// var balance int64 = 0
	account := GetAccount(address)
	if account != nil{
		fmt.Printf("nonce is :%v\n",  account.Nonce)
		return account.Balance, account.GasBalance
	}
	return -1, -1
}

func GetAccount(address string) *Account{
	var account *Account
	utils.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(accountBucket))
		accountbytes := bucket.Get([]byte(address))
		if accountbytes != nil{
			account = DeserializeAccount(accountbytes)
		}else{
			// bucket.Put()
			account = NewAccount(address, 0, 0, 0)
			bucket.Put([]byte(address), account.Serialize())
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
	account := NewAccount(address, 0 , 0, 0)
	return account
}

func NewAccount(address string, balance, nonce, gas int64) *Account{
	account := Account{
		Address : address,
		Balance : balance,
		Nonce   : nonce,
		GasBalance : gas,
	}
	return &account
}

func GetAllAccount() []*Account {
	var accounts [] *Account

	utils.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(accountBucket))
		if b == nil{
			return nil
		}
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			// fmt.Printf("key=%s, value=%s\n", k, v)
			account := DeserializeAccount(v)
			accounts = append(accounts, account)
		}
		return nil
	})
	// sort.Slice(accounts,func(i, j int) bool{
	// 	return accounts[i].Address < accounts[j].Address
	// })
	return accounts
}

func FindAccount(accounts []*Account , address string) *Account{
	for _, account := range accounts{
		if account.Address == address{
			return account
		}
	}
	return nil
}