/*
 *
 *    This file is part of go-palletone.
 *    go-palletone is free software: you can redistribute it and/or modify
 *    it under the terms of the GNU General Public License as published by
 *    the Free Software Foundation, either version 3 of the License, or
 *    (at your option) any later version.
 *    go-palletone is distributed in the hope that it will be useful,
 *    but WITHOUT ANY WARRANTY; without even the implied warranty of
 *    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 *    GNU General Public License for more details.
 *    You should have received a copy of the GNU General Public License
 *    along with go-palletone.  If not, see <http://www.gnu.org/licenses/>.
 * /
 *
 *  * @author PalletOne core developer <dev@pallet.one>
 *  * @date 2018
 *
 */

package core

import (
	"fmt"
	"github.com/palletone/go-palletone/common/log"
	"github.com/palletone/go-palletone/dag/errors"
	"sort"
)

type UtxoInterface interface {
	GetAmount() uint64
}
type Utxos []UtxoInterface

func (c Utxos) Len() int {
	return len(c)
}
func (c Utxos) Swap(i, j int) {
	c[i], c[j] = c[j], c[i]
}
func (c Utxos) Less(i, j int) bool {
	return c[i].GetAmount() < c[j].GetAmount()
}

func find_min(utxos []UtxoInterface) UtxoInterface {
	amout := utxos[0].GetAmount()
	min_utxo := utxos[0]
	for _, utxo := range utxos {
		if utxo.GetAmount() < amout {
			min_utxo = utxo
			amout = min_utxo.GetAmount()
		}
	}
	return min_utxo
}
func Select_utxo_Greedy(utxos Utxos, amount uint64) (Utxos, uint64, error) {
	var greaters Utxos
	var lessers Utxos
	var taken_utxo Utxos
	var accum uint64
	var change uint64
	logPickedAmt := ""
	for _, utxo := range utxos {
		if utxo.GetAmount() > amount {
			greaters = append(greaters, utxo)
		}
		if utxo.GetAmount() < amount {
			lessers = append(lessers, utxo)
		}
	}
	var min_greater UtxoInterface
	if len(greaters) > 0 {
		min_greater = find_min(greaters)
		change = min_greater.GetAmount() - amount
		logPickedAmt += fmt.Sprintf("%d,", min_greater.GetAmount())
		taken_utxo = append(taken_utxo, min_greater)

	} else if len(greaters) == 0 && len(lessers) > 0 {
		sort.Sort(Utxos(lessers))
		for _, utxo := range lessers {
			accum += utxo.GetAmount()
			logPickedAmt += fmt.Sprintf("%d,", utxo.GetAmount())
			taken_utxo = append(taken_utxo, utxo)
			if accum >= amount {
				change = accum - amount
				break
			}
		}
		if accum < amount {
			return nil, 0, errors.New("Amount Not Enough to pay")
		}
	}
	log.Debugf("Pickup count[%d] utxos, each amount:%s to match wanted amount:%d", len(taken_utxo), logPickedAmt, amount)
	return taken_utxo, change, nil
}
