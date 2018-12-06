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

package modules

import (
	"fmt"
	"io"

	"github.com/palletone/go-palletone/common"
	"github.com/palletone/go-palletone/common/rlp"
)

type transactionTemp struct {
	TxHash     common.Hash
	TxId       common.Hash
	TxMessages []messageTemp
}
type messageTemp struct {
	App  MessageType
	Data []byte
}

func (tx *Transaction) DecodeRLP(s *rlp.Stream) error {
	raw, err := s.Raw()
	if err != nil {
		return err
	}
	var txTemp transactionTemp
	rlp.DecodeBytes(raw, &txTemp)
	temp2Tx(txTemp, tx)
	//fmt.Println("Use DecodeRLP")
	return nil
}
func (tx *Transaction) EncodeRLP(w io.Writer) error {
	temp := tx2Temp(*tx)
	//fmt.Println("Use EncodeRLP")
	return rlp.Encode(w, temp)
}
func tx2Temp(tx Transaction) transactionTemp {
	temp := transactionTemp{TxHash: tx.TxHash, TxId: tx.TxId}

	for _, m := range tx.TxMessages {

		m1 := messageTemp{
			App: m.App,
		}
		m1.Data, _ = rlp.EncodeToBytes(m.Payload)
		temp.TxMessages = append(temp.TxMessages, m1)

	}
	return temp
}
func temp2Tx(temp transactionTemp, tx *Transaction) {
	tx.TxId = temp.TxId
	tx.TxHash = temp.TxHash

	for _, m := range temp.TxMessages {
		m1 := &Message{
			App: m.App,
		}
		if m.App == APP_PAYMENT {
			var pay PaymentPayload
			rlp.DecodeBytes(m.Data, &pay)
			m1.Payload = &pay
		} else if m.App == APP_TEXT {
			var text TextPayload
			rlp.DecodeBytes(m.Data, &text)
			m1.Payload = &text
		} else if m.App == APP_CONTRACT_INVOKE_REQUEST {
			var invokeReq ContractInvokeRequestPayload
			rlp.DecodeBytes(m.Data, &invokeReq)
			m1.Payload = &invokeReq
		} else if m.App == APP_CONTRACT_INVOKE {
			var invoke ContractInvokePayload
			rlp.DecodeBytes(m.Data, &invoke)
			m1.Payload = &invoke
		} else if m.App == APP_CONTRACT_DEPLOY {
			var deploy ContractDeployPayload
			rlp.DecodeBytes(m.Data, &deploy)
			m1.Payload = &deploy
		} else if m.App == APP_CONFIG {
			var conf ConfigPayload
			rlp.DecodeBytes(m.Data, &conf)
			m1.Payload = &conf
		} else if m.App == APP_CONTRACT_TPL {
			var tplPayload ContractTplPayload
			rlp.DecodeBytes(m.Data, &tplPayload)
			m1.Payload = &tplPayload
		} else if m.App == APP_SIGNATURE {
			var sigPayload SignaturePayload
			rlp.DecodeBytes(m.Data, &sigPayload)
			m1.Payload = &sigPayload
		} else if m.App == APP_VOTE {
			var vote VotePayload
			rlp.DecodeBytes(m.Data, &vote)
			m1.Payload = &vote
		} else if m.App == OP_MEDIATOR_CREATE {
			var mediatorCreateOp MediatorCreateOperation
			rlp.DecodeBytes(m.Data, &mediatorCreateOp)
			m1.Payload = &mediatorCreateOp
		} else {
			fmt.Println("Unknown message app type:", m.App)
		}
		tx.TxMessages = append(tx.TxMessages, m1)

	}
}
