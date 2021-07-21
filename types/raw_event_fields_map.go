package types

type EventFieldRedirection string

var RawEventFields = map[string]interface{}{
	"phase": Phase {},
	"accountid": AccountID{},
	"authorityid": AuthorityID{},
	"h160": H160{},
	"u128": U128{},
	"hash": Hash{},
	"balancestatus": BalanceStatus(1),
	"individualexposure": IndividualExposure{},
	"u64": uint64(1),
	"authorityweight": uint64(1),
	"ucompact": UCompact{},
	"exposure": Exposure{},
	"bytes16": Bytes16{},
	"bytes": EventFieldRedirection("vec<u8>"),
	"u32": U32(1),
	"electioncompute": ElectionCompute(1),
	"dispatchinfo": DispatchInfo{},
	"blocknumber": BlockNumber(1),
	"dispatchresult": DispatchResult{},
	"dispatcherror": DispatchError{},
	"proxytype": ProxyType(1),
	"votethreshold": VoteThreshold(1),
	"bool": Bool(true),
	"taskaddress<blocknumber>": TaskAddress{},
	"u8": byte(1),
	"assetid": EventFieldRedirection("u32"),
	"opaquetimeslot": EventFieldRedirection("vec<u8>"),
	"referendumindex": EventFieldRedirection("u32"),
	"proposalindex": EventFieldRedirection("u32"),
	"callhash": EventFieldRedirection("hash"),
	"eraindex": EventFieldRedirection("u32"),
	"bountyindex": EventFieldRedirection("u32"),
	"classid": EventFieldRedirection("u32"),
	"callindex": CallIndex{},
	"kind": Bytes16{},
	"status": EventFieldRedirection("u8"),
	"activeindex": EventFieldRedirection("u32"),
	"boundedvec<u8, t::valuelimit>": EventFieldRedirection("vec<u8>"),
	"boundedvec<u8, t::keylimit>": EventFieldRedirection("vec<u8>"),
	"membercount": EventFieldRedirection("u32"),
	"sessionindex": EventFieldRedirection("u32"),
	"identificationtuple": struct {
		ValidatorID        AccountID
		FullIdentification Exposure
	} {},
	"registrarindex": EventFieldRedirection("u32"),
	"propindex": EventFieldRedirection("u32"),
	// Not going to be used
	"sp_std::marker::phantomdata<(accountId, event)>": "",
}