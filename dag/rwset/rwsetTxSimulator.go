/*
	This file is part of go-palletone.
	go-palletone is free software: you can redistribute it and/or modify
	it under the terms of the GNU General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.
	go-palletone is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU General Public License for more details.
	You should have received a copy of the GNU General Public License
	along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
*/

/*
 * @author PalletOne core developers <dev@pallet.one>
 * @date 2018
 */

package rwset

import (
	"encoding/json"
	"errors"
	"github.com/palletone/go-palletone/dag"
	"github.com/palletone/go-palletone/dag/modules"
	"time"
)

type RwSetTxSimulator struct {
	txid                    string
	rwsetBuilder            *RWSetBuilder
	dag                     dag.IDag
	writePerformed          bool
	pvtdataQueriesPerformed bool
	doneInvoked             bool
}

type VersionedValue struct {
	Value   []byte
	Version *Version
}

func NewBasedTxSimulator(idag dag.IDag, txid string) *RwSetTxSimulator {
	rwsetBuilder := NewRWSetBuilder()

	logger.Debugf("constructing new tx simulator txid = [%s]", txid)
	return &RwSetTxSimulator{txid, rwsetBuilder, idag, false, false, false}
}

// GetState implements method in interface `ledger.TxSimulator`
func (s *RwSetTxSimulator) GetState(contractid []byte, ns string, key string) ([]byte, error) {
	//testValue := []byte("abc")
	if err := s.CheckDone(); err != nil {
		return nil, err
	}
	//TODO
	stateValue := modules.StateValue{
		Asset: modules.Asset{
			AssetId:  modules.PTNCOIN,
			UniqueId: modules.PTNCOIN,
			ChainId:  uint64(1),
		},
		Amount: 100000,
		Time:   time.Now(),
		Extra:  "hello",
	}
	stateValueByte, _ := json.Marshal(stateValue)
	return stateValueByte, nil

	//fmt.Println("GetState(contractid []byte, ns string, key string)===>>>\n\n", contractid, ns, key)
	//return []byte("1000"), nil
	//TODO Devin
	ver, val := s.dag.GetContractState(contractid, key)
	//TODO 这里证明数据库里面没有该账户信息，需要返回nil,nil
	if val == nil {
		logger.Errorf("get value from db[%s] failed", ns)
		return nil, nil
		//errstr := fmt.Sprintf("GetContractState [%s]-[%s] failed", ns, key)
		//		//return nil, errors.New(errstr)
	}
	//val, ver := decomposeVersionedValue(versionedValue)
	if s.rwsetBuilder != nil {
		s.rwsetBuilder.AddToReadSet(ns, key, ver)
	}

	logger.Debugf("RW:GetState,ns[%s]--key[%s]---value[%s]", ns, key, val)

	//todo change.
	//return testValue, nil
	return val, nil
}

func (s *RwSetTxSimulator) SetState(ns string, key string, value []byte) error {
	//logger.Debugf("RW:SetState,ns[%s]--key[%s]---value[%s]", ns, key, value)
	//fmt.Println("SetState(ns string, key string, value []byte)===>>>\n\n", ns, key, value)
	stateValue := &modules.StateValue{}
	_ = json.Unmarshal(value, stateValue)
	//fmt.Printf("llllllll   %#v\n", stateValue)
	if err := s.CheckDone(); err != nil {
		return err
	}
	if s.pvtdataQueriesPerformed {
		return errors.New("pvtdata Queries Performed")
	}
	//todo ValidateKeyValue
	s.rwsetBuilder.AddToWriteSet(ns, key, value)
	return nil
}

// DeleteState implements method in interface `ledger.TxSimulator`
func (s *RwSetTxSimulator) DeleteState(ns string, key string) error {
	//fmt.Println("DeleteState(ns string, key string)===>\n\n", ns, key)
	return s.SetState(ns, key, nil)
}

func (s *RwSetTxSimulator) GetRwData(ns string) (map[string]*KVRead, map[string]*KVWrite, error) {
	var rd map[string]*KVRead
	var wt map[string]*KVWrite

	logger.Infof("ns=%s", ns)

	if s.rwsetBuilder != nil {
		if s.rwsetBuilder.pubRwBuilderMap != nil {
			if s.rwsetBuilder.pubRwBuilderMap[ns] != nil {
				if s.rwsetBuilder.pubRwBuilderMap[ns].readMap != nil {
					rd = s.rwsetBuilder.pubRwBuilderMap[ns].readMap
				}
				if s.rwsetBuilder.pubRwBuilderMap[ns].writeMap != nil {
					wt = s.rwsetBuilder.pubRwBuilderMap[ns].writeMap
				}
				pubRwBuilderMap, ok := s.rwsetBuilder.pubRwBuilderMap[ns]
				if ok {
					rd = pubRwBuilderMap.readMap
					wt = pubRwBuilderMap.writeMap
				} else {
					rd = nil
					wt = nil
				}
			}
		}
	}

	return rd, wt, nil
}

//get all dag
func (s *RwSetTxSimulator) GetContractStatesById(contractid []byte) (map[string]*modules.ContractStateValue, error) {
	return s.dag.GetContractStatesById(contractid)
}

func (h *RwSetTxSimulator) CheckDone() error {
	if h.doneInvoked {
		return errors.New("This instance should not be used after calling Done()")
	}
	return nil
}

func decomposeVersionedValue(versionedValue *VersionedValue) ([]byte, *Version) {
	var value []byte
	var ver *Version
	if versionedValue != nil {
		value = versionedValue.Value
		ver = versionedValue.Version
	}
	return value, ver
}

func (h *RwSetTxSimulator) Done() {
	if h.doneInvoked {
		return
	}
	//todo
}
func (h *RwSetTxSimulator) GetTxSimulationResults() ([]byte, error) {

	return nil, nil
}
func (s *RwSetTxSimulator) GetTokenBalance(ns string, addr string, asset *modules.Asset) (map[modules.Asset]uint64, error) {

	var utxos map[modules.OutPoint]*modules.Utxo
	if asset == nil {
		utxos, _ = s.dag.GetAddrUtxos(addr)
	} else {
		utxos, _ = s.dag.GetAddr1TokenUtxos(addr, asset)
	}
	return convertUtxo2Balance(utxos), nil
}
func convertUtxo2Balance(utxos map[modules.OutPoint]*modules.Utxo) map[modules.Asset]uint64 {
	result := map[modules.Asset]uint64{}
	for _, v := range utxos {
		if val, ok := result[*v.Asset]; ok {
			result[*v.Asset] = val + v.Amount
		} else {
			result[*v.Asset] = v.Amount
		}
	}
	return result
}
func (s *RwSetTxSimulator) PayOutToken(ns string, address string, token *modules.Asset, amount uint64, lockTime uint32) error {
	s.rwsetBuilder.AddTokenPayOut(ns, address, token, amount, lockTime)
	return nil
}
func (s *RwSetTxSimulator) GetPayOutData(ns string) ([]*modules.TokenPayOut, error) {
	return s.rwsetBuilder.GetTokenPayOut(ns), nil
}
