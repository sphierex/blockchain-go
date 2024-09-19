# 交易

交易（transaction）是比特币核心所在。

在区块链中，交易一旦被创建，就没有任何人能够去修改它或者删除它。

比特币采用的 UTXO 模型，并非账户模型，并不直接存在“余额”这个概念，余额需要通过遍历整个交易历史得来。

## 比特币交易

一笔交易由一些输入（input）和输出（output）组合而来。

对于每一笔新的交易，它的输入会引用之间一笔交易的输出，引用就是花费的意思。所谓引用之前的一个输出，也就是将之前的一个输出包含在另一笔交易的输入当中，就是花费之
前的交易输出.

**交易的完整生命周期**

1. 一开始，有一个包含 coinbase 交易的创世区块。coinbase 交易中没有实际输入，因此无需签名。coinbase交易输出包含一个哈希公钥（RIPEMD16(SHA256(PubKey))使用算法）。
2. 当有人发送货币时，就会创建一个交易。交易的输入将引入先前交易的输出。每个输入都将储存一个公钥（未散列）和整个交易的签名。
3. 比特币网络中接收交易的其它节点将验证该交易。除其它事项外，他们还将检查：输入中的公钥哈希值是否于引用的输出哈希值相匹配（这可确保发送者只花费属于自己的货币）。签名是否正确（这可确保交易是由货币真正所有者创建的）。
4. 当矿工节点准备挖掘新区块时，它会将交易放入区块并开始挖掘。
5. 当区块被挖掘出来时，网络中的每个其它节点都会收到一条消息，表明该区块已经被挖掘，并将该区块添加到区块链中。
6. 当一个区块被添加到区块链后，交易就完成了，它的输出可以在新的交易中被引用。

### 交易输出/输入

关于输出的一个重要特点是他们是不可分割的，这意味着你不能引用其价值的一部分。当在新的交易中引用输出时，它将作为一个整体被使用。
如果其价值大于要求，则会生成 一个零钱并发送会发送者

输出是存储“硬币“的地方。每个输出都有一个解锁脚本，该脚本决定了解锁输出的逻辑。每个新交易都必须至少有一个输入和输出。
输入引用来自前一个交易的输出，并提供 `ScriptSIG` 输出解锁脚本中使用的数据，已解锁它并使用其值创建新的输出。

**注意**：在比特币中，先有蛋（输出），后有鸡（输入）。

当矿工开始挖掘一个区块时，它会向其中田间一个 `coinbase` 交易。`coinbase` 交易是一种特殊类型的交易，它不需要先前存在的输出。
它会凭空创造输出（即”硬币“）。这是矿工挖掘新区快所获得的奖励。

- 一个与接收者地址锁定的地址。这是将货币实际转移到其他地址。
- 一个与发送者地址锁定的地址。这是一个变化，只有当未使用的输出所含简直超过新交易所需的价值时才会创建它。

**输出时不可分割的**

### 交易如何在比特币网络中运作

- 发送方：选择自己的 UTXO 作为输入，并创建相应的输出。
- 交易结构：交易包括输入（引用之前的 UTXO）和输出（指定接收者和找零地址）。
- 交易验证：网络中的节点验证交易的合法性，矿工将交易打包进区块。
- 接收方：接受新的 UTXO，可以在后续交易中使用这些 UTXO 作为输入。

## 地址

在比特币中，你的身份是存储在你的计算机上（或存储在其它可以访问的地方）的一对（或多对）私钥和公钥。
比特币依靠多种加密算法来创建这些密钥，并保证时间上其它人无法没在没有密钥的情况下访问你的比特币。

公钥加密算法使用一对密钥：公钥和私钥。公钥并不敏感，可以向任何人披露。

## 数字签名

保证以下功能的算法：

1. 数据从发送方传输到接收方的过程中未被修改。
2. 该数据是由某个发送者创建的。
3. 发送者不能否认发送了数据。
 
### 签署

- 要签名的数据
- 私钥

### 验签

- 已签名的数据。
- 签名。
- 公钥。

数字签名不是加密，无法从签名中重建数据。类似于哈希：通过哈希算法运行数据并获得数据的唯一表示。签名和哈希之间的区别在于密钥对。

密钥对也可用于加密数据：私钥用于加密，公钥用于解密数据。然而，比特币没有使用加密算法。

比特币中输入的每笔交易都由创建交易的人签名。比特币中的每笔交易都必须经过验证才能放入区块。

- 检查输入是否有权使用以前的交易的输出。
- 检查交易签名是否正确。

## 公钥获取地址的过程

Base58 算法。相较于 Base64 算法。删除了一些字母0（零），O(大写o)，I（大写i），l（小写l），+，/ 符号。

公钥：Base58Encode(Version + Public key hash (RIPEMD160(SHA256(PubKey))) + Checksum (Sha256(Sha256(PubKeyHash))))

**公钥转换为 `Base58` 地址过程**

1. 获取公钥并使用 `RIPEMD160(SHA256(PubKey))` 哈希算法对其进行两次哈希处理。
2. 将地址生成算法的版本添加到哈希值前面。
3. 将步骤 `2` 的结果与进行哈希运算，计算校验和 `SHA256(SHA256(Payload))`。校验和是结果哈希的前四个字节。
4. 将校验和附加到 `Version+PubKeyHash` 组合。
5. 使用 `Base58` 对组合进行编码 `Version+PubKeyHash+Checknum`

## UTXO SET

**chain-state**

- 'c' + 32-byte transactions hash -> unspent transaction output record for that transaction.
- 'B' -> 32=byte block hash: the block hash up to which the database represents the unspent transaction output.

`chain-state` 不存储交易。相反，它存储所谓的 UTXO 集，即未使用的交易输出集。除此之外，它还存储“数据库表示未使用交易输出的区块哈希”。

## MerkleTree

每个区块都会构建一个 `Merkle` 树，它从叶子（树的底部）开始，其中叶子是交易哈希（比特币使用双重`sha256`哈希）。叶子的数量必须是偶数，但并非每个区块都包含偶数个交易。
如果交易数量为奇数，则最后一笔交易将被复制（在树中，不在区块中）。

## 命令

```shell
go run cmd/main.go send --from fromAccount --to toAccount --amount num
# go run cmd/main.go send --from 14uMwAoZmRBiAQNhBmtexiyjweTy3YSivh --to 13YppsEF6NqNVUqLHtmg8z4WjGzURqpFqy --amount 4
``` 

## 参考资料
- [Transaction 1](https://jeiwan.net/posts/building-blockchain-in-go-part-4/)
- [Transaction 2](https://jeiwan.net/posts/building-blockchain-in-go-part-6/)
- [Bitcoin Transaction](https://en.bitcoin.it/wiki/Transaction)
- [ECDSA](https://www.bilibili.com/video/BV1BY411M74G)
- [UTXO-SET](https://en.bitcoin.it/wiki/Bitcoin_Core_0.11_(ch_2):_Data_Storage#The_UTXO_set_.28chainstate_leveldb.29)
- [MerkleTree](https://en.bitcoin.it/wiki/Protocol_documentation#Merkle_Trees)