// Copyright 2018 The Fractal Team Authors
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

package sdk

// DposInfo dpos info
func (api *API) DposInfo() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_info")
	return info, err
}

// DposIrreversible dpos irreversible info
func (api *API) DposIrreversible() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_irreversible")
	return info, err
}

// DposCandidate candidate info by name
func (api *API) DposCandidate(name string) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_candidate", name)
	return info, err
}

// DposCandidates candidate info by name
func (api *API) DposCandidates(detail bool) ([]map[string]interface{}, error) {
	info := []map[string]interface{}{}
	err := api.client.Call(&info, "dpos_candidates", detail)
	return info, err
}

// DposVotersByCandidate get voters info of candidate
func (api *API) DposVotersByCandidate(candidate string, detail bool) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_votersByCandidate", candidate, detail)
	return info, err
}

// DposVotersByCandidateByNumber get voters info of candidate
func (api *API) DposVotersByCandidateByNumber(number uint64, candidate string, detail bool) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_votersByCandidateByNumber", number, candidate, detail)
	return info, err
}

// DposVotersByVoter get voters info of voter
func (api *API) DposVotersByVoter(voter string, detail bool) (interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_votersByVoter", voter, detail)
	return info, err
}

// DposVotersByVoterByNumber get voters info of voter
func (api *API) DposVotersByVoterByNumber(number uint64, voter string, detail bool) (interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_votersByVoterByNumber", number, voter, detail)
	return info, err
}

// DposAvailableStake state info
func (api *API) DposAvailableStake(name string) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_availableStake", name)
	return info, err
}

// DposAvailableStakeByNumber state info
func (api *API) DposAvailableStakeByNumber(number uint64, name string) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_availableStakeByNumber", number, name)
	return info, err
}

// DposValidCandidates dpos candidate info
func (api *API) DposValidCandidates() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_validCandidates")
	return info, err
}

// DposValidCandidatesByNumber dpos candidate info
func (api *API) DposValidCandidatesByNumber(number uint64) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_validCandidatesByNumber", number)
	return info, err
}

// DposNextValidCandidates dpos candidate info
func (api *API) DposNextValidCandidates() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_nextValidCandidates")
	return info, err
}

// DposNextValidCandidatesByNumber dpos candidate info
func (api *API) DposNextValidCandidatesByNumber(number uint64) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_nextValidCandidatesByNumber", number)
	return info, err
}

// DposSnapShotTime dpos snapshot time info
func (api *API) DposSnapShotTime() (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_snapShotTime")
	return info, err
}

// DposSnapShotTimeByNumber dpos snapshot time info
func (api *API) DposSnapShotTimeByNumber(number uint64) (map[string]interface{}, error) {
	info := map[string]interface{}{}
	err := api.client.Call(&info, "dpos_snapShotTimeByNumber", number)
	return info, err
}
