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
 * @author PalletOne core developer Albert·Gou <dev@pallet.one>
 * @date 2018
 */

package core

import (
	"fmt"

	"github.com/btcsuite/btcutil/base58"
	"github.com/dedis/kyber"
	"github.com/palletone/go-palletone/common/log"
	"strconv"
)

// Genesis specifies the header fields, state of a genesis block. It also defines hard
// fork switch-over blocks through the chain configuration.
type SystemConfig struct {
	//年利率
	DepositRate string `json:"depositRate"`
	//基金会地址，该地址具有一些特殊权限，比如发起参数修改的投票，发起罚没保证金等
	FoundationAddress string `json:"foundationAddress"`
	//保证金的数量
	DepositAmountForMediator  string `json:"depositAmountForMediator"`
	DepositAmountForJury      string `json:"depositAmountForJury"`
	DepositAmountForDeveloper string `json:"depositAmountForDeveloper"`
	//保证金周期
	DepositPeriod string `json:"depositPeriod"`
}

type Genesis struct {
	Version string `json:"version"`
	Alias   string `json:"alias"`
	//TokenAmount  uint64       `json:"tokenAmount"`
	TokenAmount  string       `json:"tokenAmount"`
	TokenDecimal uint32       `json:"tokenDecimal"`
	DecimalUnit  string       `json:"decimal_unit"`
	ChainID      uint64       `json:"chainId"`
	TokenHolder  string       `json:"tokenHolder"`
	Text         string       `json:"text"`
	SystemConfig SystemConfig `json:"systemConfig"`

	InitialParameters         ChainParameters          `json:"initialParameters"`
	ImmutableParameters       ImmutableChainParameters `json:"immutableChainParameters"`
	InitialTimestamp          int64                    `json:"initialTimestamp"`
	InitialActiveMediators    uint16                   `json:"initialActiveMediators"`
	InitialMediatorCandidates []*InitialMediator       `json:"initialMediatorCandidates"`
}

func (g *Genesis) GetTokenAmount() uint64 {
	amount, err := strconv.ParseInt(g.TokenAmount, 10, 64)
	if err != nil {
		log.Error("genesis", "get token amount err:", err)
		return uint64(0)
	}
	return uint64(amount)
}

type InitialMediator struct {
	AddStr      string `json:"account"`
	InitPartPub string `json:"initPubKey"`
	Node        string `json:"node"`
}

// author Albert·Gou
func ScalarToStr(sec kyber.Scalar) string {
	secB, err := sec.MarshalBinary()
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	return base58.Encode(secB)
}

// author Albert·Gou
func PointToStr(pub kyber.Point) string {
	pubB, err := pub.MarshalBinary()
	if err != nil {
		log.Error(fmt.Sprintln(err))
	}

	return base58.Encode(pubB)
}
