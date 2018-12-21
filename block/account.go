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
	Nonce  int64
	Index int64
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

func ChangeBalance(address string, balance int64){
	initaccount := InitAccount(address)

	utils.Update(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte(accountBucket))
			accountbytes := bucket.Get([]byte(address))

			// var newaccount *Account
			var account *Account
	
			if accountbytes == nil{
				account = initaccount
			}else{
				account = DeserializeAccount(accountbytes)
			}

			account.Balance += balance
	
			if account.Balance < 0 {
				log.Panic("no enough amount")
			}
	
			if accountbytes == nil{
				bucket.Put([]byte(address), account.Serialize())
			}else{
				bucket.Delete([]byte(address))
				bucket.Put([]byte(address), account.Serialize())
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

func GetBalance(address string) (int64){
	// var balance int64 = 0
	account := GetAccount(address)
	if account != nil{
		fmt.Printf("nonce is :%v\n",  account.Nonce)
		return account.Balance
	}
	return -1
}

func GetAccount(address string) *Account{
	var account *Account
	utils.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(accountBucket))
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
	// account := NewAccount(address, 0 , 0, 0)
	// return account
	var idx int64 = 0

	if dbExists(){
		utils.View(func(tx *bolt.Tx) error {
			b := tx.Bucket([]byte(accountBucket))
			if b == nil{
				return nil
			}
			c := b.Cursor()
			for k, v := c.First(); k != nil; k, v = c.Next() {
				fmt.Printf("key=%s, value=%s\n", k, v)
				idx++
			}
			return nil
		})
	}
	account := NewAccount(address, 0, 0, idx)
	return account
}

func NewAccount(address string, balance, nonce, index int64) *Account{
	account := Account{
		Address : address,
		Balance : balance,
		Nonce   : nonce,
		Index: index,
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