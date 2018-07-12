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
	"errors"
)

type RwSetTxSimulator struct {
	txid                    string
	rwsetBuilder            *RWSetBuilder
	writePerformed          bool
	pvtdataQueriesPerformed bool
	doneInvoked             bool
}

type VersionedValue struct {
	Value   []byte
	Version *Version
}

func newBasedTxSimulator(txid string) (*RwSetTxSimulator, error) {
	rwsetBuilder := NewRWSetBuilder()
	logger.Debugf("constructing new tx simulator txid = [%s]", txid)
	return &RwSetTxSimulator{ txid, rwsetBuilder, false, false, false}, nil
}

// GetState implements method in interface `ledger.TxSimulator`
func (s *RwSetTxSimulator) GetState(ns string, key string) ([]byte, error) {
	if err := s.CheckDone(); err != nil {
		return nil, err
	}
	//get value from DB !!!
	var versionedValue *VersionedValue
	//versionedValue, err := db.GetState(ns, key)
	//if err != nil {
	//	return nil, err
	//}

	val, ver := decomposeVersionedValue(versionedValue)
	if s.rwsetBuilder != nil {
		s.rwsetBuilder.AddToReadSet(ns, key, ver)
	}

	return val, nil
}

func (s *RwSetTxSimulator) SetState(ns string, key string, value []byte) error {
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
	return s.SetState(ns, key, nil)
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
