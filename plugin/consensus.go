// Copyright 2019 The Fractal Team Authors
// This file is part of the fractal project.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package plugin

import (
	"errors"
	"fmt"
	"math/big"
	"sort"
	"time"

	"github.com/fractalplatform/fractal/state"
	"github.com/fractalplatform/fractal/types"
	"github.com/fractalplatform/fractal/utils/rlp"
)

// 1. 基准时间轴
// 2. 严格时间轴
// 3. 超过2/3不出块，则所有人出块，并包含惩罚交易提议(时间提后?) //出块人数过低，则链停止
// 4. 后续出块人依次出块并包含投票(交易?)

// 1. 缺块空时间窗(VDF?)
// 2. 总值 = t * p ; max(p) = 100
// 3. 支持合约？

const (
	ConsensusKey     = "consensus"
	CandidateKey     = "candidates"
	LackBlock        = "lackblock"
	CandidateInfoKey = "info_"
	// account
	MinerAssetID = uint64(0)
)

var (
	minLockAmount = big.NewInt(1)
	maxCandidates = 32
	maxMiner      = uint64(21)
	maxWeight     = uint64(100)
	minWeight     = 0
	blockDuration = uint64(3)
	MinerAccount  string
	genesisTime   uint64
)

type CandidateInfo struct {
	OwnerAccount   string
	SignAccount    string
	RegisterNumber uint64
	Weight         uint64
	Balance        *big.Int
	Skip           bool
}

func (info *CandidateInfo) IncWeight() uint64 {
	if info.Weight >= maxWeight {
		return maxWeight
	}
	info.Weight++
	return info.Weight
}

func (info *CandidateInfo) DecWeight() uint64 {
	info.Weight = info.Weight * 90 / 100
	return info.Weight
}

func (info *CandidateInfo) WeightedSum() *big.Int {
	z := big.NewInt(int64(info.Weight))
	return z.Mul(info.Balance, z)
}

func (info *CandidateInfo) Store(stateDB *state.StateDB) {
	b, _ := rlp.EncodeToBytes(info)
	stateDB.Put(ConsensusKey, CandidateInfoKey+info.OwnerAccount, b)
}

func (info *CandidateInfo) Load(stateDB *state.StateDB, owner string) {
	b, _ := stateDB.Get(ConsensusKey, CandidateInfoKey+owner)
	rlp.DecodeBytes(b, info)
}

type Candidates struct {
	listSort []string
	info     map[string]*CandidateInfo
}

func (candidates *Candidates) Len() int {
	return len(candidates.listSort)
}

func (candidates *Candidates) Less(i, j int) bool {
	info_i := candidates.info[candidates.listSort[i]]
	info_j := candidates.info[candidates.listSort[j]]
	return info_i.WeightedSum().Cmp(info_j.WeightedSum()) < 0
}

func (candidates *Candidates) Swap(i, j int) {
	candidates.listSort[i], candidates.listSort[j] = candidates.listSort[j], candidates.listSort[i]
}

func (candidates *Candidates) sort() {
	sort.Sort(candidates)
}

func (candidates *Candidates) insert(account string, newinfo *CandidateInfo) (bool, *CandidateInfo) {
	if candidates.info[account] != nil {
		return false, nil
	}
	candidates.info[account] = newinfo
	candidates.listSort = append(candidates.listSort, account)
	candidates.sort()
	if candidates.Len() > maxCandidates {
		replaced := candidates.listSort[candidates.Len()-1]
		info := candidates.remove(replaced)
		if replaced != account {
			return true, info // return the loser
		}
	}
	return true, nil // no one out
}

func (candidates *Candidates) remove(account string) *CandidateInfo {
	info, exist := candidates.info[account]
	if !exist {
		return nil
	}
	for i, name := range candidates.listSort {
		if name == account {
			copy(candidates.listSort[i:], candidates.listSort[i+1:])
			candidates.listSort = candidates.listSort[:candidates.Len()-1]
			delete(candidates.info, account)
			return info
		}
	}
	return nil // never goto here
}

type Consensus struct {
	isInit        bool
	BlockGasLimit uint64
	LackBlock     uint64
	candidates    *Candidates
	minerIndex    uint64
	parent        *types.Header
	stateDB       *state.StateDB
}

func NewConsensus(stateDB *state.StateDB) *Consensus {
	c := &Consensus{
		parent:  nil,
		stateDB: stateDB,
		candidates: &Candidates{
			info: make(map[string]*CandidateInfo),
		},
	}
	c.loadCandidates()
	c.loadLackBlock()
	for i, n := range c.candidates.listSort {
		info := &CandidateInfo{}
		info.Load(c.stateDB, n)
		c.candidates.info[n] = info
		if n == c.parent.Coinbase {
			c.minerIndex = uint64(i)
		}
	}
	return c
}

func (c *Consensus) initRequrie() {
	if !c.isInit {
		panic("Consensus need Init() before call")
	}
}

func (c *Consensus) Init(genesisTime uint64, genesisAccount string, parent *types.Header) {
	if len(MinerAccount) == 0 {
		MinerAccount = genesisAccount
		genesisTime = genesisTime
	}
	c.parent = parent
	c.isInit = true
}

func (c *Consensus) timeSlot(n uint64) uint64 {
	ontime := genesisTime + (c.parent.Number+c.LackBlock+n)*blockDuration
	return ontime
}

func (c *Consensus) miner(n uint64) string {
	numMiner := maxMiner
	if numMiner > uint64(c.candidates.Len()) {
		numMiner = uint64(c.candidates.Len())
	}
	index := (c.minerIndex + n) % numMiner
	return c.candidates.listSort[index]
}

func (c *Consensus) storeLackBlock() {
	b, _ := rlp.EncodeToBytes(c.LackBlock)
	c.stateDB.Put(ConsensusKey, LackBlock, b)
}
func (c *Consensus) loadLackBlock() {
	b, _ := c.stateDB.Get(ConsensusKey, LackBlock)
	rlp.DecodeBytes(b, &c.LackBlock)
}

func (c *Consensus) storeCandidates() {
	b, _ := rlp.EncodeToBytes(c.candidates.listSort)
	c.stateDB.Put(ConsensusKey, CandidateKey, b)
}

func (c *Consensus) loadCandidates() {
	b, _ := c.stateDB.Get(ConsensusKey, CandidateKey)
	rlp.DecodeBytes(b, c.candidates.listSort)
}

func (c *Consensus) removeCandidate(delCandidate string) (bool, *CandidateInfo) {
	if c.candidates.Len() == 0 {
		return false, nil // impossible?
	}
	info := c.candidates.remove(delCandidate)
	return info != nil, info
}

func (c *Consensus) pushCandidate(newCandidate string, lockAmount *big.Int) (bool, *CandidateInfo) {
	info := &CandidateInfo{
		SignAccount: newCandidate,
		Weight:      90,
		//		RegisterNumber: c.parent.Number + 1,
		Balance: big.NewInt(0).Set(lockAmount),
	}
	if c.parent != nil {
		info.RegisterNumber = c.parent.Number + 1
	}
	if info.WeightedSum().Cmp(minLockAmount) < 0 {
		return false, nil
	}
	return c.candidates.insert(newCandidate, info)
}

func (c *Consensus) nextMiner(miner string) int {
	now := uint64(time.Now().Unix())
	for i := 1; i <= c.candidates.Len(); i++ {
		nextTimeout := c.timeSlot(uint64(i))
		if now < nextTimeout {
			return i // current miner
		}
	}
	return -1
}

func (c *Consensus) searchMiner(miner string) int {
	for i := 1; i <= c.candidates.Len(); i++ {
		if miner == c.miner(uint64(i)) {
			return i
		}
	}
	return -1
}

func (c *Consensus) MineDelay(miner string) time.Duration {
	// just beta
	c.initRequrie()

	i := c.nextMiner(miner)
	if i < 0 {
		// wrong miner!
		return -1
	}
	nextMiner := c.miner(uint64(i))
	if nextMiner == miner {
		return 0
	}
	now := time.Now().Unix()
	return time.Duration(int64(c.timeSlot(uint64(i)))-now) * time.Second
}

func (c *Consensus) Prepare(miner string) *types.Header {
	// just beta
	c.initRequrie()

	minerIndex := c.searchMiner(miner)
	if minerIndex < 0 {
		return nil
	}
	blcoktime := c.timeSlot(uint64(minerIndex))
	for i := 1; i < minerIndex; i++ {
		skipMiner := c.miner(uint64(i))
		info := c.candidates.info[skipMiner]
		info.Skip = true // TODO: skip once
		info.DecWeight()
		info.Store(c.stateDB)
	}
	if minerIndex > 1 {
		c.LackBlock += uint64(minerIndex) - 1
		c.minerIndex += uint64(minerIndex) - 1
		c.storeLackBlock()
		c.storeCandidates()
	}
	return &types.Header{
		ParentHash: c.parent.Hash(),
		Number:     c.parent.Number + 1,
		Time:       blcoktime,
		Coinbase:   miner,
	}
}

func (c *Consensus) CallTx(action *types.Action, pm IPM) ([]byte, error) {
	// just beta
	c.initRequrie()
	if action.Recipient() != MinerAccount {
		return nil, fmt.Errorf("recipient must be %s", MinerAccount)
	}
	if action.Value().Sign() > 0 {
		if action.AssetID() != MinerAssetID {
			return nil, fmt.Errorf("assetID must be %d", MinerAssetID)
		}
		if err := pm.TransferAsset(action.Sender(), action.Recipient(), action.AssetID(), action.Value()); err != nil {
			return nil, err
		}
	}
	//TODO: return or lock balance
	var success bool
	var info *CandidateInfo
	if action.Value().Sign() > 0 {
		success, info = c.pushCandidate(action.Sender(), action.Value())
	} else {
		success, info = c.removeCandidate(action.Sender())
	}
	if success {
		c.storeCandidates()
		if info != nil {
			err := pm.TransferAsset(MinerAccount, info.OwnerAccount, MinerAssetID, info.Balance)
			fmt.Println("success2", err)
			return nil, err
		}
		fmt.Println("success!")
		return nil, nil
	}
	return nil, errors.New("wrong candidate")
}

// Finalize assembles the final block.
func (c *Consensus) Finalize(header *types.Header, txs []*types.Transaction, receipts []*types.Receipt) (*types.Block, error) {
	// just beta
	c.initRequrie()

	// info.Dec or Inc
	header.Root = c.stateDB.IntermediateRoot()
	return types.NewBlock(header, txs, receipts), nil
}

func (c *Consensus) Difficult(header *types.Header) int64 {
	// just beta
	c.initRequrie()

	if c.miner(0) == header.Coinbase {
		return int64(c.parent.Number) - int64(c.LackBlock)
	}
	return int64(c.parent.Number) - int64(c.LackBlock) - int64(c.searchMiner(header.Coinbase))
}

func (c *Consensus) Verify(header *types.Header, miner string) error {
	// just beta
	c.initRequrie()

	// TODO: verify header
	// 1. verify number
	if header.Number != c.parent.Number+1 {
		return fmt.Errorf("wrong block.Number, get %d want %d", header.Number, c.parent.Number+1)
	}
	// 2. verify Parent hash
	if header.ParentHash != c.parent.Hash() {
		return fmt.Errorf("wrong block.ParentHash, get %s want %s", header.ParentHash, c.parent.Hash())
	}
	// 3. verify miner
	minerIndex := c.searchMiner(miner)
	if minerIndex < 0 {
		return errors.New("can not find miner")
	}
	// 4. verify block time
	timeSlot := c.timeSlot(uint64(minerIndex))
	if header.Time != timeSlot {
		return fmt.Errorf("wrong block.Time, get %s want %s", header.Time, timeSlot)
	}
	now := time.Now().Unix()
	maxTime := uint64(now) + blockDuration*5
	if timeSlot > maxTime {
		return fmt.Errorf("wrong time slot, get %d want <=%d", timeSlot, maxTime)
	}
	// 5. verify weighSum?
	// 6. verify Sign?
	// 7. verify Version
	// 8. verify ExtData?
	return nil
}

func (c *Consensus) Seal(block *types.Block) (*types.Block, error) {
	// just beta
	c.initRequrie()

	return block, nil
}

func (c *Consensus) VerifySeal(header *types.Header) error {
	// just beta
	c.initRequrie()

	return errors.New("consensus don't support VerifySeal")
}
