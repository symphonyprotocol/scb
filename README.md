# conblocks
Symphonyprotocol  Blocks and Consensus algorithm


###  build 

```
go build -o scb
```

### create blockchain
```
scb createblockchain -wif L5fR7FRHnZGL3DjsrhN8CvBYHpywL8LjxA2rjzbL7qvFqjgbNVQ5
```

###  get balance 

```
./scb getbalance -address L5fR7FRHnZGL3DjsrhN8CvBYHpywL8LjxA2rjzbL7qvFqjgbNVQ5
```

### send 

```
 ./scb send -from 1T3r9yFFM6St9wGSp7zMYP24G6pUYnL7y -to 189wh8VjXLmKSZhnP9DQwcVKfvNemQSmBp -amount 3 -wif L5fR7FRHnZGL3DjsrhN8CvBYH
pywL8LjxA2rjzbL7qvFqjgbNVQ5
```
### mine

```
./scb mine -wif L5fR7FRHnZGL3DjsrhN8CvBYHpywL8LjxA2rjzbL7qvFqjgbNVQ5
```

## printchain

```
./scb printchain
```
## module support

* In Progress