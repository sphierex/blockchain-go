# 工作量证明

区块链的一个关键点就是一个人必须经过一些列困难的工作，才能将数据放入到区块链中。 “努力工作并进行证明”的机制，就叫做工作量证明（proof-of-work）。

## 哈希计算

在区块链中，哈希被用于保证一个块的一致性。哈希算法的输入数据包含了前一个块的哈希。因此使得不太可能去修改链中的一个块。

如果一个人想要修改前面的一个块的数哈希，那么他必须要重新计算这个块以及后边所有块的哈希。

## Hashcash

比特币使用 [Hashcash](https://en.wikipedia.org/wiki/Hashcash)，一个用来防止垃圾邮件的工作量证明算法。它可以被分解为以下步骤。

1. 取一些公开的数据（比如，邮件是接收者的邮件地址。比特币中是区块头）。
2. 给这个公开数据添加一个计数器。计数器默认从零开始。
3. 将数据（data）和计算器（counter）组合在一起，获得一个哈希。
4. 检查哈希是否符合一定的条件。
   - 如果符合条件，结束。
   - 如果不符合，增加计数器，重复步骤 3 - 4。

具体过程：改变计数器 -> 计算新的哈希 -> 检查 -> 增加计算器 -> 计算哈希 -> 检查。。。

算法满足的必要条件。

- 在原始的 Hashcash 实现中，它的要求是 “一个哈希的前 20 位必须是 0”。
- 在比特币中，这个要求会根据情况进行调整，因为按照设计，无论计算能力随着时间的推移而增加或越来越多的矿工加入网络，区块都必须每10分钟生成一个。


## 引用

- [Proof Of Work](https://jeiwan.net/posts/building-blockchain-in-go-part-2/)
- [Proof Of Work (ZH-CN)](https://github.com/liuchengxu/blockchain-tutorial/blob/master/content/part-2/proof-of-work.md)
- [Hashcash](https://en.wikipedia.org/wiki/Hashcash)