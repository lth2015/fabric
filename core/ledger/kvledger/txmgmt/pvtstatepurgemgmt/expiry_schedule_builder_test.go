/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package pvtstatepurgemgmt

import (
	"fmt"
	"testing"

	"github.com/davecgh/go-spew/spew"

	"github.com/hyperledger/fabric/common/ledger/testutil"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/privacyenabledstate"
	"github.com/hyperledger/fabric/core/ledger/kvledger/txmgmt/version"
	"github.com/hyperledger/fabric/core/ledger/pvtdatapolicy"
	"github.com/hyperledger/fabric/core/ledger/util"
	"github.com/spf13/viper"
)

func TestBuildExpirySchedule(t *testing.T) {
	ledgerid := "testledger-BuildExpirySchedule"
	viper.Set(fmt.Sprintf("ledger.pvtdata.btlpolicy.%s.ns1.coll1", ledgerid), 1)
	viper.Set(fmt.Sprintf("ledger.pvtdata.btlpolicy.%s.ns1.coll2", ledgerid), 2)
	viper.Set(fmt.Sprintf("ledger.pvtdata.btlpolicy.%s.ns2.coll3", ledgerid), 3)

	btlPolicy, _ := pvtdatapolicy.GetBTLPolicy(ledgerid)

	updates := privacyenabledstate.NewUpdateBatch()
	updates.PubUpdates.Put("ns1", "pubkey1", []byte("pubvalue1"), version.NewHeight(1, 1))
	putPvtUpdates(t, updates, "ns1", "coll1", "pvtkey1", []byte("pvtvalue1"), version.NewHeight(1, 1))
	putPvtUpdates(t, updates, "ns1", "coll2", "pvtkey2", []byte("pvtvalue2"), version.NewHeight(2, 1))
	putPvtUpdates(t, updates, "ns2", "coll3", "pvtkey3", []byte("pvtvalue3"), version.NewHeight(3, 1))
	putPvtUpdates(t, updates, "ns3", "coll4", "pvtkey4", []byte("pvtvalue4"), version.NewHeight(4, 1))

	listExpinfo := buildExpirySchedule(btlPolicy, updates.PvtUpdates, updates.HashUpdates)
	t.Logf("listExpinfo=%s", spew.Sdump(listExpinfo))

	pvtdataKeys1 := newPvtdataKeys()
	pvtdataKeys1.add("ns1", "coll1", "pvtkey1", util.ComputeStringHash("pvtkey1"))

	pvtdataKeys2 := newPvtdataKeys()
	pvtdataKeys2.add("ns1", "coll2", "pvtkey2", util.ComputeStringHash("pvtkey2"))

	pvtdataKeys3 := newPvtdataKeys()
	pvtdataKeys3.add("ns2", "coll3", "pvtkey3", util.ComputeStringHash("pvtkey3"))

	expectedListExpInfo := []*expiryInfo{
		{expiryInfoKey: &expiryInfoKey{expiryBlk: 3, committingBlk: 1}, pvtdataKeys: pvtdataKeys1},
		{expiryInfoKey: &expiryInfoKey{expiryBlk: 5, committingBlk: 2}, pvtdataKeys: pvtdataKeys2},
		{expiryInfoKey: &expiryInfoKey{expiryBlk: 7, committingBlk: 3}, pvtdataKeys: pvtdataKeys3},
	}

	testutil.AssertEquals(t, len(listExpinfo), 3)
	testutil.AssertContainsAll(t, listExpinfo, expectedListExpInfo)
}

func putPvtUpdates(t *testing.T, updates *privacyenabledstate.UpdateBatch, ns, coll, key string, value []byte, ver *version.Height) {
	updates.PvtUpdates.Put(ns, coll, key, value, ver)
	updates.HashUpdates.Put(ns, coll, util.ComputeStringHash(key), util.ComputeHash(value), ver)
}

func deletePvtUpdates(t *testing.T, updates *privacyenabledstate.UpdateBatch, ns, coll, key string, ver *version.Height) {
	updates.PvtUpdates.Delete(ns, coll, key, ver)
	updates.HashUpdates.Delete(ns, coll, util.ComputeStringHash(key), ver)
}
