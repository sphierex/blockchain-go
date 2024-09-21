package blockchain

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	cmdLength   = 12
	nodeVersion = 1
)

const (
	VersionCmd   = "version"
	AddrCmd      = "addr"
	BlockCmd     = "block"
	GetDataCmd   = "get_data"
	InvCmd       = "inv"
	GetBlocksCmd = "get_blocks"
	TxCmd        = "tx"
)

type Server struct {
	Id           string
	MinerAddress string

	endpoints []string
	endpoint  string

	bc             *Blockchain
	us             *UTXOSet
	blockInTransit [][]byte
	mempool        map[string]Transaction
}

func NewServer(id, miner string) *Server {
	bc, _ := NewBlockchain(id)

	return &Server{
		Id:             id,
		MinerAddress:   miner,
		bc:             bc,
		endpoints:      []string{"localhost:3000"},
		endpoint:       fmt.Sprintf("localhost:%s", id),
		blockInTransit: make([][]byte, 0),
	}
}

func NewServerWithBlockchain(bc *Blockchain, id, miner string) *Server {
	return &Server{
		Id:             id,
		MinerAddress:   miner,
		bc:             bc,
		endpoints:      []string{"localhost:3000"},
		endpoint:       fmt.Sprintf("localhost:%s", id),
		blockInTransit: make([][]byte, 0),
	}
}

// Start starts a node.
func (n *Server) Start() error {
	ln, err := net.Listen("tcp", n.endpoint)
	if err != nil {
		return err
	}
	defer func(ln net.Listener) {
		_ = ln.Close()
	}(ln)

	if n.endpoint != n.endpoints[0] {
		n.sendVersion(n.endpoints[0])
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go n.handleConn(conn)
	}
}

func (n *Server) handleConn(conn net.Conn) {
	req, err := io.ReadAll(conn)
	if err != nil {
		return
	}

	cmd := bytesToCmd(req[:cmdLength])
	log.Printf("Receive %s cmd\n", cmd)

	switch cmd {
	case AddrCmd:
		n.handleAddr(req)
	case BlockCmd:
		n.handleBlock(req)
	case InvCmd:
		n.handleInv(req)
	case GetBlocksCmd:
		n.handleGetBlocks(req)
	case GetDataCmd:
		n.handleGetData(req)
	case TxCmd:
		n.handleTx(req)
	case VersionCmd:
		n.handleVersion(req)
	default:
		log.Printf("unknown cmd: '%s'\r\n", cmd)
	}

	_ = conn.Close()
}

func (n *Server) hasEndpoint(addr string) bool {
	for _, endpoint := range n.endpoints {
		if endpoint == addr {
			return true
		}
	}

	return false
}

// ----------------------------------------------------------------------------

type versionReq struct {
	Version    int
	BestHeight int
	FromAddr   string
}

type getBlocksReq struct {
	FromAddr string
}

type addrReq struct {
	Values []string
}

type blockReq struct {
	FromAddr string
	Block    []byte
}

type getDataReq struct {
	FromAddr string
	Type     string
	ID       []byte
}

type invReq struct {
	FromAddr string
	Kind     string
	Values   [][]byte
}

type txReq struct {
	FromAddr string
	Tx       []byte
}

func (n *Server) sendGetBlocks(addr string) {
	payload := encode(getBlocksReq{FromAddr: n.endpoint})
	req := append(cmdToBytes(GetBlocksCmd), payload...)

	n.send(addr, req)
}

func (n *Server) sendVersion(addr string) {
	bestHeight := n.bc.GetBestHeight()
	payload := encode(versionReq{
		Version:    nodeVersion,
		BestHeight: bestHeight,
		FromAddr:   n.endpoint,
	})
	req := append(cmdToBytes(VersionCmd), payload...)

	n.send(addr, req)
}

func (n *Server) sendAddr(addr string) {
	v := addrReq{Values: n.endpoints}
	v.Values = append(v.Values, n.endpoint)
	payload := encode(v)
	req := append(cmdToBytes(AddrCmd), payload...)

	n.send(addr, req)
}

func (n *Server) sendBlock(addr string, block *Block) {
	v := blockReq{
		FromAddr: n.endpoint,
		Block:    block.Serialize(),
	}
	payload := encode(v)
	req := append(cmdToBytes(BlockCmd), payload...)

	n.send(addr, req)
}

func (n *Server) sendGetData(addr, kind string, id []byte) {
	payload := encode(getDataReq{
		FromAddr: n.endpoint,
		Type:     kind,
		ID:       id,
	})
	req := append(cmdToBytes(GetDataCmd), payload...)

	n.send(addr, req)
}

func (n *Server) sendInv(addr, kind string, values [][]byte) {
	v := invReq{
		FromAddr: n.endpoint,
		Kind:     kind,
		Values:   values,
	}
	payload := encode(v)
	req := append(cmdToBytes(InvCmd), payload...)

	n.send(addr, req)
}

func (n *Server) sendTx(addr string, tx *Transaction) {
	v := txReq{
		FromAddr: n.endpoint,
		Tx:       tx.Serialize(),
	}
	payload := encode(v)
	req := append(cmdToBytes(TxCmd), payload...)

	n.send(addr, req)
}

func (n *Server) SendTx(tx *Transaction) {
	n.sendTx(n.endpoints[0], tx)
}

func (n *Server) fetchBlocks() {
	for _, endpoint := range n.endpoints {
		n.sendGetBlocks(endpoint)
	}
}

func (n *Server) send(endpoint string, v []byte) {
	conn, err := net.Dial("tcp", endpoint)
	if err != nil {
		log.Printf("%s is not available\n", endpoint)
		var nEndpoints []string
		for _, node := range n.endpoints {
			if endpoint != node {
				nEndpoints = append(nEndpoints, node)
			}
		}
		n.endpoints = nEndpoints

		return
	}
	defer conn.Close()

	_, _ = io.Copy(conn, bytes.NewReader(v))
}

// ----------------------------------------------------------------------------

func (n *Server) handleVersion(v []byte) {
	var buf bytes.Buffer
	var payload versionReq

	buf.Write(v[cmdLength:])
	err := gob.NewDecoder(&buf).Decode(&payload)
	if err != nil {
		log.Println(err)
		return
	}

	innerBestHeight := n.bc.GetBestHeight()
	foreignerBestHeight := payload.BestHeight

	if innerBestHeight < foreignerBestHeight {
		n.sendGetBlocks(payload.FromAddr)
	}
	if innerBestHeight > foreignerBestHeight {
		n.sendVersion(payload.FromAddr)
	}

	if !n.hasEndpoint(payload.FromAddr) {
		n.endpoints = append(n.endpoints, payload.FromAddr)
	}
}

func (n *Server) handleAddr(v []byte) {
	var buf bytes.Buffer
	var payload addrReq

	buf.Write(v[cmdLength:])
	err := gob.NewDecoder(&buf).Decode(&payload)
	if err != nil {
		log.Println(err)
		return
	}

	n.endpoints = append(n.endpoints, payload.Values...)
	log.Printf("Threr are %d known nodes now!\n", len(n.endpoints))

	n.fetchBlocks()
}

func (n *Server) handleBlock(v []byte) {
	var buf bytes.Buffer
	var payload blockReq

	buf.Write(v[cmdLength:])
	_ = gob.NewDecoder(&buf).Decode(&payload)

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	log.Printf("Receive a new block")
	_ = n.bc.Submit(block)
	log.Printf("Added block %x\n", block.Hash)

	if len(n.blockInTransit) > 0 {
		blockHash := n.blockInTransit[0]
		n.sendGetData(payload.FromAddr, "block", blockHash)
		n.blockInTransit = n.blockInTransit[1:]
	} else {
		UtxoSet := NewUTXOSet(n.bc)
		_ = UtxoSet.Rebuild()
	}
}

func (n *Server) handleInv(v []byte) {
	var buf bytes.Buffer
	var payload invReq

	buf.Write(v[cmdLength:])
	err := gob.NewDecoder(&buf).Decode(&payload)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("Received inventroty with %d %s\n", len(payload.Values), payload.Kind)

	switch payload.Kind {
	case "block":
		{
			n.blockInTransit = payload.Values
			blockHash := payload.Values[0]

			n.sendGetData(payload.FromAddr, "block", blockHash)

			var newInTransit [][]byte
			for _, b := range n.blockInTransit {
				if bytes.Compare(b, blockHash) != 0 {
					newInTransit = append(newInTransit, b)
				}
			}

			n.blockInTransit = newInTransit
		}
	case "tx":
		{
			txID := payload.Values[0]
			if n.mempool[hex.EncodeToString(txID)].ID == nil {
				n.sendGetData(payload.FromAddr, "tx", txID)
			}
		}
	default:
	}
}

func (n *Server) handleGetBlocks(v []byte) {
	var buf bytes.Buffer
	var payload getBlocksReq

	buf.Write(v[cmdLength:])
	err := gob.NewDecoder(&buf).Decode(&payload)
	if err != nil {
		log.Println(err)
		return
	}

	blocks := n.bc.GetBlockHashes()
	n.sendInv(payload.FromAddr, "block", blocks)
}

func (n *Server) handleGetData(v []byte) {
	var buf bytes.Buffer
	var payload getDataReq

	buf.Write(v[cmdLength:])
	err := gob.NewDecoder(&buf).Decode(&payload)
	if err != nil {
		log.Println(err)
		return
	}

	switch payload.Type {
	case "block":
		{
			block, err := n.bc.getBlockByKey(payload.ID)
			if err != nil {
				log.Printf("block hash id: %v, err: %v", payload.ID, err)
				return
			}

			n.sendBlock(payload.FromAddr, block)
		}
	case "tx":
		{
			txID := hex.EncodeToString(payload.ID)
			tx := n.mempool[txID]

			n.sendTx(payload.FromAddr, &tx)
		}
	default:
	}
}

func (n *Server) handleTx(v []byte) {
	var buf bytes.Buffer
	var payload txReq

	buf.Write(v[cmdLength:])
	err := gob.NewDecoder(&buf).Decode(&payload)
	if err != nil {
		log.Println(err)
		return
	}
	txData := payload.Tx
	tx := DeserializeTx(txData)

	fmt.Println(n.endpoints)

	if n.endpoint == n.endpoints[0] {
		for _, endpoint := range n.endpoints {
			if endpoint != n.endpoint && endpoint != payload.FromAddr {
				n.sendInv(endpoint, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(n.mempool) >= 2 && len(n.MinerAddress) > 0 {
		mineTx:
			var txs []*Transaction

			for id := range n.mempool {
				tx := n.mempool[id]
				if n.bc.VerifyTx(&tx) {
					txs = append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				log.Println("All transactions are invalid, Waiting for new ones....")
				return
			}

			cTx := NewCoinbaseTx(n.MinerAddress, "")
			txs = append(txs, cTx)

			nBlock, _ := n.bc.Mine(txs)
			if err != nil {
				log.Println(err)
				return
			}

			UtxoSet := NewUTXOSet(n.bc)
			_ = UtxoSet.Rebuild()

			log.Println("New block is mined")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(n.mempool, txID)
			}

			for _, endpoint := range n.endpoints {
				if endpoint != n.endpoint {
					n.sendInv(endpoint, "block", [][]byte{nBlock.Hash})
				}
			}

			if len(n.mempool) > 0 {
				goto mineTx
			}
		}
	}
}

// ----------------------------------------------------------------------------

func encode(v interface{}) []byte {
	var buf bytes.Buffer
	_ = gob.NewEncoder(&buf).Encode(v)

	return buf.Bytes()
}

func cmdToBytes(command string) []byte {
	var buf [cmdLength]byte
	for i, c := range command {
		buf[i] = byte(c)
	}

	return buf[:]
}

func bytesToCmd(buf []byte) string {
	var cmd []byte
	for _, b := range buf {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}

	return fmt.Sprintf("%s", cmd)
}
