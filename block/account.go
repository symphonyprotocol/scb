package block
import "bytes"
import "encoding/gob"
import "encoding/binary"
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

type AccountIncreaseMent struct{
	ChangedAccount [] *Account
	NewAccount [] *Account
}

// Serializes the account increasement
func (a *AccountIncreaseMent) Serialize() []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(a)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

// Deserializes account increasement
func DeserializeAccountIncreasement(d []byte) *AccountIncreaseMent {
	var account AccountIncreaseMent

	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(&account)
	if err != nil {
		log.Panic(err)
	}
	return &account
}


// Serializes the account
func (a *Account) Serialize() []byte {
	var b0 []byte

	b1 := [] byte(a.Address)
	b2 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b2, uint64(a.Balance))
	b3 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b3, uint64(a.Index))
	b4 := make([]byte, 8)
	binary.LittleEndian.PutUint64(b4, uint64(a.Nonce))
	
	var bytes []byte
	var len_addr byte
	len_addr = byte(len(b1))

	b0 = append(b0, len_addr)

	bytes = append(b0, b1...)
	bytes = append(bytes, b2...)
	bytes = append(bytes, b3...)
	bytes = append(bytes, b4...)

	return bytes
}

// Deserializes a Account
func DeserializeAccount(d []byte) *Account {
	var account Account
	len := d[0]

	b_addr := d[1:len+1]
	d = d[len+1:]
	b_ba := d[:8]
	d = d[8:]
	b_idx:= d[:8]
	d= d[8:]
	b_non:= d[:8]

	balance := int64(binary.LittleEndian.Uint64(b_ba))
	index := int64(binary.LittleEndian.Uint64(b_idx))
	nonce := int64(binary.LittleEndian.Uint64(b_non))

	account.Address = string(b_addr)
	account.Balance = balance
	account.Index = index
	account.Nonce = nonce
	

	return &account
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


func InitAccount(address string, idx int64) *Account{
	// account := NewAccount(address, 0 , 0, 0)
	// return account
	// var idx int64 = 0

	// if dbExists(){
	// 	utils.View(func(tx *bolt.Tx) error {
	// 		b := tx.Bucket([]byte(accountBucket))
	// 		if b == nil{
	// 			return nil
	// 		}
	// 		c := b.Cursor()
	// 		for k, v := c.First(); k != nil; k, v = c.Next() {
	// 			fmt.Printf("key=%s, value=%s\n", k, v)
	// 			idx++
	// 		}
	// 		return nil
	// 	})
	// }
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

func GetAllAccountTest() []*Account {
	var accounts [] *Account

	utils.ViewTest(func(tx *bolt.Tx) error {
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