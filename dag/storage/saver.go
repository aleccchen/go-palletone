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

package storage

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"math/big"

	"github.com/palletone/go-palletone/common"
	// "github.com/palletone/go-palletone/common/hexutil"
	"github.com/palletone/go-palletone/common/ptndb"
	"github.com/palletone/go-palletone/common/rlp"
	"github.com/palletone/go-palletone/dag/constants"
	"github.com/palletone/go-palletone/dag/dagconfig"
	"github.com/palletone/go-palletone/dag/modules"
	"github.com/palletone/go-palletone/tokenengine"
)

var (
	AssocUnstableUnits map[string]modules.Joint
)

func SaveJoint(db ptndb.Database, objJoint *modules.Joint, onDone func()) (err error) {
	if objJoint.Unsigned != "" {
		return errors.New(objJoint.Unsigned)
	}
	obj_unit := objJoint.Unit
	obj_unit_byte, _ := json.Marshal(obj_unit)

	if err = db.Put(append(UNIT_PREFIX, obj_unit.Hash().Bytes()...), obj_unit_byte); err != nil {
		return
	}
	// add key in  unit_keys
	log.Println("add unit key:", string(UNIT_PREFIX)+obj_unit.Hash().String(), AddUnitKeys(db, string(UNIT_PREFIX)+obj_unit.Hash().String()))

	if dagconfig.SConfig.Blight {
		// save  update utxo , message , transaction

	}

	if onDone != nil {
		onDone()
	}
	return
}

/**
key: [HEADER_PREFIX][chain index number]_[chain index]_[unit hash]
value: unit header rlp encoding bytes
*/
// save header
func (dagdb *DagDatabase) SaveHeader(uHash common.Hash, h *modules.Header) error {
	encNum := encodeBlockNumber(h.Number.Index)
	key := append(HEADER_PREFIX, encNum...)
	key = append(key, h.Number.Bytes()...)
	return StoreBytes(dagdb.db, append(key, uHash.Bytes()...), h)
	//key := fmt.Sprintf("%s%v_%s_%s", HEADER_PREFIX, h.Number.Index, h.Number.String(), uHash.Bytes())
	//return StoreBytes(db, []byte(key), h)
}

//這是通過modules.ChainIndex存儲hash
func (dagdb *DagDatabase) SaveNumberByHash(uHash common.Hash, number modules.ChainIndex) error {
	return StoreBytes(dagdb.db, append(UNIT_HASH_NUMBER_Prefix, uHash.Bytes()...), number)
}

//這是通過hash存儲modules.ChainIndex
func (dagdb *DagDatabase) SaveHashByNumber(uHash common.Hash, number modules.ChainIndex) error {
	i := 0
	if number.IsMain {
		i = 1
	}
	key := fmt.Sprintf("%s_%s_%d_%d", UNIT_NUMBER_PREFIX, number.AssetID.String(), i, number.Index)
	//fmt.Println("SaveHashByNumber=[]byte(key)=>",[]byte(key))
	//fmt.Println("====",uHash)
	return StoreBytes(dagdb.db, []byte(key), uHash)
}

// height and assetid can get a unit key.
func (dagdb *DagDatabase) SaveUHashIndex(cIndex modules.ChainIndex, uHash common.Hash) error {
	key := fmt.Sprintf("%s_%s_%d", UNIT_NUMBER_PREFIX, cIndex.AssetID.String(), cIndex.Index)
	return Store(dagdb.db, key, uHash.Bytes())
}

/**
key: [BODY_PREFIX][unit hash]
value: all transactions hash set's rlp encoding bytes
*/
func (dagdb *DagDatabase) SaveBody(unitHash common.Hash, txsHash []common.Hash) error {
	// db.Put(append())
	return StoreBytes(dagdb.db, append(BODY_PREFIX, unitHash.Bytes()...), txsHash)
}

func (dagdb *DagDatabase) GetBody(unitHash common.Hash) ([]common.Hash, error) {
	data, err := dagdb.db.Get(append(BODY_PREFIX, unitHash.Bytes()...))
	if err != nil {
		return nil, err
	}
	var txHashs []common.Hash
	if err := rlp.DecodeBytes(data, &txHashs); err != nil {
		return nil, err
	}
	return txHashs, nil
}

func (dagdb *DagDatabase) SaveTransactions(txs *modules.Transactions) error {
	key := fmt.Sprintf("%s%s", TRANSACTIONS_PREFIX, txs.Hash())
	return Store(dagdb.db, key, *txs)
}

/**
key: [TRANSACTION_PREFIX][tx hash]
value: transaction struct rlp encoding bytes
*/
func (dagdb *DagDatabase) SaveTransaction(tx *modules.Transaction) error {
	// save transaction
	if err := StoreBytes(dagdb.db, append(TRANSACTION_PREFIX, tx.TxHash.Bytes()...), tx); err != nil {
		return err
	}

	if err := StoreBytes(dagdb.db, append(Transaction_Index, tx.TxHash.Bytes()...), tx); err != nil {
		return err
	}
	dagdb.updateAddrTransactions(tx.Address().String(), tx.TxHash)
	// store output by addr
	for i, msg := range tx.TxMessages {
		payload, ok := msg.Payload.(modules.PaymentPayload)
		if ok {
			for _, output := range payload.Output {
				//  pkscript to addr
				addr, _ := tokenengine.GetAddressFromScript(output.PkScript[:])
				dagdb.saveOutputByAddr(addr.String(), tx.TxHash, i, *output)
			}
		}
	}

	return nil
}
func (dagdb *DagDatabase) updateAddrTransactions(addr string, hash common.Hash) error {
	if hash == (common.Hash{}) {
		return errors.New("empty tx hash.")
	}
	hashs := make([]common.Hash, 0)
	data, err := dagdb.db.Get(append(AddrTransactionsHash_Prefix, []byte(addr)...))
	if err != nil {
		if err.Error() != "leveldb: not found" {
			return err
		} else { // first store the addr
			hashs = append(hashs, hash)
			if err := StoreBytes(dagdb.db, append(AddrTransactionsHash_Prefix, []byte(addr)...), hashs); err != nil {
				return err
			}
			return nil
		}
	}
	if err := rlp.DecodeBytes(data, hashs); err != nil {
		return err
	}
	hashs = append(hashs, hash)
	if err := StoreBytes(dagdb.db, append(AddrTransactionsHash_Prefix, []byte(addr)...), hashs); err != nil {
		return err
	}
	return nil
}
func (dagdb *DagDatabase) saveOutputByAddr(addr string, hash common.Hash, msgindex int, output modules.Output) error {
	if hash == (common.Hash{}) {
		return errors.New("empty tx hash.")
	}
	key := append(AddrOutput_Prefix, []byte(addr)...)
	key = append(key, hash.Bytes()...)
	if err := StoreBytes(dagdb.db, append(key, new(big.Int).SetInt64(int64(msgindex)).Bytes()...), output); err != nil {
		return err
	}
	return nil
}
func (dagdb *DagDatabase) SaveTxLookupEntry(unit *modules.Unit) error {
	for i, tx := range unit.Transactions() {
		in := modules.TxLookupEntry{
			UnitHash:  unit.Hash(),
			UnitIndex: unit.NumberU64(),
			Index:     uint64(i),
		}
		data, err := rlp.EncodeToBytes(in)
		if err != nil {
			return err
		}
		if err := StoreBytes(dagdb.db, append(LookupPrefix, tx.TxHash.Bytes()...), data); err != nil {
			return err
		}
	}
	return nil
}

// encodeBlockNumber encodes a block number as big endian uint64
func encodeBlockNumber(number uint64) []byte {
	enc := make([]byte, 8)
	binary.BigEndian.PutUint64(enc, number)
	return enc
}

func GetUnitKeys(db ptndb.Database) []string {
	var keys []string
	if keys_byte, err := db.Get([]byte("array_units")); err != nil {
		log.Println("get units error:", err)
	} else {
		if err := rlp.DecodeBytes(keys_byte[:], &keys); err != nil {
			log.Println("error:", err)
		}
	}
	return keys
}
func AddUnitKeys(db ptndb.Database, key string) error {
	keys := GetUnitKeys(db)
	if len(keys) <= 0 {
		return errors.New("null keys.")
	}
	for _, v := range keys {

		if v == key {
			return errors.New("key is already exist.")
		}
	}
	keys = append(keys, key)

	return Store(db, "array_units", keys)
}
func ConvertBytes(val interface{}) (re []byte) {
	var err error
	if re, err = json.Marshal(val); err != nil {
		log.Println("json.marshal error:", err)
	}
	return re[:]
}

func IsGenesisUnit(unit string) bool {
	return unit == constants.GENESIS_UNIT
}

func GetKeysWithTag(db ptndb.Database, tag string) []string {
	var keys []string

	if keys_byte, err := db.Get([]byte(tag)); err != nil {
		log.Println("get keys error:", err)
	} else {
		if err := json.Unmarshal(keys_byte, &keys); err != nil {
			log.Println("error:", err)
		}
	}
	return keys
}
func AddKeysWithTag(db ptndb.Database, key, tag string) error {
	keys := GetKeysWithTag(db, tag)
	if len(keys) <= 0 {
		return errors.New("null keys.")
	}
	log.Println("keys:=", keys)
	for _, v := range keys {
		if v == key {
			return errors.New("key is already exist.")
		}
	}
	keys = append(keys, key)

	if err := db.Put([]byte(tag), ConvertBytes(keys)); err != nil {
		return err
	}
	return nil

}

func SaveContract(db ptndb.Database, contract *modules.Contract) (common.Hash, error) {
	if common.EmptyHash(contract.CodeHash) {
		contract.CodeHash = rlp.RlpHash(contract.Code)
	}
	// key = cs+ rlphash(contract)
	if common.EmptyHash(contract.Id) {
		ids := rlp.RlpHash(contract)
		if len(ids) > len(contract.Id) {
			id := ids[len(ids)-common.HashLength:]
			copy(contract.Id[common.HashLength-len(id):], id)
		} else {
			//*contract.Id = new(common.Hash)
			copy(contract.Id[common.HashLength-len(ids):], ids[:])
		}

	}

	return contract.Id, StoreBytes(db, append(CONTRACT_PTEFIX, contract.Id[:]...), contract)
}

//  get  unit chain version
// GetUnitChainVersion reads the version number from db.
func GetUnitChainVersion(db ptndb.Database) int {
	var vsn uint

	enc, _ := db.Get([]byte("UnitchainVersion"))
	rlp.DecodeBytes(enc, &vsn)
	return int(vsn)
}

// SaveUnitChainVersion writes vsn as the version number to db.
func SaveUnitChainVersion(db ptndb.Database, vsn int) error {
	enc, _ := rlp.EncodeToBytes(uint(vsn))
	return db.Put([]byte("UnitchainVersion"), enc)
}

/**
保存合约属性信息
To save contract
*/
func SaveContractState(db ptndb.Database, prefix []byte, id []byte, name string, value interface{}, version *modules.StateVersion) error {
	key := fmt.Sprintf("%s%s^*^%s^*^%s",
		prefix,
		id,
		name,
		version.String())
	if err := Store(db, key, value); err != nil {
		log.Println("Save contract template", "error", err.Error())
		return err
	}
	return nil
}
func (statedb *StateDatabase) SaveContractTemplate(templateId []byte, bytecode []byte, version string) error {
	// key:[CONTRACT_TPL][Template id]_bytecode_[template version]
	// key := fmt.Sprintf("%s%s^*^bytecode^*^%s",
	// 	CONTRACT_TPL,
	// 	hexutil.Encode(templateId),
	// 	version)

	key := append(CONTRACT_TPL, templateId...)
	key = append(key, []byte("^*^bytecode^*^")...)
	key = append(key, []byte(version)...)
	if err := statedb.db.Put(key, bytecode); err != nil {
		return err
	}
	return nil
}
func (statedb *StateDatabase) SaveContractTemplateState(id []byte, name string, value interface{}, version *modules.StateVersion) error {
	return SaveContractState(statedb.db, CONTRACT_TPL, id, name, value, version)
}
func (statedb *StateDatabase) SaveContractState(id []byte, name string, value interface{}, version *modules.StateVersion) error {
	return SaveContractState(statedb.db, CONTRACT_STATE_PREFIX, id, name, value, version)
}
func (statedb *StateDatabase) DeleteState(key []byte) error {
	return statedb.db.Delete(key)
}

func (utxodb *UtxoDatabase) SaveUtxoEntity(key []byte, utxo modules.Utxo) error {
	return StoreBytes(utxodb.db, key, utxo)
}

func (utxodb *UtxoDatabase) DeleteUtxo(key []byte) error {
	return utxodb.db.Delete(key)
}
func (idxdb *IndexDatabase) SaveIndexValue(key []byte, value interface{}) error {
	return StoreBytes(idxdb.db, key, value)
}
func (statedb *StateDatabase) SaveAssetInfo(assetInfo *modules.AssetInfo) error {
	key := assetInfo.Tokey()
	return StoreBytes(statedb.db, key, assetInfo)
}
