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

package log

var DefaultConfig = Config{
	OutputPaths:      []string{"stdout", "./log/all.log"},
	ErrorOutputPaths: []string{"stderr", "./log/error.log"},
	OpenModule:       []string{"all"},
	LoggerLvl:        "DEBUG",
	Encoding:         "console",
	Development:      true,
}

type Config struct {
	OutputPaths      []string `json:"outputPaths" yaml:"outputPaths"`           // output file path
	ErrorOutputPaths []string `json:"errorOutputPaths" yaml:"errorOutputPaths"` // error file path
	OpenModule       []string // open module
	LoggerLvl        string   `json:"level" yaml:"level"`       // log level
	Encoding         string   `json:"encoding" yaml:"encoding"` // encoding
	Development      bool     `json:"development" yaml:"development"`
}
