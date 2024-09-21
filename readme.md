# blockchain-go - 用 golang 从零开始构建区块链系列

[![Actions Status](https://github.com/sphierex/blockchain-go/workflows/main.yml/badge.svg)](https://github.com/sphierex/blockchain-go/actions)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**注意：需使用 Golang 1.18.9**

a simple blockchain demo for learning

## the code for each article

1. part1: Basic Prototype [commit 1b242f4](https://github.com/sphierex/blockchain-go/commit/1b242f4c55de89e43f0fb7881e33c275b36cb048)
2. part2: Proof-of-Work [commit fdf3149](https://github.com/sphierex/blockchain-go/commit/fdf3149cf5a4b614ebaed5427d51b8451f946ca3)
3. part3: Persistence and Cli [commit 77b2ded](https://github.com/sphierex/blockchain-go/commit/77b2dededf8c9d069e82dd1d30ea19c9ff826e46)
4. part4: Transaction 1 [commit 4699e80](https://github.com/sphierex/blockchain-go/commit/4699e80ff6bbb2e8d1bf297f3f088990657c7588)
5. part5: Addresses [commit 27c1a8a](https://github.com/sphierex/blockchain-go/commit/27c1a8a688d97da4ec8e0372fefeef579a935f30)
6. part6: Transaction 2 [commit d99b1a1](https://github.com/sphierex/blockchain-go/commit/d99b1a1ce2ee48efe2c0f3db5ba4970b99acf2dd)
7. part7: Network [commit 25a9151](https://github.com/sphierex/blockchain-go/commit/25a91510d844263c2446c09dffcae4eceb362f2a)
8. part8: Optimize code

## 命令

```shell
go run cmd/main.go create-wallet
go run cmd/main.go create-wallet
go run cmd/main.go print-addresses

# 13pGasXsfb6Wcejyxa7kAX5P47av1FM4AF
# 1CSv68gmr1mMFWjAvjfg5AWgPBhz66jqsF

go run cmd/main.go create-chain --address 13pGasXsfb6Wcejyxa7kAX5P47av1FM4AF
go run cmd/main.go print-chain

go run cmd/main.go transfer --from 13pGasXsfb6Wcejyxa7kAX5P47av1FM4AF --to 1CSv68gmr1mMFWjAvjfg5AWgPBhz66jqsF --amount 5 --mine
go run cmd/main.go get-balance --address 13pGasXsfb6Wcejyxa7kAX5P47av1FM4AF
go run cmd/main.go get-balance --address 1CSv68gmr1mMFWjAvjfg5AWgPBhz66jqsF
```