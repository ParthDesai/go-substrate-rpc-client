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
	"proxytype": ProxyType(1),
	"votethreshold": VoteThreshold(1),
	"bool": Bool(true),
	"taskaddress<blocknumber>": TaskAddress{},
	"u8": byte(1),
	"assetid": EventFieldRedirection("u32"),
	"opaquetimeslot": EventFieldRedirection("vec<u8>"),
}