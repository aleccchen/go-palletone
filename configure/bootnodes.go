// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package configure

// MainnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on
// the main PalletOne network.
var MainnetBootnodes = []string{}

// TestnetBootnodes are the pnode URLs of the P2P bootstrap nodes running on the
// Ropsten test network.
var TestnetBootnodes = []string{
	"pnode://a0b8243122e960c49503a133729a66371dacd485868fbda37b2eab9a09cde4205fb6662f68a2f599b41011a730c7e0e638bf34c93af4478141e62744426ea27f@124.251.111.62:30303",
	"pnode://a7c68390def05508aa9833e2f8da227196446fd509236a77d7deba42f6b64127f5e94c870e7e37e1330f010b230a3dd646d574c31b4c932bc4cd6cf3c42582fd@124.251.111.62:30305",
	"pnode://7ffbde2e904e0d7f0d9caf12d804be7c7630bb4ddddd55dd7b2648c6ad0ca19e0c3cc9f000724e9500ae8b18a4f50ecf710ba442bc3cc6da8197e99f59849934@60.205.177.166:30306",
}
