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

package web3ext

const Mediator_JS = `
web3._extend({
	property: 'mediator',
	methods: [
		new web3._extend.Method({
			name: 'schedule',
			call: 'mediator_schedule',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'voteResult',
			call: 'mediator_voteResult',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'getInitDKS',
			call: 'mediator_getInitDKS',
			params: 0,
		}),
		new web3._extend.Method({
			name: 'register',
			call: 'mediator_register',
			params: 1,
		}),
	],
	properties: [
		new web3._extend.Property({
			name: 'list',
			getter: 'mediator_list'
		}),
	]
});
`
